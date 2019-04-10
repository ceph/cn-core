#
# Ceph Nano Core (C) 2018 Red Hat, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
#
# Below main package has canonical imports for 'go get' and 'go build'
# to work with all other clones of github.com/ceph/cn repository. For
# more information refer https://golang.org/doc/go1.4#canonicalimports
#

FROM ceph/daemon:v4.0.0-stable-4.0-nautilus-centos-7-x86_64

ADD cn-core /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/cn-core", "init"]