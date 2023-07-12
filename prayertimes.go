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

// handles prayer times geographicacl calculations and times till next prayers
// calculations.

package main

import (
	"errors"
	"time"
)

type prayerTimes struct {
	Fajr    time.Time
	Dhuhr   time.Time
	Asr     time.Time
	Maghrib time.Time
	Ishaa   time.Time
}

func NewPrayerTimes() (*prayerTimes, error) {
	return nil, errors.New("Unimplemented.")
}

func (p *prayerTimes) TimeToNextPrayer(ts time.Time) (time.Duration, error) {
	return 0, errors.New("Unimplemented.")
}

func (p *prayerTimes) TimeFromLastPrayer(ts time.Time) (time.Duration, error) {
	return 0, errors.New("Unimplemented.")
}
