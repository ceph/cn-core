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

func bootstrapMgr() {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal("Failed to get the hostname.")
	}

	monDataPath := cephDataPath + "/mon/ceph-" + hostname
	monKeyringPath := monDataPath + "/keyring"
	mgrDataPath := cephDataPath + "/mgr/ceph-" + hostname
	mgrKeyringPath := mgrDataPath + "/keyring"

	// if there is no key, we assume there is no monitor
	if _, err := os.Stat(mgrKeyringPath); os.IsNotExist(err) {
		// run prereq
		mgrPreReq(mgrDataPath, monKeyringPath)

		// generate mgr keyring
		generateMgrKeyring(hostname, mgrKeyringPath)

		// chown mgr keyring
		err = os.Chown(mgrKeyringPath, cephUID, cephGID)
		if err != nil {
			log.Fatal(err)
		}
	}

	// start ceph mgr!
	mgrStart(hostname)
}

func mgrPreReq(mgrDataPath, monKeyringPath string) {
	log.Println("init mgr: run prerequisites")
	if _, err := os.Stat(mgrDataPath); os.IsNotExist(err) {
		err = os.MkdirAll(mgrDataPath, 0755)
		if err != nil {
			log.Fatal(err)
		}
		err = os.Chown(mgrDataPath, cephUID, cephGID)
		if err != nil {
			log.Fatal(err)
		}
	}

	out, args, err := fetchAdminKeyring(monKeyringPath)
	if err != nil {
		fmt.Printf("The command was: %s\n", args)
		fmt.Printf("The error was: %s\n", out)
		log.Fatal(err)
	}
	err = os.Chown(adminKeyringPath, cephUID, cephGID)
	if err != nil {
		log.Fatal(err)
	}
}

func generateMgrKeyring(hostname, mgrKeyringPath string) {
	log.Println("init mgr: generating manager keyring")

	cmd := exec.Command("ceph", "auth", "get-or-create", "mgr."+hostname, "mon", `allow *`, "-o", mgrKeyringPath)

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("The command was: %s\n", cmd.Args)
		fmt.Printf("The error was: %s\n", out)
		log.Fatal(err)
	}
}

func mgrStart(hostname string) {
	log.Println("init mgr: running manager")

	cmd := exec.Command("ceph-mgr", "--setuser", "ceph", "--setgroup", "ceph", "-i", hostname)

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("The command was: %s\n", cmd.Args)
		fmt.Printf("The error was: %s\n", out)
		log.Fatal(err)
	}
}
