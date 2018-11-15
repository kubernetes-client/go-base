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

package config

import (
	"strconv"
	"time"

	"github.com/golang/glog"
	"k8s.io/client/kubernetes/config/api"
)

func loadAzureToken(authInfo *api.AuthInfo, ld *KubeConfigLoader) (string, bool) {
	if authInfo.AuthProvider == nil || authInfo.AuthProvider.Name != "azure" {
		return "", false
	}

	expiry, found := authInfo.AuthProvider.Config["expires-on"]
	if found {
		ts, err := strconv.ParseInt(expiry, 10, 64)
		if err != nil {
			glog.Errorf("Failed to parse expires time: %v", expiry)
			return "", false
		}
		if ts < time.Now().Unix() {
			glog.Errorf("Token is expired (and refresh isn't implemented)")
			return "", false
		}
	}

	return "Bearer " + authInfo.AuthProvider.Config["access-token"], true
}
