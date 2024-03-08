/*
   Copyright The KWasm Authors.

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

package shim

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/spinkube/runtime-class-manager/pkg/config"
	"github.com/spinkube/runtime-class-manager/pkg/state"
)

func Install(config *config.Config, shimName string) (string, bool, error) {
	shimPath := config.AssetPath(shimName)
	srcFile, err := os.OpenFile(shimPath, os.O_RDONLY, 0o000)
	if err != nil {
		return "", false, fmt.Errorf("error opening shim: %w", err)
	}
	dstFilePath := path.Join(config.Kwasm.Path, "bin", shimName)
	dstFilePathHost := config.PathWithHost(dstFilePath)

	err = os.MkdirAll(path.Dir(dstFilePathHost), 0o755)
	if err != nil {
		return dstFilePath, false, fmt.Errorf("error creating directory: %w", err)
	}

	dstFile, err := os.OpenFile(dstFilePathHost, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o755)
	if err != nil {
		return "", false, fmt.Errorf("error opening file: %w", err)
	}

	st, err := state.Get(config)
	if err != nil {
		return "", false, fmt.Errorf("error getting state: %w", err)
	}
	shimSha256 := sha256.New()

	_, err = io.Copy(io.MultiWriter(dstFile, shimSha256), srcFile)
	runtimeName := RuntimeName(shimName)
	changed := st.ShimChanged(runtimeName, shimSha256.Sum(nil), dstFilePath)
	if changed {
		st.UpdateShim(runtimeName, state.Shim{
			Path:   dstFilePath,
			Sha256: shimSha256.Sum(nil),
		})
		if err := st.Write(); err != nil {
			return "", false, fmt.Errorf("error writing state: %w", err)
		}
	}

	return dstFilePath, changed, fmt.Errorf("error installing shim: %w", err)
}