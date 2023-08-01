#!/bin/bash

# Copyright 2023 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

if [ -z "$switch_id" ]; then
    echo "Terminating. Please add switch_id as an env variable e.g. -e switch_id=ADDME"
    exit
fi

if [ -z "$homeassistant_ip" ]; then
    echo "Terminating. Please add homeassistant_ip an env variable e.g. -e homeassistant_ip=localhost:8123"
    exit
fi

if [ -z "$homeassistant_token" ]; then
    echo "Terminating. Please add homeassistant_token as an env variable e.g. -e homeassistant_token=ADDME"
    exit
fi

if [ -z "$adhan_mp3_fpath" ]; then
    echo "Terminating. Please add adhan_mp3_fpath as an env variable e.g. -e adhan_mp3_fpath=ADDME"
    exit
fi

exec /adhan-homeassistant-pi \
    --switch_id $switch_id \
    --homeassistant_ip $homeassistant_ip \
    --homeassistant_token $homeassistant_token \
    --adhan_mp3_fpath $adhan_mp3_fpath
