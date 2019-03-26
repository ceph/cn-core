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
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
)

const (
	osdDataPath            = cephDataPath + "/osd/ceph-0"
	osdKeyringPath         = osdDataPath + "/keyring"
	osdCrushChooseleafType = "0"
	osdJournalSize         = "100"
	osdObjectstore         = "bluestore"
)

var (
	bluestoreBlockSize = "10737418240"
)

func bootstrapOsd() {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal("Failed to get the hostname.")
	}

	monDataPath := cephDataPath + "/mon/ceph-" + hostname
	monKeyringPath := monDataPath + "/keyring"

	// check for block device
	osdDeviceEnv := os.Getenv("OSD_DEVICE")
	bluestoreBlockSizeEnv := os.Getenv("BLUESTORE_BLOCK_SIZE")
	if len(osdDeviceEnv) > 0 {

		log.Println("init osd: checking for block device")

		testDev, err := getFileType(osdDeviceEnv)
		if err != nil {
			log.Fatal(err)
		}
		if testDev == "blockdev" {
			if len(bluestoreBlockSizeEnv) > 0 {
				size := toBytes(bluestoreBlockSizeEnv)
				// override the default value with the size indicated by the BLUESTORE_BLOCK_SIZE environment variable
				bluestoreBlockSize = string(size)
			} else {
				// using blockdev command to fetch the actual size of the block device
				cmd := exec.Command("blockdev", "--getsize64", osdDeviceEnv)

				out, err := cmd.CombinedOutput()
				if err != nil {
					fmt.Printf("The command was: %s\n", cmd.Args)
					fmt.Printf("The error was: %s\n", out)
					log.Fatal(err)
				} else {
					// override the default value with the actual size of the block device available
					bluestoreBlockSize = string(out)
				}
			}
		} else {
			log.Fatalf("Invalid %s, only block device is supported", osdDeviceEnv)
		}

	}

	// if there is no key, we assume there is no monitor
	if _, err := os.Stat(osdKeyringPath); os.IsNotExist(err) {
		// run prereq
		osdPreReq()

		if len(osdDeviceEnv) > 0 {
			// export client.bootstrap-osd keyring to bootstrap-osd/ceph.keyring file
			cmd := exec.Command("ceph", "auth", "export", "client.bootstrap-osd", "-o", "/var/lib/ceph/bootstrap-osd/ceph.keyring")

			out, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("The command was: %s\n", cmd.Args)
				fmt.Printf("The error was: %s\n", out)
				log.Fatal(err)
			}

			log.Println("init osd: preparing block device")

			cmd = exec.Command("ceph-volume", "lvm", "prepare", "--data", osdDeviceEnv)

			out, err = cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("The command was: %s\n", cmd.Args)
				fmt.Printf("The error was: %s\n", out)
				log.Fatal(err)
			}
		} else {
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
	}

	if len(osdDeviceEnv) > 0 {
		cmd := exec.Command("ceph-volume", "lvm", "list", "--format", "json")

		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("The command was: %s\n", cmd.Args)
			fmt.Printf("The error was: %s\n", out)
			log.Fatal(err)
		} else {
			// fetch the osd fsid value
			osdID := "0"
			var result map[string]interface{}
			json.Unmarshal([]byte(out), &result)
			result1 := result[osdID].([]interface{})
			result2 := result1[0].(map[string]interface{})
			result3 := result2["tags"].(map[string]interface{})
			osdFSID := result3["ceph.osd_fsid"]
			osdFSIDstr, ok := osdFSID.(string)
			if ok {
				log.Println("init osd: activating block device")

				cmd := exec.Command("ceph-volume", "lvm", "activate", "--no-systemd", "--bluestore", "0", osdFSIDstr)

				out, err := cmd.CombinedOutput()
				if err != nil {
					fmt.Printf("The command was: %s\n", cmd.Args)
					fmt.Printf("The error was: %s\n", out)
					log.Fatal(err)
				}

			} else {
				log.Fatal("Could not initiate block device activation. Failed to retrieve osd_fsid.")
			}
		}
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

	cmd := exec.Command("ceph-osd", "--setuser", "ceph", "--setgroup", "ceph", "--conf", cephConfFilePath, "--mkfs", "-i", "0", "--osd-data", osdDataPath)

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("The command was: %s\n", cmd.Args)
		fmt.Printf("The error was: %s\n", out)
		log.Fatal(err)
	}
}

func osdStart() {
	log.Println("init osd: running osd")
	memAvailable := getAvailableRAM()
	osdMemoryTarget, osdMemoryBase, osdMemoryCacheMin := tuneMemory(memAvailable)

	cmd := exec.Command("ceph-osd", "--setuser", "ceph", "--setgroup", "ceph", "-i", "0",
		"--osd-crush-chooseleaf-type", osdCrushChooseleafType,
		"--osd-journal-size", osdJournalSize,
		"--osd-pool-default-size", osdPoolDefaultSize,
		"--osd-objectstore", osdObjectstore,
		"--osd-memory-target", strconv.FormatUint(osdMemoryTarget, 10),
		"--osd-memory-base", strconv.FormatUint(osdMemoryBase, 10),
		"--osd-memory-cache-min", strconv.FormatUint(osdMemoryCacheMin, 10),
		"--bluestore-block-size", bluestoreBlockSize)

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("The command was: %s\n", cmd.Args)
		fmt.Printf("The error was: %s\n", out)
		log.Fatal(err)
	}
}
