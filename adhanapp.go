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
	speakerSwitchID     = flag.String("switch_id", "Unimplemented", "Id of the speaker switch in home assistant.")
	homeassistantIp     = flag.String("homeassistant_ip", "Unimplemented", "Ip of the local home assistant instance.")
	homeeassistantToken = flag.String("homeassistant_token", "Unimplemented", "Autherization token for home assistant.")
	adhan_mp3_fpath     = flag.String("adhan_mp3_fpath", "", "Path to the Adhan mp3 file e.g. /Users/userA/adhan.mp3")
)

const (
	SAMPLE_RATE     = 44100
	NUM_CHANNELS    = 2
	AUDIO_BIT_DEPTH = 2
)

const (
	FIVE_SECONDS = 5 * time.Second

	ONE_MINUTE   = 1 * time.Minute
	TWO_MINUTES  = 2 * time.Minute
	FIVE_MINUTES = 5 * time.Minute
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

	for {
		// If Adhan is already playing, sleep for 5 minutes. We should loop again after
		// sleep duration to turn off the speakers.
		if adhanPlayer.IsPlaying() {
			sleep(FIVE_MINUTES)
			continue
		}
		now := time.Now()

		times, err := NewPrayerTimes()
		if err != nil {
			log.Fatalf("Failed to retrieve Prayer times: %v", err)
		}

		TimeFromLastPrayer, timeToNextPrayer, err := times.TimesToNearestPrayers(now)
		if err != nil {
			log.Fatalf("Failed to get TimesToNearestPrayers: %v", err)
		}
		log.Printf("Time left till next prayer: %v", timeToNextPrayer)

		switch {
		// Play the Adhan (1) If time for prayer or (2) the last prayer was less than
		// 2 minutes ago and Adhan did not play yet.
		case TimeFromLastPrayer < TWO_MINUTES || timeToNextPrayer == 0:
			if _, err := homeassistant.TurnSwitchOn(); err != nil {
				log.Fatalf("error making a switch action: %v", err)
			}

			// give chance for the speaker to turn on before playing.
			sleep(FIVE_SECONDS)

			if err := adhanPlayer.Play(); err != nil {
				log.Fatalf("error playing the Adhan: %v", err)
			}

		// Turn off speakers and Sleep till 5 minutes before next Prayer.
		case timeToNextPrayer > FIVE_MINUTES:
			if _, err := homeassistant.TurnSwitchOff(); err != nil {
				log.Fatalf("error making a switch action: %v", err)
			}
			log.Printf("Next prayer: %v", timeToNextPrayer)
			sleep(timeToNextPrayer - FIVE_MINUTES)
		default:
			sleep(ONE_MINUTE)
		}
	}
}
