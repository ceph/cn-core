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
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gofrs/uuid"
)

func generateHeader() []byte {
	header := new(bytes.Buffer)
	var data = []interface{}{
		uint16(1),
		uint32(time.Now().UnixNano()),
		uint32(0),
		uint16(16),
	}

	for _, v := range data {
		// assume LittleEndian for Byte Order, this then assumes x86 system
		// Note from https://docs.python.org/2/library/struct.html#byte-order-size-and-alignment
		// Native byte order is big-endian or little-endian, depending on the host system.
		// For example, Intel x86 and AMD64 (x86-64) are little-endian; Motorola 68000 and PowerPC G5 are big-endian;
		// ARM and Intel Itanium feature switchable endianness (bi-endian).
		err := binary.Write(header, binary.LittleEndian, v)
		if err != nil {
			log.Fatal("binary.Write failed with:", err)
		}
	}

	return header.Bytes()
}

func generateSecret() string {
	key := make([]byte, 16)

	// generates random bytes from 'key'
	_, err := rand.Read(key)
	if err != nil {
		log.Fatal(err)
	}

	// generate header
	header := generateHeader()

	// merge slices
	secretByte := append(header[:], key[:]...)

	return base64.StdEncoding.EncodeToString(secretByte)
}

func generateUUID() string {
	uuid, err := uuid.NewV4()
	if err != nil {
		log.Fatalf("failed to generate UUID: %v", err)
	}

	return uuid.String()
}

func generateCephConf(hostname, rgwEngine, rgwPort string) (string, string) {
	fsid := generateUUID()

	return fmt.Sprintf(cephConfTemplate, fsid, hostname, hostname, hostname, hostname, rgwEngine, rgwPort), fsid
}

func writeCephConf(hostname, cephConfFilePath string) string {
	log.Println("init mon: writing ceph configuration file")

	cephConf, fsid := generateCephConf(hostname, rgwEngine, rgwPort)
	cephConfBytes := []byte(cephConf)

	err := ioutil.WriteFile(cephConfFilePath, cephConfBytes, 0644)
	if err != nil {
		log.Fatal(err)
	}

	return fsid
}

// Thanks https://stackoverflow.com/questions/33161284/recursively-create-a-directory-with-a-certain-owner-and-group
func chownR(path string, uid, gid int) error {
	return filepath.Walk(path, func(name string, info os.FileInfo, err error) error {
		if err == nil {
			err = os.Chown(name, uid, gid)
		}
		return err
	})
}

func getAwsKeys() (string, string) {
	jsonFile, err := os.Open(cnUserDetailsFile)
	if err != nil {
		fmt.Println(err)
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	// declare structures for json
	type s3Details []struct {
		AccessKey string `json:"Access_key"`
		SecretKey string `json:"Secret_key"`
	}
	type jason struct {
		Keys s3Details
	}
	// assign variable to our json struct
	var parsedMap jason

	json.Unmarshal(byteValue, &parsedMap)

	cnAccessKey := parsedMap.Keys[0].AccessKey
	cnSecretKey := parsedMap.Keys[0].SecretKey

	return cnAccessKey, cnSecretKey
}

func fetchAdminKeyring(monKeyringPath string) ([]byte, []string, error) {
	log.Println("init mgr: fetching admin keyring")

	cmd := exec.Command("ceph", "-n", "mon.", "-k", monKeyringPath, "auth", "get-or-create", "client.admin", "-o", adminKeyringPath)
	out, err := cmd.CombinedOutput()
	return out, cmd.Args, err
}

func sedFile(path, old, new string) {
	read, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	newContents := strings.Replace(string(read), old, new, -1)

	err = ioutil.WriteFile(path, []byte(newContents), 0)
	if err != nil {
		log.Fatal(err)
	}
}

func configureClients(client string, arg ...string) {
	cnAccessKey, cnSecretKey := getAwsKeys()

	switch client {
	case "s3cmd":
		log.Println("init rgw: configure s3cmd client")
		sedFile(s3CmdFilePath, "AWS_ACCESS_KEY_PLACEHOLDER", cnAccessKey)
		sedFile(s3CmdFilePath, "AWS_SECRET_KEY_PLACEHOLDER", cnSecretKey)
		sedFile(s3CmdFilePath, "localhost", arg[0]) // this is always one arg, not sure why making the string default makes it a slice...

	case "dashboard":
		// Read ENV and search for a value for dashExposedIP
		dashExposedIPEnv := os.Getenv("EXPOSED_IP")
		if len(dashExposedIPEnv) > 0 {
			dashExposedIP = dashExposedIPEnv
		}

		log.Println("init dashboard: configure dashboard")
		path := dashboardDir + "static/js/base.js"
		sedFile(path, "ENDPOINT", dashExposedIP+":"+rgwPort)
		sedFile(path, "ACCESS_KEY", cnAccessKey)
		sedFile(path, "SECRET_KEY", cnSecretKey)

		err := os.Link(dashboardDir+"sree.cfg.sample", dashboardDir+"sree.cfg")
		if err != nil {
			log.Fatal(err)
		}
		path = dashboardDir + "sree.cfg"
		sedFile(path, "RGW_CIVETWEB_PORT_VALUE", rgwPort)
		sedFile(path, "SREE_PORT_VALUE", dashPort)
	}

}

func cephHealth() error {
	// A WaitGroup waits for a collection of goroutines to finish.
	// This is useful for us since we are collecting both stderr and stdout
	// We have to be able to tell if one of two failed
	// This is for error handling purpose
	// Inspiration from https://github.com/boz/ephemerald/blob/master/lifecycle/action_exec.go#L130-L163
	var wg sync.WaitGroup

	log.Println("init: running ceph health watcher")

	// declare command to execute
	cmd := exec.Command("ceph", "-w")

	// get an io reader for stdout
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	// get an io reader for stderr
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	// start the command
	if err := cmd.Start(); err != nil {
		return err
	}

	// Add 2 waiters since we have two goroutine
	wg.Add(2)

	// Go routine that reads stdout
	go func() {
		defer wg.Done()
		readPipe(stdout)
	}()

	// Go routine that reads stderr
	go func() {
		defer wg.Done()
		readPipe(stderr)
	}()

	// Wait for both waiters to complete
	// If stdout succeeds this means waiting 'forever'
	wg.Wait()

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}

func readPipe(reader io.Reader) {
	r := bufio.NewReader(reader)

	for true {
		line, _, _ := r.ReadLine()
		if line != nil {
			outStr := string(line)
			fmt.Println(outStr)
		} else {
			// this means EOF and we stop the iteration
			// this likely means the 'ceph -w' died
			break
		}
	}
}
