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

func (l *KubeConfigLoader) loadAuthProviderToken() bool {
	if l.user.AuthProvider == nil {
		return false
	}
	ok := false
	if l.user.AuthProvider.Name == "gcp" {
		l.restConfig.token, ok = loadGCPToken(&l.user, l)
	}
	if l.user.AuthProvider.Name == "azure" {
		l.restConfig.token, ok = loadAzureToken(&l.user, l)
	}
	return ok
}
