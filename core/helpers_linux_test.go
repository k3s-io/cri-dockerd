//go:build linux
// +build linux

/*
Copyright 2021 Mirantis

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

package core

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSeccompSecurityOpts(t *testing.T) {
	tests := []struct {
		msg            string
		seccompProfile *runtimeapi.SecurityProfile
		expectedOpts   []string
	}{{
		msg:            "No security annotations",
		seccompProfile: nil,
		expectedOpts:   []string{"seccomp=unconfined"},
	}, {
		msg: "Seccomp unconfined",
		seccompProfile: &runtimeapi.SecurityProfile{
			ProfileType: runtimeapi.SecurityProfile_Unconfined,
		},
		expectedOpts: []string{"seccomp=unconfined"},
	}, {
		msg: "Seccomp default",
		seccompProfile: &runtimeapi.SecurityProfile{
			ProfileType: runtimeapi.SecurityProfile_RuntimeDefault,
		},
		expectedOpts: nil,
	}, {
		msg: "Seccomp deprecated default",
		seccompProfile: &runtimeapi.SecurityProfile{
			ProfileType:  runtimeapi.SecurityProfile_RuntimeDefault,
			LocalhostRef: "docker/default",
		},
		expectedOpts: nil,
	}}

	for i, test := range tests {
		opts, err := getSeccompSecurityOpts(test.seccompProfile, '=')
		assert.NoError(t, err, "TestCase[%d]: %s", i, test.msg)
		assert.Len(t, opts, len(test.expectedOpts), "TestCase[%d]: %s", i, test.msg)
		for _, opt := range test.expectedOpts {
			assert.Contains(t, opts, opt, "TestCase[%d]: %s", i, test.msg)
		}
	}
}

func TestLoadSeccompLocalhostProfiles(t *testing.T) {
	tmpdir := filepath.Join(os.TempDir(), "seccomp-local-profile-test")
	defer os.RemoveAll(tmpdir)
	testProfile := `{"foo": "bar"}`
	err := os.WriteFile(filepath.Join(tmpdir, "test"), []byte(testProfile), 0644)
	require.NoError(t, err)

	tests := []struct {
		msg            string
		seccompProfile *runtimeapi.SecurityProfile
		expectedOpts   []string
		expectErr      bool
	}{{
		msg: "Seccomp localhost/test profile should return correct seccomp profiles",
		seccompProfile: &runtimeapi.SecurityProfile{
			ProfileType:  runtimeapi.SecurityProfile_Localhost,
			LocalhostRef: filepath.Join(tmpdir, "test"),
		},
		expectedOpts: []string{`seccomp={"foo":"bar"}`},
		expectErr:    false,
	}, {
		msg: "Non-existent profile should return error",
		seccompProfile: &runtimeapi.SecurityProfile{
			ProfileType:  runtimeapi.SecurityProfile_Localhost,
			LocalhostRef: "localhost/" + filepath.Join(tmpdir, "fixtures/non-existent"),
		},
		expectedOpts: nil,
		expectErr:    true,
	}, {
		msg: "Relative profile path should return error",
		seccompProfile: &runtimeapi.SecurityProfile{
			ProfileType:  runtimeapi.SecurityProfile_Localhost,
			LocalhostRef: "localhost/fixtures/test",
		},
		expectedOpts: nil,
		expectErr:    true,
	}}

	for i, test := range tests {
		opts, err := getSeccompSecurityOpts(test.seccompProfile, '=')
		if test.expectErr {
			assert.Error(t, err, fmt.Sprintf("TestCase[%d]: %s", i, test.msg))
			continue
		}
		assert.NoError(t, err, "TestCase[%d]: %s", i, test.msg)
		assert.Len(t, opts, len(test.expectedOpts), "TestCase[%d]: %s", i, test.msg)
		for _, opt := range test.expectedOpts {
			assert.Contains(t, opts, opt, "TestCase[%d]: %s", i, test.msg)
		}
	}
}
