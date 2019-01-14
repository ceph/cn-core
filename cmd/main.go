/*
 * Ceph Nano Core (C) 2019 Red Hat, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

/*
 * Below main package has canonical imports for 'go get' and 'go build'
 * to work with all other clones of github.com/ceph/cn repository. For
 * more information refer https://golang.org/doc/go1.4#canonicalimports
 */

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const (
	cliName          = "cn-core"
	cliDescription   = `Ceph Nano Core - Bootstrap Ceph AIO.`
	cephDataPath     = "/var/lib/ceph"
	cephConfigPath   = "/etc/ceph"
	cephConfFilePath = cephConfigPath + "/ceph.conf"
	cnCoreRgwUserUID = "cn"
	adminKeyringPath = "/etc/ceph/ceph.client.admin.keyring"
	cephLogPath      = "/var/log/ceph"
	cephUID          = 167 // 167 is Ceph'user ID and Group on CentOS systems
	cephGID          = 167 // 167 is Ceph'user GID and Group on CentOS systems
)

var (
	// cnCoreVersion is the version
	cnCoreVersion = "undefined"

	rootCmd = &cobra.Command{
		Use:        cliName,
		Short:      cliDescription,
		SuggestFor: []string{"cn-core"},
	}
)

// Main is the main function calling the whole program
func Main(version string) {
	cnCoreVersion = version

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {

	rootCmd.AddCommand(
		cliInitCluster(),
		cliVersionCnCore(),
	)
	rootCmd.SetHelpCommand(&cobra.Command{
		Use:    "no-help",
		Hidden: true,
	})
	cobra.EnableCommandSorting = false

}
