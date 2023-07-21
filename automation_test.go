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
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

// Actions enum. Used by mocks to test the validity of the action sequences.
const (
	playAction = 1 << iota
	isPlayingAction

	TurnSwitchOnAction
	TurnSwitchOffAction
)

type adhanPlayerMock struct {
	isPlaying    bool
	actionLogger *[]int
}

func (a *adhanPlayerMock) Play() error {
	*a.actionLogger = append(*a.actionLogger, playAction)
	return nil
}

func (a *adhanPlayerMock) IsPlaying() bool {
	if a.isPlaying {
		*a.actionLogger = append(*a.actionLogger, isPlayingAction)
		return true
	}
	return false
}

type homeassistantMock struct {
	actionLogger *[]int
}

func (h *homeassistantMock) TurnSwitchOn() (string, error) {
	*h.actionLogger = append(*h.actionLogger, TurnSwitchOnAction)
	return "success", nil
}

func (h *homeassistantMock) TurnSwitchOff() (string, error) {
	*h.actionLogger = append(*h.actionLogger, TurnSwitchOffAction)
	return "success", nil
}

type prayerTimesMock struct {
	prayerTimes
}

func (p *prayerTimesMock) GetTodayPrayerTimes(now time.Time) error {
	parse := func(s string) time.Time {
		c, _ := time.Parse("15:04", s)
		return c
	}

	p.prayerTimes = prayerTimes{
		Fajr:    &prayer{time: parse("09:00")},
		Dhuhr:   &prayer{time: parse("12:00")},
		Asr:     &prayer{time: parse("15:00")},
		Maghrib: &prayer{time: parse("18:00")},
		Ishaa:   &prayer{time: parse("21:00")},
	}
	return nil
}

func TestRunAndSleep(t *testing.T) {
	parse := func(s string) time.Time {
		c, err := time.Parse("15:04", s)
		if err != nil {
			t.Fatalf("Failed to parse the time %v: %v", s, err)
		}
		return c
	}

	for _, test := range []struct {
		description string
		isPlaying   bool
		now         time.Time

		wantSleepDuration  time.Duration
		wantActionSequence []int
	}{
		{
			description: "Adhan currently playing should send automation to sleep",
			isPlaying:   true,

			wantSleepDuration:  FIVE_MINUTES,
			wantActionSequence: []int{isPlayingAction},
		},
		{
			description: "Fajr time should turnSwitchOn and play",
			now:         parse("09:00"),

			wantSleepDuration:  ONE_MINUTE,
			wantActionSequence: []int{TurnSwitchOnAction, playAction},
		},
		{
			description: "1 minutes after Dhuhr should turnSwitchOn and play",
			now:         parse("12:01"),

			wantSleepDuration:  ONE_MINUTE,
			wantActionSequence: []int{TurnSwitchOnAction, playAction},
		},
		{
			description: "10 minutes after Dhuhr should turnSwitchOff and Sleep",
			now:         parse("12:10"),

			// Sleep from 12:10 to 5 minutes to 15:00 (Asr)
			wantSleepDuration:  time.Minute*45 + time.Hour*2,
			wantActionSequence: []int{TurnSwitchOffAction},
		},
		{
			description: "5 minutes before Asr time should Sleep (default 1 minute)",
			now:         parse("14:55"),

			wantSleepDuration:  ONE_MINUTE,
			wantActionSequence: []int{},
		},
		{
			description: "7 minutes before Asr time should sleep 2 minutes (till 5 min before Asr)",
			now:         parse("14:53"),

			wantSleepDuration:  TWO_MINUTES,
			wantActionSequence: []int{TurnSwitchOffAction},
		},
	} {
		t.Run(test.description, func(t *testing.T) {
			actions := []int{}
			a := automation{
				&adhanPlayerMock{isPlaying: test.isPlaying, actionLogger: &actions},
				&homeassistantMock{actionLogger: &actions},
				&prayerTimesMock{}}

			sleepDuration, err := a.RunAndSleep(test.now)
			if err != nil {
				t.Errorf("RunAndSleep expects no error. Got %v", err)
			}
			if sleepDuration != test.wantSleepDuration {
				t.Errorf("RunAndSleep sleep duration mismatch. Got %v, want %v", sleepDuration, test.wantSleepDuration)
			}
			if !cmp.Equal(actions, test.wantActionSequence) {
				t.Errorf("RunAndSleep sleep action seqeuence mismatch. Got %v, want %v", actions, test.wantActionSequence)
			}
		})
	}
}

func TestNewAutomation(t *testing.T) {
	for _, test := range []struct {
		description string
		ha          *homeassistant
		ap          *adhanPlayer
		pt          *munichPrayerTimes
	}{
		{
			description: "Homeassistant is missing",
			ap:          &adhanPlayer{},
			pt:          &munichPrayerTimes{},
		},
		{
			description: "AdhanPlayer is missing",
			ha:          &homeassistant{},
			pt:          &munichPrayerTimes{},
		},
		{
			description: "MunichPrayerTimes is missing",
			ap:          &adhanPlayer{},
			ha:          &homeassistant{},
		},
	} {
		t.Run(test.description, func(t *testing.T) {
			if _, err := NewAutomation(test.ap, test.ha, test.pt); err == nil {
				t.Errorf("NewAutomation expected an error on init. Got none.")
			}
		})
	}
}
