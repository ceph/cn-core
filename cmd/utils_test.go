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
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGenerateSecret(t *testing.T) {
	key := generateSecret()
	assert.Equal(t, 40, len(key), "Wrong keyring length!")
}

func TestGenerateUUID(t *testing.T) {
	uuid, _ := uuid.NewV4()
	assert.Equal(t, 36, len(uuid.String()), "Wrong UUID length!")
}

func TestGenerateCephConf(t *testing.T) {
	fsid := "7ff73783-cec6-4ace-b655-a6bc4f2532a8"
	hostname := "toto"
	osdMemoryTarget := 1
	osdMemoryBase := 2
	osdMemoryCacheMin := 3
	bluestoreSizeMin := 4
	rgwName := "toto"
	expectedCephConf := `
[global]
fsid = 7ff73783-cec6-4ace-b655-a6bc4f2532a8
mon initial members = toto
mon host = [v2:127.0.0.1:3300,v1:127.0.0.1:6789]
osd crush chooseleaf type = 0
osd journal size = 100
public network = 0.0.0.0/0
cluster network = 0.0.0.0/0
log file = /dev/null
osd pool default size = 1
osd data = /var/lib/ceph/osd/ceph-0
osd objectstore = bluestore
osd memory target = 1
osd memory base = 2
osd memory cache min = 3
bluestore_block_size = 4

[client.rgw.toto]
rgw dns name = toto
rgw enable usage log = true
rgw usage log tick interval = 1
rgw usage log flush threshold = 1
rgw usage max shards = 32
rgw usage max user shards = 1
log file = /var/log/ceph/client.rgw.toto.log
rgw frontends = civetweb port=0.0.0.0:8000
`
	assert.Equal(t, expectedCephConf, fmt.Sprintf(cephConfTemplate, fsid, hostname, osdMemoryTarget, osdMemoryBase, osdMemoryCacheMin, bluestoreSizeMin, rgwName, rgwName, rgwName, rgwEngine, rgwPort), "Ceph configuration file generation error!")
}

func TestValidateAvaibleMemory(t *testing.T) {
	memLimit := 511
	err := validateAvaibleMemory(cnMemMin, memLimit)
	assert.NotNil(t, err)
}

func TestTuneMemory(t *testing.T) {
	memAvailable := uint64(508 * 1024 * 1024)
	osdMemoryTarget, osdMemoryBase, osdMemoryCacheMin := tuneMemory(memAvailable)

	expectedOsdMemoryTarget := uint64(458 * 1024 * 1024)
	assert.Equal(t, expectedOsdMemoryTarget, osdMemoryTarget)

	expectedOsdMemoryBase := uint64(254 * 1024 * 1024)
	assert.Equal(t, expectedOsdMemoryBase, osdMemoryBase)

	expectedOsdMemoryCacheMin := uint64(356 * 1024 * 1024)
	assert.Equal(t, expectedOsdMemoryCacheMin, osdMemoryCacheMin)

}
