/*
Copyright 2019 The Kubernetes Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package secretsstore

import (
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"golang.org/x/net/context"
)

var counter uint64

func (s *SecretsStore) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	if err := s.validateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		return nil, err
	}
	if len(req.GetName()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "volume name is empty")
	}
	if req.GetVolumeCapabilities() == nil {
		return nil, status.Error(codes.InvalidArgument, "volume_capabilities is empty")
	}
	capacityBytes := req.GetCapacityRange().GetRequiredBytes()
	volumeContext := req.GetParameters()
	volName := req.GetName()

	if volumeContext == nil {
		volumeContext = make(map[string]string)
	}
	volumeContext["providerName"] = "mock_provider"

	// check if volume with same name exists
	existingVol, exists := s.findVolumeByName(volName)
	// if volume exists and capacity is different then error
	if exists && existingVol.CapacityBytes != capacityBytes {
		return nil, status.Error(codes.AlreadyExists, "volume with same name and diff capacity exists")
	}
	volumeID := existingVol.VolumeId
	if !exists {
		volumeID = fmt.Sprintf("%s-%d", req.GetName(), atomic.AddUint64(&counter, 1))
	}
	newVolume := csi.Volume{
		VolumeId:      volumeID,
		CapacityBytes: capacityBytes,
		VolumeContext: volumeContext,
	}

	s.addVolume(volName, newVolume)
	return &csi.CreateVolumeResponse{Volume: &newVolume}, nil
}

func (s *SecretsStore) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	if err := s.validateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		return nil, err
	}
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "volume id missing in request")
	}
	return &csi.DeleteVolumeResponse{}, nil
}

func (s *SecretsStore) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "volume id missing in request")
	}
	if req.GetVolumeCapabilities() == nil {
		return nil, status.Error(codes.InvalidArgument, "volume_capabilities is empty")
	}
	reqVolID := req.GetVolumeId()
	if _, exists := s.findVolumeByID(reqVolID); exists {
		return &csi.ValidateVolumeCapabilitiesResponse{}, nil
	}
	return nil, status.Error(codes.NotFound, reqVolID)
}

func (s *SecretsStore) findVolumeByName(volName string) (csi.Volume, bool) {
	return s.findVolume("name", volName)
}

func (s *SecretsStore) findVolumeByID(volID string) (csi.Volume, bool) {
	return s.findVolume("id", volID)
}

func (s *SecretsStore) addVolume(name string, vol csi.Volume) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.vols[name] = vol
}

func (s *SecretsStore) findVolume(key, nameOrID string) (csi.Volume, bool) {
	return s.findVolumeInternal(key, nameOrID)
}

func (s *SecretsStore) findVolumeInternal(key, nameOrID string) (csi.Volume, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch key {
	case "name":
		vol, ok := s.vols[nameOrID]
		return vol, ok

	case "id":
		for _, vol := range s.vols {
			if strings.EqualFold(nameOrID, vol.VolumeId) {
				return vol, true
			}
		}
	}
	return csi.Volume{}, false
}

func isMockProvider(provider string) bool {
	return strings.EqualFold(provider, "mock_provider")
}

func isMockTargetPath(targetPath string) bool {
	return strings.EqualFold(targetPath, "/tmp/csi/mount")
}

func (s *SecretsStore) validateControllerServiceRequest(c csi.ControllerServiceCapability_RPC_Type) error {
	if c == csi.ControllerServiceCapability_RPC_UNKNOWN {
		return nil
	}

	for _, cap := range s.getControllerServiceCapabilities() {
		if c == cap.GetRpc().GetType() {
			return nil
		}
	}
	return status.Error(codes.InvalidArgument, c.String())
}

func (s *SecretsStore) getControllerServiceCapabilities() []*csi.ControllerServiceCapability {
	var csc []*csi.ControllerServiceCapability
	return csc
}
