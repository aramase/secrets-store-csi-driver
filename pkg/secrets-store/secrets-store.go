/*
Copyright 2018 The Kubernetes Authors.

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
	"context"
	"sync"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"k8s.io/utils/mount"

	"sigs.k8s.io/controller-runtime/pkg/client"

	csicommon "sigs.k8s.io/secrets-store-csi-driver/pkg/csi-common"
	"sigs.k8s.io/secrets-store-csi-driver/pkg/version"

	"k8s.io/klog/v2"
)

// SecretsStore implements the IdentityServer, ControllerServer and
// NodeServer CSI interfaces.
type SecretsStore struct {
	providerVolumePath string
	mounter            mount.Interface
	reporter           StatsReporter
	nodeID             string
	client             client.Client
	providerClients    *PluginClientBuilder

	// mutex and volumes are used by controllerserver
	mu   sync.Mutex
	vols map[string]csi.Volume
}

// NewSecretsStore returns a new secrets store driver
func NewSecretsStore(providerVolumePath, nodeID string, mounter mount.Interface, providerClients *PluginClientBuilder, client client.Client, statsReporter StatsReporter) (*SecretsStore, error) {
	return &SecretsStore{
		providerVolumePath: providerVolumePath,
		mounter:            mounter,
		reporter:           statsReporter,
		nodeID:             nodeID,
		client:             client,
		providerClients:    providerClients,
		mu:                 sync.Mutex{},
		vols:               make(map[string]csi.Volume),
	}, nil
}

// Run starts the CSI plugin
func (s *SecretsStore) Run(ctx context.Context, driverName, nodeID, endpoint, providerVolumePath string, providerClients *PluginClientBuilder, client client.Client) {
	klog.Infof("Driver: %v ", driverName)
	klog.Infof("Version: %s, BuildTime: %s", version.BuildVersion, version.BuildTime)
	klog.Infof("Provider Volume Path: %s", providerVolumePath)
	klog.Infof("GRPC supported providers will be dynamically created")

	server := csicommon.NewNonBlockingGRPCServer()
	// s implements ControllerServer, NodeServer and IdentityServer
	server.Start(ctx, endpoint, s, s, s)
	server.Wait()
}
