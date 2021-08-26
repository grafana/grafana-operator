#!/bin/bash -ex
#
# Copyright (c) 2019 Red Hat, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

# This script is executed by a Jenkins job for each change request. If it
# doesn't succeed the change won't be merged.

OPSDK_VERSION=v0.18.2
OPSDK_DOWNLOAD_URL="https://github.com/operator-framework/operator-sdk/releases/download/v0.18.2/operator-sdk-$OPSDK_VERSION-x86_64-linux-gnu"

mkdir -p ${PWD}/.bin

# Download OPM
if [ ! -f "${PWD}/.bin/operator-sdk" ]; then
  curl -Lo "${PWD}/.bin/operator-sdk" -k $OPSDK_DOWNLOAD_URL
  chmod +x "${PWD}/.bin/operator-sdk"
fi

export OPERATOR_SDK=${PWD}/.bin/operator-sdk

# Print OPM version
$OPERATOR_SDK version

make image/build