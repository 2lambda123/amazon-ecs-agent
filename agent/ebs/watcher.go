// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//	http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package ebs

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	ecsapi "github.com/aws/amazon-ecs-agent/agent/api"
	ecsengine "github.com/aws/amazon-ecs-agent/agent/engine"
	"github.com/aws/amazon-ecs-agent/agent/engine/dockerstate"
	attachmentinfo "github.com/aws/amazon-ecs-agent/ecs-agent/api/attachmentinfo"
	apiebs "github.com/aws/amazon-ecs-agent/ecs-agent/api/resource"
	apiattachmentstatus "github.com/aws/amazon-ecs-agent/ecs-agent/api/status"
	apitaskstatus "github.com/aws/amazon-ecs-agent/ecs-agent/api/task/status"
	csi "github.com/aws/amazon-ecs-agent/ecs-agent/csiclient"
	apieni "github.com/aws/amazon-ecs-agent/ecs-agent/netlib/model/networkinterface"
	log "github.com/cihub/seelog"

	v1 "k8s.io/api/core/v1"
)

const (
	// We are capping off this duration to 30s assuming worst-case behavior
	nodeStageTimeout = 2 * time.Second

	// ebsStatusSentMsg is the error message to use when trying to send an ebs status that's
	// already been sent
	ebsStatusSentMsg = "ebs status already sent"
)

type EBSWatcher struct {
	ctx        context.Context
	cancel     context.CancelFunc
	agentState dockerstate.TaskEngineState
	// TODO: The dataClient will be used to save to agent's data client as well as start the ACK timer. This will be added once the data client functionality have been added
	// dataClient     data.Client
	discoveryClient apiebs.EBSDiscovery
	csiClient       csi.CSIClient
	scanTicker      *time.Ticker
	// TODO: The dockerTaskEngine.stateChangeEvent will be used to send over the state change event for EBS attachments once it's been found and mounted/resize/format.
	taskEngine ecsengine.TaskEngine
}

// NewWatcher is used to return a new instance of the EBSWatcher struct
func NewWatcher(ctx context.Context,
	state dockerstate.TaskEngineState,
	taskEngine ecsengine.TaskEngine) *EBSWatcher {
	derivedContext, cancel := context.WithCancel(ctx)
	discoveryClient := apiebs.NewDiscoveryClient(derivedContext)
	// TODO pull this socket out into config
	csiClient := csi.NewCSIClient("/var/run/ecs/ebs-csi-driver/csi-driver.sock")
	return &EBSWatcher{
		ctx:             derivedContext,
		cancel:          cancel,
		agentState:      state,
		discoveryClient: discoveryClient,
		csiClient:       &csiClient,
		taskEngine:      taskEngine,
	}
}

// Start is used to kick off the periodic scanning process of the EBS volume attachments for the EBS watcher.
// It will be start and continue to run whenever there's a pending EBS volume attachment that hasn't been found.
// If there aren't any, the scan ticker will not start up/scan for volumes.
func (w *EBSWatcher) Start() {
	log.Info("Starting EBS watcher.")
	w.scanTicker = time.NewTicker(apiebs.ScanPeriod)
	for {
		select {
		case <-w.scanTicker.C:
			pendingEBS := w.agentState.GetAllPendingEBSAttachmentsWithKey()
			if len(pendingEBS) > 0 {
				foundVolumes := apiebs.ScanEBSVolumes(pendingEBS, w.discoveryClient)
				w.overrideDeviceName(foundVolumes)
				if err := w.StageAll(foundVolumes); err != nil {
					log.Errorf("stage error: %s", err)
					continue
				}
				w.NotifyAttached(foundVolumes)
			}
		case <-w.ctx.Done():
			w.scanTicker.Stop()
			log.Info("EBS Watcher Stopped due to agent stop")
			return
		}
	}
}

// Stop will stop the EBS watcher
func (w *EBSWatcher) Stop() {
	log.Info("Stopping EBS watcher.")
	w.cancel()
}

func (w *EBSWatcher) HandleResourceAttachment(ebs *apiebs.ResourceAttachment) {
	err := w.HandleEBSResourceAttachment(ebs)
	if err != nil {
		log.Errorf("Unable to handle resource attachment payload %s", err)
	}
}

