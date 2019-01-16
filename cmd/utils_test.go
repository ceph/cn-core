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

func TestGenerateHeader(t *testing.T) {
	header := generateHeader()
	assert.Equal(t, 12, len(header), "Wrong header length!")
}

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
	expectedCephConf := `
[global]
fsid = 7ff73783-cec6-4ace-b655-a6bc4f2532a8
mon initial members = toto
mon host = 127.0.0.1
osd crush chooseleaf type = 0
osd journal size = 100
public network = 0.0.0.0/0
cluster network = 0.0.0.0/0
log file = /dev/null
osd pool default size = 1
osd data = /var/lib/ceph/osd/ceph-0
osd objectstore = bluestore

[client.rgw.toto]
rgw dns name = toto
rgw enable usage log = true
rgw usage log tick interval = 1
rgw usage log flush threshold = 1
rgw usage max shards = 32
rgw usage max user shards = 1
log file = /var/log/ceph/client.rgw.toto.log
rgw frontends = civetweb port=0.0.0.0:8000
rgw enable static website = true

[client.rgw.toto.static]
rgw dns name = toto
rgw dns s3website name = toto.static
rgw enable apis = s3, s3website
rgw resolve cname = true
`
	assert.Equal(t, expectedCephConf, fmt.Sprintf(cephConfTemplate, fsid, hostname, hostname, hostname, hostname, rgwEngine, rgwPort, hostname, hostname, hostname), "Ceph configuration file generation error!")
}

// func TestTuneMemory(t *testing.T) {

// }
