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
	"errors"
	"flag"
	"log"
	"time"
)

var (
	speakerSwitchID        = flag.String("switch_id", "", "Id of the speaker switch in home assistant.")
	homeassistantIp        = flag.String("homeassistant_ip", "", "IP of the local home assistant instance.")
	homeassistantToken     = flag.String("homeassistant_token", "", "Autherization token for home assistant.")
	adhan_mp3_fpath        = flag.String("adhan_mp3_fpath", "adhan.mp3", "Path to the Adhan mp3 file e.g. /Users/userA/adhan.mp3")
	speaker_pause_duration = flag.Duration("speaker_pause", 10*time.Second, "Waiting period between switching on the speaker and playing adhan (default: 10 seconds).")
)

const (
	SAMPLE_RATE     = 44100
	NUM_CHANNELS    = 2
	AUDIO_BIT_DEPTH = 2
)

func sleep(t time.Duration) {
	log.Printf("Sleeping for %v until %v", t, time.Now().Add(t))
	time.Sleep(t)
}

func assertFlags() error {
	switch {
	case *speakerSwitchID == "":
		return errors.New("switch_id flag is not set.")
	case *homeassistantIp == "":
		return errors.New("homeassistant_ip flag is not set.")
	case *homeassistantToken == "":
		return errors.New("homeassistant_token flag is not set.")
	}
	return nil
}

func main() {
	flag.Parse()
	if err := assertFlags(); err != nil {
		log.Fatalf("Some flags are uninitialized: %v", err)
	}

	homeassistant, err := NewHomeAssistant(
		HTTPClient(NewHTTPClient(*homeassistantToken)),
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

	if err := automation.ValidateAllActions(); err != nil {
		log.Fatalf("Failed to validate all actions: %v", err)
	}

	for {
		sleepDuration, err := automation.RunAndSleep(time.Now())
		if err != nil {
			log.Fatalf("Running the automation failed: %v", err)
		}

		sleep(sleepDuration)
	}
}
