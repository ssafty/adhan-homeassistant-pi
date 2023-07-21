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

// automation contains all the logic for adhan playing, waiting and rewinding.

package main

import (
	"errors"
	"fmt"
	"log"
	"time"
)

const (
	FIVE_SECONDS = 5 * time.Second

	ONE_MINUTE   = 1 * time.Minute
	TWO_MINUTES  = 2 * time.Minute
	FIVE_MINUTES = 5 * time.Minute
)

type automation struct {
	adhanPlayer   IAdhanPlayer
	homeassistant IHomeAssistant
	prayerTimes   IPrayerTimes
}

func NewAutomation(ap *adhanPlayer, ha *homeassistant, pa *munichPrayerTimes) (*automation, error) {
	if ap == nil {
		return nil, errors.New("Automation expects a non-nil AdhanPlayer.")
	}
	if ha == nil {
		return nil, errors.New("Automation expects a non-nil Homeassistant instance.")
	}
	if pa == nil {
		return nil, errors.New("Automation expects a non-nil PrayerTimes instance.")
	}
	return &automation{ap, ha, pa}, nil
}

// returns sleep amount
func (a *automation) RunAndSleep(now time.Time) (time.Duration, error) {
	if a.adhanPlayer.IsPlaying() {
		return FIVE_MINUTES, nil
	}

	if err := a.prayerTimes.GetTodayPrayerTimes(now); err != nil {
		return 0, fmt.Errorf("Failed to repopulate Prayertimes: %w", err)
	}

	prevPrayer, nextPrayer, err := a.prayerTimes.GetNearestPrayers(now)
	if err != nil {
		return 0, fmt.Errorf("Failed to get TimesToNearestPrayers: %w", err)
	}

	timeFromPrevPrayer := prevPrayer.TimeToPrayer(now)
	timeToNextPrayer := nextPrayer.TimeToPrayer(now)
	log.Printf("Time left till %v: %v", nextPrayer, timeToNextPrayer)

	switch {
	// Play the Adhan (1) If time for prayer or (2) the last prayer was less than
	// 2 minutes ago and Adhan did not play yet.
	case timeFromPrevPrayer < TWO_MINUTES || timeToNextPrayer == 0:
		if _, err := a.homeassistant.TurnSwitchOn(); err != nil {
			return 0, fmt.Errorf("error making a switch action: %w", err)
		}

		// give chance for the speaker to turn on before playing.
		sleep(FIVE_SECONDS)

		if err := a.adhanPlayer.Play(); err != nil {
			return 0, fmt.Errorf("error playing the Adhan: %w", err)
		}

	// Turn off speakers and Sleep till 5 minutes before next Prayer.
	case timeToNextPrayer > FIVE_MINUTES:
		if _, err := a.homeassistant.TurnSwitchOff(); err != nil {
			return 0, fmt.Errorf("error making a switch action: %w", err)
		}
		log.Printf("Next prayer: %v", timeToNextPrayer)
		return timeToNextPrayer - FIVE_MINUTES, nil
	}

	return ONE_MINUTE, nil
}