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
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

const (
	monInitialKeyringTemplate = `
[mon.]
	key = %s
	caps mon = "allow *"
`
	cephConfTemplate = `
[global]
fsid = %s
mon initial members = %s
mon host = 127.0.0.1
osd crush chooseleaf type = 0
osd journal size = 100
public network = 0.0.0.0/0
cluster network = 0.0.0.0/0
log file = /dev/null
osd pool default size = 1
osd data = /var/lib/ceph/osd/ceph-0
osd objectstore = bluestore

[client.rgw.%s]
rgw dns name = %s
rgw enable usage log = true
rgw usage log tick interval = 1
rgw usage log flush threshold = 1
rgw usage max shards = 32
rgw usage max user shards = 1
log file = /var/log/ceph/client.rgw.%s.log
rgw frontends = %s port=0.0.0.0:%s
`

	monMapPath            = "/etc/ceph/monmap"
	monInitialKeyringPath = "/etc/ceph/initial-mon-keyring"
	monIP                 = "127.0.0.1"
	monListenIPPort       = monIP + ":" + monPort
	rgwEngine             = "civetweb"
	monPort               = "6789"
)

var (
	rgwPort = "8000"
)

func bootstrapMon() {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal("Failed to get the hostname.")
	}

	// Read ENV and search for a value for rgwPort
	rgwPortEnv := os.Getenv("RGW_CIVETWEB_PORT")
	if len(rgwPortEnv) > 0 {
		rgwPort = rgwPortEnv
	}

	monDataPath := cephDataPath + "/mon/ceph-" + hostname
	monKeyringPath := monDataPath + "/keyring"

	// if there is no key, we assume there is no monitor
	if _, err := os.Stat(monKeyringPath); os.IsNotExist(err) {
		// run prereq
		monPreReq(monDataPath)

		// write mon initial keyring
		writeKeyring(monInitialKeyringPath)

		// write ceph.conf
		fsid := writeCephConf(hostname, cephConfFilePath)
		err = os.Chown(cephConfFilePath, cephUID, cephGID)
		if err != nil {
			log.Fatal(err)
		}

		// generate monmap
		generateMonMap(hostname, fsid, monMapPath)

		// chown monmap
		err = os.Chown(monMapPath, cephUID, cephGID)
		if err != nil {
			log.Fatal(err)
		}

		// populate mon store
		monMkfs(hostname, monInitialKeyringPath, monDataPath, monMapPath)
	}

	// start ceph mon!
	monStart(hostname, monDataPath)
}

func monPreReq(monDataPath string) {
	log.Println("init mon: run prerequisites")
	if _, err := os.Stat(monDataPath); os.IsNotExist(err) {
		err = os.MkdirAll(monDataPath, 0755)
		if err != nil {
			log.Fatal(err)
		}
		err = os.Chown(monDataPath, cephUID, cephGID)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func generateMonInitialKeyring() string {
	monInitialKeyring := generateSecret()

	return fmt.Sprintf(monInitialKeyringTemplate, monInitialKeyring)
}

func writeKeyring(monInitialKeyringPath string) error {
	log.Println("init mon: writing monitor initial keyring")
	keyring := generateMonInitialKeyring()
	keyringBytes := []byte(keyring)

	if err := ioutil.WriteFile(monInitialKeyringPath, []byte(keyringBytes), 0600); err != nil {
		return fmt.Errorf("failed to write monitor keyring to %s: %+v", monInitialKeyringPath, err)
	}

	err := os.Chown(monInitialKeyringPath, cephUID, cephGID)
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func generateMonMap(hostname, fsid, monMapPath string) {
	log.Println("init mon: generating monitor map")

	cmd := exec.Command("monmaptool", "--create", "--add", hostname, monListenIPPort, "--fsid", fsid, monMapPath)

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("The command was: %s\n", cmd.Args)
		fmt.Printf("The error was: %s\n", out)
		log.Fatal(err)
	}
}

func monMkfs(hostname, monInitialKeyringPath, monDataPath, monMapPath string) {
	log.Println("init mon: populating monitor store")

	cmd := exec.Command("ceph-mon", "--setuser", "ceph", "--setgroup", "ceph", "--mkfs", "-i", hostname, "--inject-monmap", monMapPath, "--keyring", monInitialKeyringPath, "--mon-data", monDataPath)

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("The command was: %s\n", cmd.Args)
		fmt.Printf("The error was: %s\n", out)
		log.Fatal(err)
	}
}

func monStart(hostname, monDataPath string) {
	log.Println("init mon: running monitor")

	cmd := exec.Command("ceph-mon", "--setuser", "ceph", "--setgroup", "ceph", "-i", hostname, "--mon-data", monDataPath, "--public-addr", monListenIPPort)

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("The command was: %s\n", cmd.Args)
		fmt.Printf("The error was: %s\n", out)
		log.Fatal(err)
	}
}
