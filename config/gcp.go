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
	"context"
	"fmt"

	"github.com/golang/glog"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"k8s.io/client/kubernetes/config/api"
)

const (
	gcpRFC3339Format = "2006-01-02 15:04:05"
)

// GoogleCredentialLoader defines the interface for getting GCP token
type GoogleCredentialLoader interface {
	GetGoogleCredentials() (*oauth2.Token, error)
}

func loadGCPToken(authInfo *api.AuthInfo, ld *KubeConfigLoader) (string, bool) {
	if authInfo.AuthProvider == nil || authInfo.AuthProvider.Name != "gcp" {
		return "", false
	}

	// Refresh GCP token if necessary
	if authInfo.AuthProvider.Config == nil {
		if err := refreshGCPToken(authInfo, ld); err != nil {
			glog.Errorf("failed to refresh GCP token: %v", err)
			return "", false
		}
	}
	if _, ok := authInfo.AuthProvider.Config["expiry"]; !ok {
		if err := refreshGCPToken(authInfo, ld); err != nil {
			glog.Errorf("failed to refresh GCP token: %v", err)
			return "", false
		}
	}
	expired, err := isExpired(authInfo.AuthProvider.Config["expiry"])
	if err != nil {
		glog.Errorf("failed to determine if GCP token is expired: %v", err)
		return "", false
	}

	if expired {
		if err := refreshGCPToken(authInfo, ld); err != nil {
			glog.Errorf("failed to refresh GCP token: %v", err)
			return "", false
		}
	}

	// Use GCP access token
	return "Bearer " + authInfo.AuthProvider.Config["access-token"], true
}

func refreshGCPToken(authInfo *api.AuthInfo, l *KubeConfigLoader) error {
	if authInfo.AuthProvider.Config == nil {
		authInfo.AuthProvider.Config = map[string]string{}
	}

	// Get *oauth2.Token through Google APIs
	if l.gcLoader == nil {
		l.gcLoader = DefaultGoogleCredentialLoader{}
	}
	credentials, err := l.gcLoader.GetGoogleCredentials()
	if err != nil {
		return err
	}

	// Store credentials to Config
	authInfo.AuthProvider.Config["access-token"] = credentials.AccessToken
	authInfo.AuthProvider.Config["expiry"] = credentials.Expiry.Format(gcpRFC3339Format)

	setUserWithName(l.rawConfig.AuthInfos, l.currentContext.AuthInfo, &l.user)
	// Persist kube config file
	if l.skipConfigPersist {
		if err := l.persistConfig(); err != nil {
			return err
		}
	}
	return nil
}

// DefaultGoogleCredentialLoader provides the default method for getting GCP token
type DefaultGoogleCredentialLoader struct{}

// GetGoogleCredentials fetches GCP using default locations
func (l DefaultGoogleCredentialLoader) GetGoogleCredentials() (*oauth2.Token, error) {
	credentials, err := google.FindDefaultCredentials(context.Background(), "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return nil, fmt.Errorf("failed to get Google credentials: %v", err)
	}
	return credentials.TokenSource.Token()
}
