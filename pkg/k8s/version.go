// Copyright 2016-2019 Authors of Cilium
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package version keeps track of the Kubernetes version the client is
// connected to.
// Updated source code used from Cilium code repository. github.com/cilium/cilium
// https://github.com/cilium/cilium/blob/master/pkg/k8s/version/version.go

package k8s

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	go_version "github.com/blang/semver"
	"github.com/golang/glog"
	"k8s.io/client-go/kubernetes"
)

type cachedVersion struct {
	mutex   sync.RWMutex
	version go_version.Version
}

var (
	cached = cachedVersion{}

	// Make sure your Kubernetes cluster has a master version of 1.16.0 or higher to use
	// apiextensions.k8s.io/v1, or 1.7.0 or higher for apiextensions.k8s.io/v1beta1.
	isGEThanAPIExtV1      = mustCompile(">=1.16.0")
	isGEThanAPIExtV1Beta1 = mustCompile(">=1.7.0")
)

func mustCompile(constraint string) go_version.Range {
	verCheck, err := go_version.ParseRange(constraint)
	if err != nil {
		panic(fmt.Errorf("cannot compile go-version constraint '%s' %s", constraint, err))
	}

	return verCheck
}

func parseVersion(version string) (go_version.Version, error) {
	ver, err := go_version.ParseTolerant(version)
	if err != nil {
		return ver, err
	}

	if len(ver.Pre) == 0 {
		return ver, nil
	}

	for _, pre := range ver.Pre {
		if strings.Contains(pre.VersionStr, "rc") ||
			strings.Contains(pre.VersionStr, "beta") ||
			strings.Contains(pre.VersionStr, "alpha") {

			return ver, nil
		}
	}

	strSegments := make([]string, 3)
	strSegments[0] = strconv.Itoa(int(ver.Major))
	strSegments[1] = strconv.Itoa(int(ver.Minor))
	strSegments[2] = strconv.Itoa(int(ver.Patch))
	verStr := strings.Join(strSegments, ".")
	return go_version.ParseTolerant(verStr)
}

// CanUseAPIExtV1 returns true if we can use k8s apiextension/v1 else false
func CanUseAPIExtV1() bool {
	return isGEThanAPIExtV1(Version())
}

// CanUseAPIExtV1 returns true if we can use k8s apiextension/v1beta1 else false
func CanUseAPIExtV1Beta1() bool {
	return isGEThanAPIExtV1Beta1(Version())
}

// Version returns the version of the Kubernetes apiserver
func Version() go_version.Version {
	cached.mutex.RLock()
	ver := cached.version
	cached.mutex.RUnlock()
	return ver
}

func updateVersion(version go_version.Version) {
	cached.mutex.Lock()
	defer cached.mutex.Unlock()

	cached.version = version
	glog.Infof("using kubernetes API server version: %s", version)
}

// Update retrieves the version of the Kubernetes apiserver.
// This function must be called after connectivity to the
// apiserver has been established.
func UpdateVersion(client kubernetes.Interface) error {
	sv, err := client.Discovery().ServerVersion()
	if err != nil {
		return err
	}

	// Try GitVersion first. In case of error fallback to MajorMinor
	if sv.GitVersion != "" {
		// This is a string like "v1.9.0"
		ver, err := parseVersion(sv.GitVersion)
		if err == nil {
			updateVersion(ver)
			return nil
		}
	}

	if sv.Major != "" && sv.Minor != "" {
		ver, err := parseVersion(fmt.Sprintf("%s.%s", sv.Major, sv.Minor))
		if err == nil {
			updateVersion(ver)
			return nil
		}
	}

	if err != nil {
		return fmt.Errorf("cannot parse k8s server version from %+v: %s", sv, err)
	}
	return fmt.Errorf("cannot parse k8s server version from %+v", sv)
}
