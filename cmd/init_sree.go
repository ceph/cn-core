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
	"log"
	"os"
	"os/exec"

	"github.com/mholt/archiver"
)

const (
	dashboardDirExtractTo = "/opt/ceph-container/sree/"
	dashboardDir          = dashboardDirExtractTo + "Sree-0.1/"
)

var (
	dashPort = "5000"
)

func bootstrapSree() {
	if _, err := os.Stat(dashboardDirExtractTo); os.IsNotExist(err) {
		// Read ENV and search for a value for dashPort
		dashPortEnv := os.Getenv("SREE_PORT")
		if len(dashPortEnv) > 0 {
			dashPort = dashPortEnv
		}

		// run pre-req
		sreePreReq()

		// untar dashboard -  /opt/ceph-container/tmp/
		err := archiver.Unarchive("/opt/ceph-container/tmp/sree.tar.gz", dashboardDirExtractTo)
		if err != nil {
			log.Fatal(err)
		}

		// Always run this, after reboot the IP might change on the host (EXPOSED_IP)
		// This is coming from cn itself
		// configure sree dashboard
		configureClients("dashboard")
	}

	// start cn dashboard!
	sreeStart()
}

func sreePreReq() {
	log.Println("init dashboard: run prerequisites")
	if _, err := os.Stat(dashboardDirExtractTo); os.IsNotExist(err) {
		err = os.MkdirAll(dashboardDirExtractTo, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func sreeStart() {
	log.Println("init dashboard: running dashboard on port " + dashPort)

	cmd := exec.Command("python", "app.py")
	cmd.Dir = dashboardDir

	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
}
