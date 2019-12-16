#!/usr/bin/env bash

# Copyright 2019 The Tekton Authors
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

set -o errexit
set -o nounset
set -o pipefail

go get -d github.com/tektoncd/plumbing
source $(go list -m -f '{{.Dir}}' github.com/tektoncd/plumbing)/scripts/library.sh

cd ${REPO_ROOT_DIR}

# Prune modules.
go mod tidy
go mod vendor

update_licenses third_party/