// HandleResourceAttachment processes the resource attachment message. It will:
// 1. Check whether we already have this attachment in state and if so it's a noop.
// 2. Start the EBS CSI driver if it's not already running
// 3. Otherwise add the attachment to state, start its ack timer, and save to the agent state.
func (w *EBSWatcher) HandleEBSResourceAttachment(ebs *apiebs.ResourceAttachment) error {
	attachmentType := ebs.GetAttachmentType()
	if attachmentType != apiebs.EBSTaskAttach {
		log.Warnf("Resource type not Elastic Block Storage. Skip handling resource attachment with type: %v.", attachmentType)
		return nil
	}

	volumeId := ebs.GetAttachmentProperties(apiebs.VolumeIdKey)
	ebsAttachment, ok := w.agentState.GetEBSByVolumeId(volumeId)
	if ok {
		log.Infof("EBS Volume attachment already exists. Skip handling EBS attachment %v.", ebs.EBSToString())
		return ebsAttachment.StartTimer(func() {
			w.handleEBSAckTimeout(volumeId)
		})
	}

	// start EBS CSI Driver Managed Daemon
	if w.taskEngine.GetEbsDaemonTask() != nil {
		log.Infof("engine ebs CSI driver is running with taskID: %v", w.taskEngine.GetEbsDaemonTask().GetID())
	} else {
		if ebsCsiDaemonManager, ok := w.taskEngine.GetDaemonManagers()["ebs-csi-driver"]; ok {
			if csiTask, err := ebsCsiDaemonManager.CreateDaemonTask(); err != nil {
				// fail attachment and return
				log.Errorf("Unable to start ebsCsiDaemon for task %s in the engine: error: %s", csiTask.GetID(), err)
				csiTask.SetKnownStatus(apitaskstatus.TaskStopped)
				csiTask.SetDesiredStatus(apitaskstatus.TaskStopped)
				w.taskEngine.EmitTaskEvent(csiTask, err.Error())
				return err
			} else {
				w.taskEngine.SetEbsDaemonTask(csiTask)
				w.taskEngine.AddTask(csiTask)
				log.Infof("task_engine: Added EBS CSI task to engine")
			}
		} else {
			log.Errorf("CSI Driver is not Initialized")
		}
	}

	if err := w.addEBSAttachmentToState(ebs); err != nil {
		return fmt.Errorf("%w; attach %v message handler: unable to add ebs attachment to engine state: %v",
			err, attachmentType, ebs.EBSToString())
	}
	log.Debug("EBS Attachment has been added to state successfully")
	return nil
}

func (w *EBSWatcher) overrideDeviceName(foundVolumes map[string]string) {
	for volumeId, deviceName := range foundVolumes {
		ebs, ok := w.agentState.GetEBSByVolumeId(volumeId)
		if !ok {
			log.Warnf("Unable to find EBS volume with volume ID: %s", volumeId)
			continue
		}
		ebs.SetDeviceName(deviceName)
	}
}

// assumes CSI Driver Managed Daemon is running else call will timeout
func (w *EBSWatcher) StageAll(foundVolumes map[string]string) error {
	for volID, deviceName := range foundVolumes {
		// get volume details from attachment
		ebsAttachment, ok := w.agentState.GetEBSByVolumeId(volID)
		if !ok {
			continue
		}
		if ebsAttachment.IsSent() {
			log.Warnf("State change event has already been emitted for EBS volume: %v.", ebsAttachment.EBSToString())
			continue
			// return nil
		}
		if ebsAttachment.HasExpired() {
			log.Warnf("From stage all, EBS status expired, no longer tracking EBS volume: %v.", ebsAttachment.EBSToString())
			continue
			//return nil
		}
		if ebsAttachment.IsAttached() {
			continue
			//return nil
		}

		// hostPath := ebsAttachment.GetAttachmentProperties(apiebs.SourceVolumeHostPathKey)

		preHostPath := ebsAttachment.GetAttachmentProperties(apiebs.SourceVolumeHostPathKey)
		hostPath := filepath.Join("/mnt/ecs/ebs", preHostPath[strings.LastIndex(preHostPath, "/")+1:])
		filesystemType := ebsAttachment.GetAttachmentProperties(apiebs.FileSystemTypeName)

		// // CSI NodeStage stub equired fields
		stubSecrets := make(map[string]string)
		stubVolumeContext := make(map[string]string)
		stubMountOptions := []string{}
		stubFsGroup, _ := strconv.ParseInt("123456", 10, 8)
		publishContext := map[string]string{"devicePath": deviceName}
		// call CSI NodeStage
		// todo handle async timout failure
		// errChan := make(chan error)
		timeoutCtx, cancelFunc := context.WithTimeout(w.ctx, nodeStageTimeout)
		defer func() {
			log.Infof("cancelFunc called in node stage timeout")
			cancelFunc()
		}()

		// go func(timeoutCtx context.Context) {
		// 	err := w.stageAllHelper(timeoutCtx, volID, hostPath, deviceName, filesystemType)
		// 	errChan <- err
		// }(timeoutCtx)

		err := w.csiClient.NodeStageVolume(timeoutCtx,
			volID,
			publishContext,
			hostPath,
			filesystemType,
			v1.ReadWriteMany,
			stubSecrets,
			stubVolumeContext,
			stubMountOptions,
			&stubFsGroup)

		// err := w.stageAllHelper(timeoutCtx, volID, hostPath, deviceName, filesystemType)

		// select {
		// case <-timeoutCtx.Done():
		// 	if timeoutCtx.Err() != nil {
		// 		log.Warnf("Context canceled %vs", timeoutCtx.Err())
		// 		return timeoutCtx.Err()
		// 	}
		// case err := <-errChan:
		// 	if err != nil {
		// 		log.Errorf("Failed to initialize EBS volume: error: %s", err)
		// 		return err
		// 	}
		// }
		if err != nil {
			log.Errorf("Failed to initialize EBS volume: error: %s", err)
			return err
		}
		// set attached status
		log.Infof("We've set attached status for %v", ebsAttachment.EBSToString())
		ebsAttachment.SetAttachedStatus()
	}
	return nil
}

