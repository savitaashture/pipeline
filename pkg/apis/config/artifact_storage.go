/*
Copyright 2024 The Tekton Authors

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
	"fmt"
	"os"
	"strconv"

	corev1 "k8s.io/api/core/v1"
)

const (
	// ArtifactStorageConfigName is the name of the ConfigMap for artifact storage configuration.
	ArtifactStorageConfigName = "config-artifact-storage"

	ociRepositoryKey = "oci-repository"
	insecureKey      = "insecure"
)

// ArtifactStorage holds configuration for OCI-based artifact storage.
type ArtifactStorage struct {
	// OCIRepository is the base OCI repository for artifact storage
	// e.g., "ghcr.io/org/project/artifacts" or "registry:5000/artifacts"
	OCIRepository string
	// Insecure allows plain HTTP for the OCI registry
	Insecure bool
}

// NewArtifactStorageFromMap creates an ArtifactStorage from a map of string values.
func NewArtifactStorageFromMap(cfgMap map[string]string) (*ArtifactStorage, error) {
	as := &ArtifactStorage{}

	if v, ok := cfgMap[ociRepositoryKey]; ok {
		as.OCIRepository = v
	}

	if v, ok := cfgMap[insecureKey]; ok && v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return nil, fmt.Errorf("failed parsing artifact storage config %q: %w for key %s", v, err, insecureKey)
		}
		as.Insecure = b
	}

	return as, nil
}

// NewArtifactStorageFromConfigMap creates an ArtifactStorage from a ConfigMap.
func NewArtifactStorageFromConfigMap(config *corev1.ConfigMap) (*ArtifactStorage, error) {
	return NewArtifactStorageFromMap(config.Data)
}

// GetArtifactStorageConfigName returns the name of the ConfigMap for artifact storage.
func GetArtifactStorageConfigName() string {
	if e := os.Getenv("CONFIG_ARTIFACT_STORAGE_NAME"); e != "" {
		return e
	}
	return ArtifactStorageConfigName
}
