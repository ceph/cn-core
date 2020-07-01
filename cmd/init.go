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
	"os"
	"strings"

	"github.com/spf13/cobra"
)

type fn func()

var (
	daemons          string
	dashExposedIP    string
	validValueDaemon = []string{"mon", "mgr", "osd", "rgw", "dash", "health"}
	bootstrapMap     = map[string]fn{
		"mon":  bootstrapMon,
		"mgr":  bootstrapMgr,
		"osd":  bootstrapOsd,
		"rgw":  bootstrapRgw,
		"dash": bootstrapSree,
	}
)

const (
	cnMemMin         uint64 = 512         // minimum amount of memory in MB to run cn-core
	bluestoreSizeMin uint64 = 10737418240 // minimum amount of space for BlueStore in bytes
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
	cmd.Flags().StringVarP(&daemons, "daemon", "d", "", "Specify which daemon to bootstrap. Valid choices are: "+strings.Join(validValueDaemon, ", ")+".")
	cmd.Flags().StringVar(&rgwPort, "rgw-port", rgwPort, "Specify binding port for Rados Gateway.")
	cmd.Flags().StringVar(&dashPort, "dash-port", dashPort, "Specify binding port for Sree dashboard.")
	cmd.Flags().StringVar(&dashExposedIP, "dash-exposed-ip", dashExposedIP, "Specify binding port for Sree dashboard.")

	return cmd
}

// initCluster initialize the Ceph cluster
func initCluster(cmd *cobra.Command, args []string) {
	memLimit := getMemLimit()
	err := validateAvaibleMemory(cnMemMin, memLimit)
	if err != nil {
		log.Fatal(err)
	}
	// validate available bluestore block size, if the user has provided a dedicated directory
	osdPathEnv := os.Getenv("OSD_PATH")
	if len(osdPathEnv) > 0 {
		err := validateAvailableBluestoreSize(bluestoreSizeMin, osdPathEnv)
		if err != nil {
			log.Fatal(err)
		}
	}
	daemonsList := strings.Split(daemons, ",")
	if len(daemonsList) > 0 && daemonsList[0] != "" {
		for i := 0; i < len(validValueDaemon); i++ {
			for j := 0; j < len(daemonsList); j++ {
				if daemonsList[j] == validValueDaemon[i] {
					if daemonsList[j] == "health" {
						err := cephHealth()
						if err != nil {
							log.Fatal(err)
						}
					} else {
						if daemonsList[j] != "" {
							bootstrapMap[daemonsList[j]]()
						}
					}
				}
			}
		}
	} else {
		log.Printf("init: no daemon was selected. Deploying %s.\n", strings.Join(validValueDaemon, ", "))
		bootstrapMon()
		bootstrapMgr()
		bootstrapOsd()
		bootstrapRgw()
		bootstrapSree()
		// bootstrap is done, now watching ceph status
		err := cephHealth()
		if err != nil {
			log.Fatal(err)
		}
	}
	// This makes cn happy when looking for the container status
	fmt.Println("SUCCESS")
}
