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
)

func TestTimeToClosestPrayer(t *testing.T) {
	parse := func(s string) time.Time {
		c, err := time.Parse("15:04", s)
		if err != nil {
			t.Fatalf("Failed to parse the time %v: %v", s, err)
		}
		return c
	}

	prayerTimes := &prayerTimes{
		Fajr:    &prayer{time: parse("09:00")},
		Dhuhr:   &prayer{time: parse("12:00")},
		Asr:     &prayer{time: parse("15:00")},
		Maghrib: &prayer{time: parse("18:00")},
		Ishaa:   &prayer{time: parse("21:00")},
	}

	for _, test := range []struct {
		description string
		clock       string
		wantPrev    time.Duration
		wantNext    time.Duration
	}{
		{
			description: "40 minutes before Fajr",
			clock:       "08:20",
			// Ishaa - 23 hours.
			wantPrev: time.Hour*10 + time.Minute*20,
			wantNext: time.Minute * 40,
		},
		{
			description: "30 minutes after Fajr",
			clock:       "09:30",
			wantPrev:    time.Minute * 30,
			wantNext:    time.Hour*2 + time.Minute*30,
		},
		{
			description: "30 minutes before Dhuhr",
			clock:       "11:30",
			wantPrev:    time.Hour*2 + time.Minute*30,
			wantNext:    time.Minute * 30,
		},
		{
			description: "30 minutes after Dhuhr",
			clock:       "12:30",
			wantPrev:    time.Minute * 30,
			wantNext:    time.Hour*2 + time.Minute*30,
		},
		{
			description: "30 minutes before Asr",
			clock:       "14:30",
			wantPrev:    time.Hour*2 + time.Minute*30,
			wantNext:    time.Minute * 30,
		},
		{
			description: "30 minutes after Asr",
			clock:       "15:30",
			wantPrev:    time.Minute * 30,
			wantNext:    time.Hour*2 + time.Minute*30,
		},
		{
			description: "30 minutes before Maghrib",
			clock:       "17:30",
			wantPrev:    time.Hour*2 + time.Minute*30,
			wantNext:    time.Minute * 30,
		},
		{
			description: "30 minutes after Maghrib",
			clock:       "18:30",
			wantPrev:    time.Minute * 30,
			wantNext:    time.Hour*2 + time.Minute*30,
		},
		{
			description: "30 minutes before Ishaa",
			clock:       "20:30",
			wantPrev:    time.Hour*2 + time.Minute*30,
			wantNext:    time.Minute * 30,
		},
		{
			description: "30 minutes after Ishaa",
			clock:       "21:30",
			wantPrev:    time.Minute * 30,
			// Fajr + 23 hours.
			wantNext: time.Hour*10 + time.Minute*30,
		},
		{
			description: "Dhuhr Time",
			clock:       "12:00",
			wantPrev:    time.Hour * 3,
			wantNext:    0,
		},
	} {
		t.Run(test.description, func(t *testing.T) {
			now := parse(test.clock)
			prev, next, err := prayerTimes.GetNearestPrayers(now)
			if err != nil {
				t.Fatalf("TimeToClosestPrayer returned error, expected None: %v", err)
			}
			prevTime := prev.TimeToPrayer(now)
			nextTime := next.TimeToPrayer(now)

			if test.wantPrev != prevTime {
				t.Errorf("Time to previous prayer mismatch. Want %v got %v", test.wantPrev, prevTime)
			}
			if test.wantNext != nextTime {
				t.Errorf("Time to next prayer mismatch. Want %v got %v", test.wantNext, nextTime)
			}
		})
	}
}
