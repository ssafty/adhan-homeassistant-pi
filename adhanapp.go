// Copyright 2023 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"log"
	"time"
)

var (
	speakerSwitchID        = flag.String("switch_id", "Unimplemented", "Id of the speaker switch in home assistant.")
	homeassistantIp        = flag.String("homeassistant_ip", "Unimplemented", "Ip of the local home assistant instance.")
	homeeassistantToken    = flag.String("homeassistant_token", "Unimplemented", "Autherization token for home assistant.")
	adhan_mp3_fpath        = flag.String("adhan_mp3_fpath", "", "Path to the Adhan mp3 file e.g. /Users/userA/adhan.mp3")
	speaker_pause_duration = flag.Duration("speaker_pause", 10*time.Second, "Waiting period between switching on the speaker and playing adhan (default: 10 seconds).")
)

const (
	SAMPLE_RATE     = 44100
	NUM_CHANNELS    = 2
	AUDIO_BIT_DEPTH = 2
)

func sleep(t time.Duration) {
	log.Printf("Sleeping for %v", t)
	time.Sleep(t)
}

func main() {
	flag.Parse()

	homeassistant, err := NewHomeAssistant(
		HTTPClient(NewHTTPClient(*homeeassistantToken)),
		SwitchID(*speakerSwitchID),
		IPAddress(*homeassistantIp))
	if err != nil {
		log.Fatalf("Failed to initialize NewHomeAssistant: %v", err)
	}

	adhanPlayer, err := NewAdhanPlayer(
		FilePath(*adhan_mp3_fpath),
		SamplingRate(SAMPLE_RATE),
		NumChannels(NUM_CHANNELS),
		AudioBitDepth(AUDIO_BIT_DEPTH),
	)
	if err != nil {
		log.Fatalf("Failed to initialize NewAdhanPlayer: %v", err)
	}

	prayerTimes, err := NewMunichPrayerTimes()
	if err != nil {
		log.Fatalf("Failed to initialize NewPrayerTimes: %v", err)
	}

	automation, err := NewAutomation(adhanPlayer, homeassistant, prayerTimes, SpeakerPause(speaker_pause_duration))
	if err != nil {
		log.Fatalf("Failed to initialize NewAutomation: %v", err)
	}

	for {
		sleepDuration, err := automation.RunAndSleep(time.Now())
		if err != nil {
			log.Fatalf("Running the automation failed: %v", err)
		}

		sleep(sleepDuration)
	}
}
