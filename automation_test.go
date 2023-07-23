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
	aPlay = 1 << iota
	aIsPlaying

	aTurnSwitchOn
	aTurnSwitchOff
)

type adhanPlayerMock struct {
	forcePlay    bool
	actionLogger *[]int

	isPlaying bool
}

func (a *adhanPlayerMock) Play() error {
	a.isPlaying = true
	*a.actionLogger = append(*a.actionLogger, aPlay)
	return nil
}

func (a *adhanPlayerMock) IsPlaying() bool {
	if a.forcePlay || a.isPlaying {
		a.isPlaying = false
		*a.actionLogger = append(*a.actionLogger, aIsPlaying)
		return true
	}
	return false
}

type homeassistantMock struct {
	actionLogger *[]int
}

func (h *homeassistantMock) TurnSwitchOn() (string, error) {
	*h.actionLogger = append(*h.actionLogger, aTurnSwitchOn)
	return "success", nil
}

func (h *homeassistantMock) TurnSwitchOff() (string, error) {
	*h.actionLogger = append(*h.actionLogger, aTurnSwitchOff)
	return "success", nil
}

type prayerTimesMock struct {
	prayerTimes
}

func (p *prayerTimesMock) GetTodayPrayerTimes(now time.Time) error {
	parse := func(s string) time.Time {
		c, _ := time.Parse("15:04", s)
		return time.Date(now.Year(), now.Month(), now.Day(), c.Hour(), c.Minute(), 0, 0, now.Location())
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
		forcePlay   bool
		now         time.Time

		wantSleepDuration  time.Duration
		wantActionSequence []int
	}{
		{
			description: "Adhan currently playing should send automation to sleep",
			forcePlay:   true,

			wantSleepDuration:  FIVE_MINUTES,
			wantActionSequence: []int{aIsPlaying},
		},
		{
			description: "Fajr time should turnSwitchOn and play",
			now:         parse("09:00"),

			wantSleepDuration:  ONE_MINUTE,
			wantActionSequence: []int{aTurnSwitchOn, aPlay},
		},
		{
			description: "1 minutes after Dhuhr should turnSwitchOn and play",
			now:         parse("12:01"),

			wantSleepDuration:  ONE_MINUTE,
			wantActionSequence: []int{aTurnSwitchOn, aPlay},
		},
		{
			description: "10 minutes after Dhuhr should turnSwitchOff and Sleep",
			now:         parse("12:10"),

			// Sleep from 12:10 to 5 minutes to 15:00 (Asr)
			wantSleepDuration:  time.Minute*45 + time.Hour*2,
			wantActionSequence: []int{aTurnSwitchOff},
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
			wantActionSequence: []int{aTurnSwitchOff},
		},
	} {
		t.Run(test.description, func(t *testing.T) {
			actions := []int{}
			a := automation{
				&adhanPlayerMock{forcePlay: test.forcePlay, actionLogger: &actions},
				&homeassistantMock{actionLogger: &actions},
				&prayerTimesMock{},
				func() *time.Duration { d := time.Duration(0 * time.Second); return &d }()}

			sleepDuration, err := a.RunAndSleep(test.now)
			if err != nil {
				t.Errorf("RunAndSleep expects no error. Got %v", err)
			}
			if sleepDuration != test.wantSleepDuration {
				t.Errorf("RunAndSleep sleep duration mismatch. Got %v, want %v", sleepDuration, test.wantSleepDuration)
			}
			if !cmp.Equal(actions, test.wantActionSequence) {
				t.Errorf("RunAndSleep sleep action sequence mismatch. Got %v, want %v", actions, test.wantActionSequence)
			}
		})
	}
}