// func (w *EBSWatcher) stageAllHelper(ctx context.Context, volID, hostPath, deviceName, filesystemType string) error {
// 	stubSecrets := make(map[string]string)
// 	stubVolumeContext := make(map[string]string)
// 	stubMountOptions := []string{}
// 	stubFsGroup, _ := strconv.ParseInt("123456", 10, 8)
// 	publishContext := map[string]string{"devicePath": deviceName}
// 	err := w.csiClient.NodeStageVolume(ctx,
// 		volID,
// 		publishContext,
// 		hostPath,
// 		filesystemType,
// 		v1.ReadWriteMany,
// 		stubSecrets,
// 		stubVolumeContext,
// 		stubMountOptions,
// 		&stubFsGroup)

// 	return err
// }

// NotifyAttached will go through the list of found EBS volumes from the scanning process and mark them as found.
func (w *EBSWatcher) NotifyAttached(foundVolumes map[string]string) {
	for volID := range foundVolumes {
		w.notifyAttachedEBS(volID)
	}
}

// notifyAttachedEBS will mark it as found within the agent state
func (w *EBSWatcher) notifyAttachedEBS(volumeId string) {
	// TODO: Add the EBS volume to data client
	ebs, ok := w.agentState.GetEBSByVolumeId(volumeId)
	if !ok {
		log.Warnf("Unable to find EBS volume with volume ID: %v within agent state.", volumeId)
		return
	}

	if ebs.HasExpired() {
		log.Warnf("EBS status expired, no longer tracking EBS volume: %v.", ebs.EBSToString())
		return
	}

	if ebs.IsSent() {
		log.Warnf("State change event has already been emitted for EBS volume: %v.", ebs.EBSToString())
		return
	}
	// We found an EBS volume which has the expiration time set in future and
	// needs to be acknowledged as having been 'attached' to the Instance
	if err := w.sendEBSStateChange(ebs); err != nil {
		// skip logging status sent error as it's redundant and doesn't really indicate a problem
		if !strings.Contains(err.Error(), ebsStatusSentMsg) {
			log.Warnf("Unable to send state EBS change, %s", err)
			return
		}
	}
	ebs.SetSentStatus()
	log.Infof("We've set sent status for %v", ebs.EBSToString())
	ebs.StopAckTimer()
	log.Infof("Successfully found attached EBS volume: %v", ebs.EBSToString())
}

// removeEBSAttachment removes a EBS volume with a specific volume ID
func (w *EBSWatcher) removeEBSAttachment(volumeID string) {
	// TODO: Remove the EBS volume from the data client.
	w.agentState.RemoveEBSAttachment(volumeID)
}

// addEBSAttachmentToState adds an EBS attachment to state, and start its ack timer
func (w *EBSWatcher) addEBSAttachmentToState(ebs *apiebs.ResourceAttachment) error {
	volumeId := ebs.AttachmentProperties[apiebs.VolumeIdKey]
	err := ebs.StartTimer(func() {
		w.handleEBSAckTimeout(volumeId)
	})
	if err != nil {
		log.Warnf("Unable to handl ack timeout")
		return err
	}

	w.agentState.AddEBSAttachment(ebs)
	return nil
}

// handleEBSAckTimeout removes EBS attachment from agent state after the EBS ack timeout
func (w *EBSWatcher) handleEBSAckTimeout(volumeId string) {
	ebsAttachment, ok := w.agentState.GetEBSByVolumeId(volumeId)
	if !ok {
		log.Warnf("Ignoring unmanaged EBS attachment volume ID=%v", volumeId)
		return
	}
	if !ebsAttachment.IsSent() {
		log.Warnf("Timed out waiting for EBS ack; removing EBS attachment record %v", ebsAttachment.EBSToString())
		w.removeEBSAttachment(volumeId)
	}
}

func (w *EBSWatcher) sendEBSStateChange(ebsvol *apiebs.ResourceAttachment) error {
	if ebsvol == nil {
		return errors.New("ebs watcher send EBS state change: nil volume")
	}
	go w.emitEBSAttachedEvent(ebsvol)
	return nil
}

func (w *EBSWatcher) emitEBSAttachedEvent(ebsvol *apiebs.ResourceAttachment) {
	attachmentInfo := attachmentinfo.AttachmentInfo{
		AttachmentARN:        ebsvol.GetAttachmentARN(),
		Status:               apiattachmentstatus.AttachmentAttached,
		ExpiresAt:            ebsvol.GetExpiresAt(),
		ClusterARN:           ebsvol.GetClusterARN(),
		ContainerInstanceARN: ebsvol.GetContainerInstanceARN(),
	}
	attachmentChange := ecsapi.AttachmentStateChange{
		Attachment: &apieni.ENIAttachment{AttachmentInfo: attachmentInfo},
	}

	log.Infof("Emitting EBS volume attached event for: %v", ebsvol)
	w.taskEngine.StateChangeEvents() <- attachmentChange
}
