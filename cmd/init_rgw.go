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
	cnUserDetailsFile = "/opt/ceph-container/tmp/cn_user_details"
	s3CmdFilePath     = "/root/.s3cfg"
)

func bootstrapRgw() {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal("Failed to get the hostname.")
	}

	monDataPath := cephDataPath + "/mon/ceph-" + hostname
	monKeyringPath := monDataPath + "/keyring"
	rgwDataPath := cephDataPath + "/radosgw/ceph-rgw." + rgwName
	rgwKeyringPath := rgwDataPath + "/keyring"
	rgwHost := rgwName + ":" + rgwPort

	// if there is no key, we assume there is no monitor
	if _, err := os.Stat(rgwKeyringPath); os.IsNotExist(err) {
		// run prereq
		rgwPreReq(rgwDataPath)

		// generate rgw keyring
		generateRgwKeyring(rgwName, rgwKeyringPath)

		// chown rgw keyring
		err = os.Chown(rgwKeyringPath, cephUID, cephGID)
		if err != nil {
			log.Fatal(err)
		}
	}

	// start rgw!
	rgwStart(rgwName, rgwKeyringPath)

	// create cn user
	if _, err := os.Stat(cnUserDetailsFile); os.IsNotExist(err) {
		// create cn user
		cnUserDetails, err := rgwCreateUser(monKeyringPath)
		if err != nil {
			log.Fatal(err)
		}

		// write cn user details to a file
		err = ioutil.WriteFile(cnUserDetailsFile, cnUserDetails, 0644)
		if err != nil {
			log.Fatal(err)
		}

		// symlink for seemless transition between cn-core and demo.sh
		// so cn can find the credentials
		err = os.Symlink(cnUserDetailsFile, "/nano_user_details")
		if err != nil {
			log.Fatal(err)
		}

		// configure s3cmd
		configureClients("s3cmd", rgwHost)
	}
}

func rgwPreReq(rgwDataPath string) {
	log.Println("init rgw: run prerequisites")
	dirs := [2]string{cephLogPath, rgwDataPath}
	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err = os.MkdirAll(dir, 0755)
			if err != nil {
				log.Fatal(err)
			}
			err = os.Chown(dir, cephUID, cephGID)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func generateRgwKeyring(rgwName, rgwKeyringPath string) {
	log.Println("init rgw: generating rgw keyring")

	cmd := exec.Command("ceph", "auth", "get-or-create", "client.rgw."+rgwName, "mon", `allow rw`, "osd", `allow rwx`, "-o", rgwKeyringPath)

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("The command was: %s\n", cmd.Args)
		fmt.Printf("The error was: %s\n", out)
		log.Fatal(err)
	}
}

func rgwStart(rgwName, rgwKeyringPath string) {
	log.Println("init rgw: running rgw on port " + rgwPort)

	cmd := exec.Command("radosgw", "--setuser", "ceph", "--setgroup", "ceph", "-n", "client.rgw."+rgwName, "-k", rgwKeyringPath)

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("The command was: %s\n", cmd.Args)
		fmt.Printf("The error was: %s\n", out)
		log.Fatal(err)
	}
}

func rgwCreateUser(monKeyringPath string) ([]byte, error) {
	log.Println("init rgw: creating rgw user")

	cmd := exec.Command("radosgw-admin", "user", "create", "--uid="+cnCoreRgwUserUID, "--display-name=Ceph Nano user", "--caps=buckets=*;users=*;usage=*;metadata=*")

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("The command was: %s\n", cmd.Args)
		fmt.Printf("The error was: %s\n", out)
		log.Fatal(err)
	}

	return out, nil
}
