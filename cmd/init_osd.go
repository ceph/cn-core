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
	"os/exec"
)

const (
	osdDataPath    = cephDataPath + "/osd/ceph-0"
	osdKeyringPath = osdDataPath + "/keyring"
)

func bootstrapOsd() {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal("Failed to get the hostname.")
	}

	monDataPath := cephDataPath + "/mon/ceph-" + hostname
	monKeyringPath := monDataPath + "/keyring"

	// if there is no key, we assume there is no monitor
	if _, err := os.Stat(osdKeyringPath); os.IsNotExist(err) {
		// run prereq
		osdPreReq()

		// generate osd keyring
		generateOsdKeyring(monKeyringPath, osdKeyringPath)

		// chown osd keyring
		err = os.Chown(osdKeyringPath, cephUID, cephGID)
		if err != nil {
			log.Fatal(err)
		}

		// populate osd store
		osdMkfs(monInitialKeyringPath)
	}

	// chown osd data path
	err = chownR(osdDataPath, cephUID, cephGID)
	if err != nil {
		log.Fatal(err)
	}

	// start ceph osd!
	osdStart()
}

func osdPreReq() {
	log.Println("init osd: run prerequisites")
	if _, err := os.Stat(osdDataPath); os.IsNotExist(err) {
		err = os.MkdirAll(osdDataPath, 0755)
		if err != nil {
			log.Fatal(err)
		}
		err = os.Chown(osdDataPath, cephUID, cephGID)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func generateOsdKeyring(monKeyringPath, osdKeyringPath string) {
	log.Println("init osd: generating osd keyring")

	cmd := exec.Command("ceph", "-n", "mon.", "-k", monKeyringPath, "auth", "get-or-create", "osd.0", "mon", `allow profile osd`, "osd", `allow *`, "mgr", `allow profile osd`, "-o", osdKeyringPath)

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("The command was: %s\n", cmd.Args)
		fmt.Printf("The error was: %s\n", out)
		log.Fatal(err)
	}
}

func osdMkfs(monInitialKeyringPath string) {
	log.Println("init osd: populating osd store")

	cmd := exec.Command("ceph-osd", "--conf", cephConfFilePath, "--mkfs", "-i", "0", "--osd-data", osdDataPath)

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("The command was: %s\n", cmd.Args)
		fmt.Printf("The error was: %s\n", out)
		log.Fatal(err)
	}
}

func osdStart() {
	log.Println("init osd: running osd")

	cmd := exec.Command("ceph-osd", "--setuser", "ceph", "--setgroup", "-i", "0")

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("The command was: %s\n", cmd.Args)
		fmt.Printf("The error was: %s\n", out)
		log.Fatal(err)
	}
}
