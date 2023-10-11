/*
Copyright 2022 The Kubernetes Authors.

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

package scaleway

import (
	"fmt"

	"k8s.io/klog/v2"
	"sigs.k8s.io/etcdadm/etcd-manager/pkg/privateapi/discovery"
)

// Volumes also allows the discovery of peer nodes
var _ discovery.Interface = &Volumes{}

// Poll returns all etcd cluster peers.
func (a *Volumes) Poll() (map[string]discovery.Node, error) {
	peers := make(map[string]discovery.Node)
	klog.V(2).Infof("Discovering peers with volumes matching labels: %v", a.matchTags)

	etcdVolumes, err := getMatchingVolumes(a.instanceAPI, a.zone, a.matchTags)
	if err != nil {
		return nil, fmt.Errorf("failed to get matching volumes: %w", err)
	}

	for _, volume := range etcdVolumes {
		if volume.Server == nil {
			// Volume doesn't have a server attached yet
			continue
		}
		serverID := volume.Server.ID

		ip, err := a.getServerIP(serverID)
		if err != nil {
			return nil, fmt.Errorf("getting IP for server %s: %w", serverID, err)
		}

		klog.V(2).Infof("Discovered volume %s(%s) of type %s attached to server %s", volume.Name, volume.ID, volume.VolumeType, serverID)
		// We use the etcd node ID as the persistent identifier, because the data determines who we are
		node := discovery.Node{
			ID:        "vol-" + volume.ID,
			Endpoints: []discovery.NodeEndpoint{{IP: ip}},
		}
		peers[node.ID] = node
	}

	return peers, nil
}
