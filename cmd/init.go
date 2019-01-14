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
	"log"
	"strings"

	"github.com/spf13/cobra"
)

var (
	daemon           string
	dashExposedIP    string
	validValueDaemon = []string{"mon", "mgr", "osd", "rgw", "dash", "health"}
)

// cliInitCluster is the Cobra CLI call
func cliInitCluster() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Init a Ceph cluster",
		Args:  cobra.NoArgs,
		Run:   initCluster,
		Example: "cn-core init\n" +
			"cn-core init --daemon mon \n",
	}
	cmd.Flags().SortFlags = false
	cmd.Flags().StringVarP(&daemon, "daemon", "d", "", "Specify which daemon to bootstrap. Valid choices are: "+strings.Join(validValueDaemon, ", ")+".")
	cmd.Flags().StringVar(&rgwPort, "rgw-port", rgwPort, "Specify binding port for Rados Gateway.")
	cmd.Flags().StringVar(&dashPort, "dash-port", dashPort, "Specify binding port for Sree dashboard.")
	cmd.Flags().StringVar(&dashExposedIP, "dash-exposed-ip", dashExposedIP, "Specify binding port for Sree dashboard.")

	return cmd
}

// initCluster initialize the Ceph cluster
func initCluster(cmd *cobra.Command, args []string) {
	switch daemon {
	case "mon":
		bootstrapMon()
	case "mgr":
		bootstrapMgr()
	case "osd":
		bootstrapOsd()
	case "rgw":
		bootstrapRgw()
	case "dash":
		bootstrapSree()
	case "health":
		err := cephHealth()
		if err != nil {
			log.Fatal(err)
		}
	default:
		log.Printf("init: no daemon was selected. Deploying %s.\n", strings.Join(validValueDaemon, ", "))
		bootstrapMon()
		bootstrapMgr()
		bootstrapOsd()
		bootstrapRgw()
		bootstrapSree()

		// This makes cn happy when looking for the container status
		fmt.Println("SUCCESS")

		// bootstrap is done, now watching ceph status
		err := cephHealth()
		if err != nil {
			log.Fatal(err)
		}
	}
}
