/*
Copyright 2021 k0s Authors

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

package cleanup

import (
	"fmt"
	"os/exec"

	"github.com/k0sproject/k0s/pkg/constant"
	"github.com/k0sproject/k0s/pkg/crictl"
	"github.com/sirupsen/logrus"
)

type Config struct {
	containerdBinPath    string
	containerdCmd        *exec.Cmd
	containerdSockerPath string
	criCtl               *crictl.CriCtl
	dataDir              string
	runDir               string
	CfgFile              string
	K0sVars              constant.CfgVars
}

func NewConfig(dataDir string) *Config {
	runDir := "/run/k0s" // https://github.com/k0sproject/k0s/pull/591/commits/c3f932de85a0b209908ad39b817750efc4987395
	criSocketPath := fmt.Sprintf("unix:///%s/containerd.sock", runDir)

	return &Config{
		dataDir:              dataDir,
		runDir:               runDir,
		containerdSockerPath: fmt.Sprintf("%s/containerd.sock", runDir),
		containerdBinPath:    fmt.Sprintf("%s/%s", dataDir, "bin/containerd"),
		criCtl:               crictl.NewCriCtl(criSocketPath),
	}
}

func (c *Config) Cleanup() error {
	var msg []error
	cleanupSteps := []Step{
		&containerd{Config: c},
		&users{Config: c},
		&services{Config: c},
		&directories{Config: c},
		&cni{Config: c},
	}

	for _, step := range cleanupSteps {
		if step.NeedsToRun() {
			logrus.Info("* ", step.Name())
			err := step.Run()
			if err != nil {
				logrus.Debug(err)
				msg = append(msg, err)
			}
		}
	}
	if len(msg) > 0 {
		return fmt.Errorf("errors received during clean-up: %v", msg)
	}
	return nil
}

// Step interface is used to implement cleanup steps
type Step interface {
	// NeedsToRun checks if the step needs to run
	NeedsToRun() bool
	// Run impelements specific cleanup operations
	Run() error
	// Name returns name of the step for conveninece
	Name() string
}
