#!/bin/bash

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