// TestRunAndSleepIntegration emulates an entire day (+spillover) of decision making.
func TestRunAndSleepIntegration(t *testing.T) {
	parse := func(s string) time.Time {
		c, err := time.Parse("15:04", s)
		if err != nil {
			t.Fatalf("Failed to parse the time %v: %v", s, err)
		}
		return c
	}

	startingTime := parse("01:04")
	maxActions := 60

	gotActions := []int{}
	gotTotalSleep := time.Minute * 0

	a := automation{
		&adhanPlayerMock{isPlaying: false, actionLogger: &gotActions},
		&homeassistantMock{actionLogger: &gotActions},
		&prayerTimesMock{},
		func() *time.Duration { d := time.Duration(0 * time.Second); return &d }()}

	for i := 0; i < maxActions; i++ {
		sleepDuration, err := a.RunAndSleep(startingTime)
		if err != nil {
			t.Errorf("RunAndSleep expects no error. Got %v", err)
		}

		gotTotalSleep += sleepDuration
		startingTime = startingTime.Add(sleepDuration)
	}

	if wantActionSequence := []int{aTurnSwitchOff,
		aTurnSwitchOn, aPlay, aIsPlaying, aTurnSwitchOff,
		aTurnSwitchOn, aPlay, aIsPlaying, aTurnSwitchOff,
		aTurnSwitchOn, aPlay, aIsPlaying, aTurnSwitchOff,
		aTurnSwitchOn, aPlay, aIsPlaying, aTurnSwitchOff,
		aTurnSwitchOn, aPlay, aIsPlaying, aTurnSwitchOff,
		// next day
		aTurnSwitchOff, // Redundant because the previous action sleeps for 23 hours.
		aTurnSwitchOn, aPlay, aIsPlaying, aTurnSwitchOff,
		aTurnSwitchOn, aPlay, aIsPlaying, aTurnSwitchOff,
	}; !cmp.Equal(gotActions, wantActionSequence) {
		t.Errorf("RunAndSleep sleep action sequence mismatch. Got %v, want %v", gotActions, wantActionSequence)
	}

	// sleep(7h51min) = time.now("01:04") - fajr time - 5 minutes.
	// (a) sleep(1 min) till prayer times. This will happen 6 times.
	// (b) sleep(5 min) till Adhan is done playing. Then sleep(1h49min) till next prayer.
	// both (a) and (b) will happen 4 more times a day. At Ishaa, sleep(5 min)
	// till adhan is done playing then sleep till (Fajr + 23 hours - 5 minutes).
	// Next day, We are 1 hour away from Fajr - 5 minutes. sleep(1 hour).
	// repeat (a) and (b) for two more prayers.
	// sleep(1 minute) * 2 at the end as we reached our decisions limit.
	if want := 37*time.Hour + 53*time.Minute; gotTotalSleep != want {
		t.Errorf("RunAndSleep total sleep duration mismatch. Got %v, want %v", gotTotalSleep, want)
	}
}

func TestNewAutomation(t *testing.T) {
	for _, test := range []struct {
		description string
		ha          *homeassistant
		ap          *adhanPlayer
		pt          *munichPrayerTimes
		pause       time.Duration
	}{
		{
			description: "Homeassistant is missing",
			ap:          &adhanPlayer{},
			pt:          &munichPrayerTimes{},
			pause:       time.Second,
		},
		{
			description: "AdhanPlayer is missing",
			ha:          &homeassistant{},
			pt:          &munichPrayerTimes{},
			pause:       time.Second,
		},
		{
			description: "MunichPrayerTimes is missing",
			ap:          &adhanPlayer{},
			ha:          &homeassistant{},
			pause:       time.Second,
		},
		{
			description: "Speaker pause is missing",
			ap:          &adhanPlayer{},
			pt:          &munichPrayerTimes{},
			ha:          &homeassistant{},
		},
		{
			description: "Speaker pause is negative",
			ap:          &adhanPlayer{},
			pt:          &munichPrayerTimes{},
			ha:          &homeassistant{},
			pause:       -1 * time.Second,
		},
	} {
		t.Run(test.description, func(t *testing.T) {
			if _, err := NewAutomation(test.ap, test.ha, test.pt); err == nil {
				t.Errorf("NewAutomation expected an error on init. Got none.")
			}
		})
	}
}
