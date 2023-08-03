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
	"fmt"
	"log"
	"time"
)

type prayer struct {
	name string
	time time.Time
}

// returns duration from now till an input prayer.
func (p *prayer) TimeToPrayer(now time.Time) time.Duration {
	if now.Compare(p.time) > 0 {
		// prayer has passed
		return now.Sub(p.time)
	}
	// prayer is coming
	return p.time.Sub(now)
}

// IPrayerTimes is an interface to be used by specific cities or a global
// prayer time calculator.
type IPrayerTimes interface {
	GetTodayPrayerTimes(now time.Time) error
	GetNearestPrayers(now time.Time) (*prayer, *prayer, error)
}

// prayerTimes contains all 5 prayers and the date (yyyy-mm-dd) for caching.
type prayerTimes struct {
	Fajr    *prayer
	Dhuhr   *prayer
	Asr     *prayer
	Maghrib *prayer
	Ishaa   *prayer

	date time.Time
}

// GetNearestPrayers returns the previous and next *prayers given a timestamp.
func (p *prayerTimes) GetNearestPrayers(now time.Time) (*prayer, *prayer, error) {
	switch {
	case now.Equal(p.Fajr.time) || now.Before(p.Fajr.time):
		// TODO(ssafty): use exact prayer times from previous day.
		ishaa_last_day := &prayer{
			name: "Ishaa",
			time: p.Ishaa.time.Add(-23 * time.Hour)}
		return ishaa_last_day, p.Fajr, nil
	case isBetweenPrayers(p.Fajr.time, now, p.Dhuhr.time):
		return p.Fajr, p.Dhuhr, nil
	case isBetweenPrayers(p.Dhuhr.time, now, p.Asr.time):
		return p.Dhuhr, p.Asr, nil
	case isBetweenPrayers(p.Asr.time, now, p.Maghrib.time):
		return p.Asr, p.Maghrib, nil
	case isBetweenPrayers(p.Maghrib.time, now, p.Ishaa.time):
		return p.Maghrib, p.Ishaa, nil
	case now.After(p.Ishaa.time):
		// TODO(ssafty): use exact prayer times from next day.
		fajr_next_day := &prayer{
			name: "Fajr",
			time: p.Fajr.time.Add(23 * time.Hour)}
		return p.Ishaa, fajr_next_day, nil
	default:
		return nil, nil, fmt.Errorf("Failed to find Time to closest prayer for timestamp: %v and prayerTimes: %v", now, p)
	}
}

// isSameDay returns True if all input timestamps have the same date.
func isSameDay(tss ...time.Time) bool {
	if len(tss) < 2 {
		return true
	}

	y, m, d := tss[0].Date()
	for _, ts := range tss {
		yy, mm, dd := ts.Date()
		if yy != y || mm != m || dd != d {
			return false
		}
	}
	return true
}

// isBetweenPrayers returns True if input timestamp is within prayerA and prayerB
func isBetweenPrayers(prayerA, ts time.Time, prayerB time.Time) bool {
	return ts.After(prayerA) && ts.Before(prayerB) || ts.Equal(prayerB)
}

// GetDate returns Date (yyyy-mm-dd) of a given timestamp.
func GetDate(n time.Time) time.Time {
	return time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, n.Location())
}

// Munich.

// munichPrayerTimes extends prayerTimes to reuse GetNearestPrayers.
type munichPrayerTimes struct {
	prayerTimes
}

func NewMunichPrayerTimes() (*munichPrayerTimes, error) {
	pt := &munichPrayerTimes{}
	if err := pt.GetTodayPrayerTimes(time.Now()); err != nil {
		return nil, fmt.Errorf("Error initializing NewPrayerTimes for %s: %w", time.Now().Format("2006-01-02"), err)
	}
	return pt, nil
}

// GetTodayPrayerTimes is specific to Munich as it reads from static prayer times.
func (p *munichPrayerTimes) GetTodayPrayerTimes(now time.Time) error {
	if !p.date.IsZero() && p.date == GetDate(now) {
		return nil
	}

	pts := []time.Time{}
	for i := 0; i < 6; i++ {
		t := munich2023[now.Month()-1][6*(now.Day()-1)+i]
		parsed, err := time.Parse("15:04", t)
		if err != nil {
			return fmt.Errorf("Error parsing prayertime %v: %w", t, err)
		}

		tt := time.Date(now.Year(), now.Month(), now.Day(), parsed.Hour(), parsed.Minute(), 0, 0, now.Location())

		if i == 1 {
			// We don't consider Ishraq.
			continue
		}
		pts = append(pts, tt)
	}

	p.Fajr = &prayer{name: "Fajr", time: pts[0]}
	p.Dhuhr = &prayer{name: "Dhuhr", time: pts[1]}
	p.Asr = &prayer{name: "Asr", time: pts[2]}
	p.Maghrib = &prayer{name: "Maghrib", time: pts[3]}
	p.Ishaa = &prayer{name: "Ishaa", time: pts[4]}
	p.date = GetDate(now)

	if !isSameDay(now, p.Fajr.time, p.Dhuhr.time, p.Asr.time, p.Maghrib.time, p.Ishaa.time) {
		return fmt.Errorf("Failed to find time to closest prayer. Found Inconsistency of dates between now (%v) and today's prayers(%v)", now, p)
	}

	log.Printf("PrayerTimes today: %v", *p)
	return nil
}

// source: https://www.islamisches-zentrum-muenchen.de/
var munich2023 = [][]string{
	{"06:10", "07:59", "12:22", "14:15", "16:35", "18:17", "06:11", "07:59", "12:23", "14:15", "16:36", "18:18", "06:11", "07:59", "12:23", "14:16", "16:37", "18:19", "06:11", "07:59", "12:23", "14:17", "16:38", "18:20", "06:11", "07:59", "12:24", "14:18", "16:39", "18:21", "06:11", "07:59", "12:24", "14:19", "16:40", "18:22", "06:10", "07:58", "12:25", "14:20", "16:41", "18:23", "06:10", "07:58", "12:25", "14:21", "16:42", "18:24", "06:10", "07:58", "12:26", "14:22", "16:44", "18:25", "06:10", "07:57", "12:26", "14:23", "16:45", "18:26", "06:10", "07:57", "12:26", "14:24", "16:46", "18:27", "06:09", "07:56", "12:27", "14:26", "16:47", "18:28", "06:09", "07:56", "12:27", "14:27", "16:49", "18:29", "06:09", "07:55", "12:28", "14:28", "16:50", "18:30", "06:08", "07:54", "12:28", "14:29", "16:51", "18:32", "06:08", "07:54", "12:28", "14:30", "16:53", "18:33", "06:07", "07:53", "12:29", "14:31", "16:54", "18:34", "06:07", "07:52", "12:29", "14:33", "16:56", "18:35", "06:06", "07:51", "12:29", "14:34", "16:57", "18:36", "06:05", "07:51", "12:30", "14:35", "16:59", "18:38", "06:05", "07:50", "12:30", "14:36", "17:00", "18:39", "06:04", "07:49", "12:30", "14:38", "17:02", "18:40", "06:03", "07:48", "12:30", "14:39", "17:03", "18:41", "06:03", "07:47", "12:31", "14:40", "17:05", "18:43", "06:02", "07:46", "12:31", "14:41", "17:06", "18:44", "06:01", "07:45", "12:31", "14:43", "17:08", "18:45", "06:00", "07:44", "12:31", "14:44", "17:09", "18:47", "05:59", "07:42", "12:32", "14:45", "17:11", "18:48", "05:58", "07:41", "12:32", "14:47", "17:12", "18:49", "05:57", "07:40", "12:32", "14:48", "17:14", "18:51", "05:56", "07:39", "12:32", "14:49", "17:16", "18:52"},
	{"05:55", "07:37", "12:32", "14:51", "17:17", "18:53", "05:54", "07:36", "12:32", "14:52", "17:19", "18:55", "05:53", "07:35", "12:32", "14:53", "17:20", "18:56", "05:52", "07:33", "12:33", "14:54", "17:22", "18:58", "05:50", "07:32", "12:33", "14:56", "17:24", "18:59", "05:49", "07:30", "12:33", "14:57", "17:25", "19:00", "05:48", "07:29", "12:33", "14:58", "17:27", "19:02", "05:47", "07:27", "12:33", "15:00", "17:28", "19:03", "05:45", "07:26", "12:33", "15:01", "17:30", "19:05", "05:44", "07:24", "12:33", "15:02", "17:32", "19:06", "05:42", "07:23", "12:33", "15:03", "17:33", "19:07", "05:41", "07:21", "12:33", "15:05", "17:35", "19:09", "05:40", "07:19", "12:33", "15:06", "17:36", "19:10", "05:38", "07:18", "12:33", "15:07", "17:38", "19:12", "05:37", "07:16", "12:33", "15:08", "17:40", "19:13", "05:35", "07:14", "12:33", "15:10", "17:41", "19:15", "05:33", "07:13", "12:33", "15:11", "17:43", "19:16", "05:32", "07:11", "12:33", "15:12", "17:44", "19:18", "05:30", "07:09", "12:33", "15:13", "17:46", "19:19", "05:28", "07:07", "12:32", "15:15", "17:48", "19:21", "05:27", "07:06", "12:32", "15:16", "17:49", "19:22", "05:25", "07:04", "12:32", "15:17", "17:51", "19:23", "05:23", "07:02", "12:32", "15:18", "17:52", "19:25", "05:22", "07:00", "12:32", "15:19", "17:54", "19:26", "05:20", "06:58", "12:32", "15:20", "17:55", "19:28", "05:18", "06:56", "12:32", "15:21", "17:57", "19:29", "05:16", "06:54", "12:32", "15:23", "17:59", "19:31", "05:14", "06:53", "12:31", "15:24", "18:00", "19:32"},
	{"05:12", "06:51", "12:31", "15:25", "18:02", "19:34", "05:10", "06:49", "12:31", "15:26", "18:03", "19:36", "05:08", "06:47", "12:31", "15:27", "18:05", "19:37", "05:06", "06:45", "12:31", "15:28", "18:06", "19:39", "05:04", "06:43", "12:30", "15:29", "18:08", "19:40", "05:02", "06:41", "12:30", "15:30", "18:09", "19:42", "05:00", "06:39", "12:30", "15:31", "18:11", "19:43", "04:58", "06:37", "12:30", "15:32", "18:12", "19:45", "04:56", "06:35", "12:29", "15:33", "18:14", "19:46", "04:54", "06:33", "12:29", "15:34", "18:15", "19:48", "04:52", "06:31", "12:29", "15:35", "18:17", "19:49", "04:50", "06:29", "12:29", "15:36", "18:18", "19:51", "04:48", "06:27", "12:28", "15:37", "18:20", "19:53", "04:46", "06:25", "12:28", "15:38", "18:21", "19:54", "04:43", "06:23", "12:28", "15:39", "18:23", "19:56", "04:41", "06:21", "12:28", "15:40", "18:24", "19:57", "04:39", "06:19", "12:27", "15:41", "18:26", "19:59", "04:37", "06:17", "12:27", "15:42", "18:27", "20:01", "04:35", "06:15", "12:27", "15:42", "18:29", "20:02", "04:32", "06:12", "12:26", "15:43", "18:30", "20:04", "04:30", "06:10", "12:26", "15:44", "18:32", "20:06", "04:28", "06:08", "12:26", "15:45", "18:33", "20:07", "04:25", "06:06", "12:25", "15:46", "18:35", "20:09", "04:23", "06:04", "12:25", "15:47", "18:36", "20:11", "04:21", "06:02", "12:25", "15:48", "18:38", "20:12", "05:18", "07:00", "13:25", "16:48", "19:39", "21:18", "05:16", "06:58", "13:24", "16:49", "19:40", "21:16", "05:14", "06:56", "13:24", "16:50", "19:42", "21:18", "05:11", "06:54", "13:24", "16:51", "19:43", "21:19", "05:09", "06:52", "13:23", "16:51", "19:45", "21:21", "05:06", "06:50", "13:23", "16:52", "19:46", "21:23"},
	{"05:04", "06:48", "13:23", "16:53", "19:48", "21:25", "05:01", "06:46", "13:22", "16:54", "19:49", "21:27", "04:59", "06:44", "13:22", "16:54", "19:51", "21:28", "04:57", "06:42", "13:22", "16:55", "19:52", "21:30", "04:54", "06:40", "13:22", "16:56", "19:54", "21:32", "04:52", "06:38", "13:21", "16:57", "19:55", "21:34", "04:49", "06:36", "13:21", "16:57", "19:56", "21:36", "04:47", "06:34", "13:21", "16:58", "19:58", "21:38", "04:44", "06:32", "13:20", "16:59", "19:59", "21:40", "04:41", "06:30", "13:20", "16:59", "20:01", "21:42", "04:39", "06:28", "13:20", "17:00", "20:02", "21:44", "04:36", "06:26", "13:20", "17:01", "20:04", "21:46", "04:34", "06:24", "13:19", "17:01", "20:05", "21:48", "04:31", "06:22", "13:19", "17:02", "20:07", "21:50", "04:29", "06:20", "13:19", "17:03", "20:08", "21:52", "04:26", "06:18", "13:19", "17:03", "20:09", "21:54", "04:23", "06:16", "13:18", "17:04", "20:11", "21:56", "04:21", "06:14", "13:18", "17:04", "20:12", "21:58", "04:18", "06:12", "13:18", "17:05", "20:14", "22:00", "04:16", "06:10", "13:18", "17:06", "20:15", "22:02", "04:13", "06:08", "13:18", "17:06", "20:17", "22:04", "04:10", "06:07", "13:17", "17:07", "20:18", "22:06", "04:08", "06:05", "13:17", "17:07", "20:20", "22:08", "04:05", "06:03", "13:17", "17:08", "20:21", "22:11", "04:02", "06:01", "13:17", "17:09", "20:22", "22:13", "04:00", "05:59", "13:17", "17:09", "20:24", "22:15", "03:57", "05:58", "13:16", "17:10", "20:25", "22:17", "03:54", "05:56", "13:16", "17:10", "20:27", "22:19", "03:52", "05:54", "13:16", "17:11", "20:28", "22:22", "03:49", "05:52", "13:16", "17:11", "20:30", "22:24"},
	{"03:46", "05:51", "13:16", "17:12", "20:31", "22:26", "03:44", "05:49", "13:16", "17:12", "20:32", "22:29", "03:41", "05:48", "13:16", "17:13", "20:34", "22:31", "03:38", "05:46", "13:16", "17:13", "20:35", "22:33", "03:37", "05:44", "13:15", "17:14", "20:37", "22:36", "03:37", "05:43", "13:15", "17:15", "20:38", "22:38", "03:37", "05:41", "13:15", "17:15", "20:39", "22:40", "03:37", "05:40", "13:15", "17:16", "20:41", "22:43", "03:37", "05:38", "13:15", "17:16", "20:42", "22:45", "03:37", "05:37", "13:15", "17:17", "20:44", "22:48", "03:37", "05:35", "13:15", "17:17", "20:45", "22:48", "03:37", "05:34", "13:15", "17:18", "20:46", "22:48", "03:37", "05:33", "13:15", "17:18", "20:48", "22:48", "03:37", "05:31", "13:15", "17:19", "20:49", "22:48", "03:37", "05:30", "13:15", "17:19", "20:50", "22:48", "03:37", "05:29", "13:15", "17:20", "20:52", "22:48", "03:37", "05:27", "13:15", "17:20", "20:53", "22:48", "03:37", "05:26", "13:15", "17:21", "20:54", "22:48", "03:37", "05:25", "13:15", "17:21", "20:55", "22:48", "03:37", "05:24", "13:15", "17:21", "20:57", "22:48", "03:37", "05:23", "13:15", "17:22", "20:58", "22:48", "03:37", "05:22", "13:15", "17:22", "20:59", "22:48", "03:37", "05:21", "13:15", "17:23", "21:00", "22:48", "03:37", "05:20", "13:16", "17:23", "21:01", "22:48", "03:37", "05:19", "13:16", "17:24", "21:03", "22:48", "03:37", "05:18", "13:16", "17:24", "21:04", "22:48", "03:37", "05:17", "13:16", "17:25", "21:05", "22:48", "03:37", "05:16", "13:16", "17:25", "21:06", "22:48", "03:37", "05:15", "13:16", "17:25", "21:07", "22:48", "03:37", "05:15", "13:16", "17:26", "21:08", "22:48", "03:37", "05:14", "13:16", "17:26", "21:09", "22:48"},
	{"03:37", "05:13", "13:17", "17:27", "21:10", "22:48", "03:37", "05:12", "13:17", "17:27", "21:11", "22:48", "03:37", "05:12", "13:17", "17:28", "21:12", "22:48", "03:37", "05:11", "13:17", "17:28", "21:13", "22:48", "03:37", "05:11", "13:17", "17:28", "21:14", "22:48", "03:37", "05:10", "13:17", "17:29", "21:14", "22:48", "03:37", "05:10", "13:18", "17:29", "21:15", "22:48", "03:37", "05:09", "13:18", "17:29", "21:16", "22:48", "03:37", "05:09", "13:18", "17:30", "21:17", "22:48", "03:37", "05:09", "13:18", "17:30", "21:17", "22:48", "03:37", "05:09", "13:18", "17:30", "21:18", "22:48", "03:47", "05:08", "13:19", "17:31", "21:19", "22:48", "03:47", "05:08", "13:19", "17:31", "21:19", "22:48", "03:37", "05:08", "13:19", "17:31", "21:20", "22:48", "03:37", "05:08", "13:19", "17:32", "21:20", "22:48", "03:37", "05:08", "13:19", "17:32", "21:21", "22:48", "03:37", "05:08", "13:20", "17:32", "21:21", "22:48", "03:37", "05:08", "13:20", "17:33", "21:22", "22:48", "03:37", "05:08", "13:20", "17:33", "21:22", "22:48", "03:37", "05:08", "13:20", "17:33", "21:22", "22:48", "03:37", "05:08", "13:20", "17:33", "21:22", "22:48", "03:37", "05:09", "13:21", "17:33", "21:23", "22:48", "03:38", "05:09", "13:21", "17:34", "21:23", "22:48", "03:38", "05:09", "13:21", "17:34", "21:23", "22:48", "03:38", "05:10", "13:21", "17:34", "21:23", "22:48", "03:38", "05:10", "13:22", "17:34", "21:23", "22:48", "03:38", "05:10", "13:22", "17:34", "21:23", "22:48", "03:38", "05:11", "13:22", "17:34", "21:23", "22:48", "03:38", "05:11", "13:22", "17:35", "21:23", "22:48", "03:38", "05:12", "13:22", "17:35", "21:23", "22:48"},
	{"03:38", "05:12", "13:23", "17:35", "21:23", "22:48", "03:38", "05:13", "13:23", "17:35", "21:22", "22:48", "03:38", "05:14", "13:23", "17:35", "21:22", "22:48", "03:38", "05:14", "13:23", "17:35", "21:22", "22:48", "03:38", "05:15", "13:23", "17:35", "21:21", "22:48", "03:38", "05:16", "13:23", "17:35", "21:21", "22:48", "03:38", "05:17", "13:24", "17:35", "21:21", "22:48", "03:38", "05:18", "13:24", "17:35", "21:20", "22:48", "03:38", "05:18", "13:24", "17:35", "21:19", "22:48", "03:38", "05:19", "13:24", "17:35", "21:19", "22:48", "03:38", "05:20", "13:24", "17:35", "21:18", "22:48", "03:38", "05:21", "13:24", "17:35", "21:18", "22:48", "03:38", "05:22", "13:24", "17:34", "21:17", "22:48", "03:38", "05:23", "13:25", "17:34", "21:16", "22:48", "03:38", "05:24", "13:25", "17:34", "21:15", "22:48", "03:38", "05:25", "13:25", "17:34", "21:15", "22:48", "03:38", "05:26", "13:25", "17:34", "21:14", "22:48", "03:38", "05:27", "13:25", "17:33", "21:13", "22:48", "03:38", "05:28", "13:25", "17:33", "21:12", "22:48", "03:38", "05:29", "13:25", "17:33", "21:11", "22:48", "03:38", "05:31", "13:25", "17:32", "21:10", "22:48", "03:38", "05:32", "13:25", "17:32", "21:09", "22:48", "03:38", "05:33", "13:25", "17:32", "21:08", "22:48", "03:38", "05:34", "13:25", "17:31", "21:06", "22:48", "03:38", "05:35", "13:25", "17:31", "21:05", "22:48", "03:38", "05:36", "13:25", "17:31", "21:04", "22:48", "03:38", "05:38", "13:25", "17:30", "21:03", "22:48", "03:38", "05:39", "13:25", "17:30", "21:02", "22:48", "03:38", "05:40", "13:25", "17:29", "21:00", "22:48", "03:38", "05:41", "13:25", "17:29", "20:59", "22:48", "03:38", "05:43", "13:25", "17:28", "20:58", "22:48"},
	{"03:38", "05:44", "13:25", "17:28", "20:56", "22:48", "03:38", "05:45", "13:25", "17:27", "20:55", "22:48", "03:38", "05:47", "13:25", "17:26", "20:53", "22:48", "03:38", "05:48", "13:25", "17:26", "20:52", "22:48", "03:38", "05:49", "13:25", "17:25", "20:50", "22:48", "03:40", "05:51", "13:25", "17:24", "20:49", "22:48", "03:42", "05:52", "13:25", "17:24", "20:47", "22:47", "03:44", "05:53", "13:24", "17:23", "20:46", "22:45", "03:47", "05:55", "13:24", "17:22", "20:44", "22:42", "03:49", "05:56", "13:24", "17:21", "20:42", "22:40", "03:52", "05:57", "13:24", "17:21", "20:41", "22:37", "03:54", "05:59", "13:24", "17:20", "20:39", "22:34", "03:56", "06:00", "13:24", "17:19", "20:37", "22:32", "03:59", "06:01", "13:24", "17:18", "20:36", "22:29", "04:01", "06:03", "13:23", "17:17", "20:34", "22:27", "04:03", "06:04", "13:23", "17:16", "20:32", "22:24", "04:06", "06:05", "13:23", "17:15", "20:30", "22:22", "04:08", "06:07", "13:23", "17:15", "20:29", "22:19", "04:10", "06:08", "13:22", "17:14", "20:27", "22:16", "04:12", "06:10", "13:22", "17:13", "20:25", "22:14", "04:15", "06:11", "13:22", "17:12", "20:2", "22:11", "04:17", "06:12", "13:22", "17:11", "20:21", "22:09", "04:19", "06:14", "13:22", "17:10", "20:19", "22:06", "04:21", "06:15", "13:21", "17:08", "20:17", "22:04", "04:23", "06:16", "13:21", "17:07", "20:16", "22:01", "04:25", "06:18", "13:21", "17:06", "20:14", "21:59", "04:27", "06:19", "13:20", "17:05", "20:12", "21:56", "04:29", "06:21", "13:20", "17:04", "20:10", "21:53", "04:31", "06:22", "13:20", "17:03", "20:08", "21:51", "04:33", "06:23", "13:20", "17:02", "20:06", "21:48", "04:35", "06:25", "13:19", "17:00", "20:04", "21:46"},
	{"04:37", "06:26", "13:19", "16:59", "20:02", "21:48", "04:39", "06:27", "13:19", "16:58", "20:00", "21:41", "04:41", "06:29", "13:18", "16:57", "19:58", "21:38", "04:43", "06:30", "13:18", "16:56", "19:56", "21:36", "04:45", "06:32", "13:18", "16:56", "19:54", "21:33", "04:47", "06:33", "13:17", "16:53", "19:52", "21:31", "04:48", "06:34", "13:17", "16:52", "19:50", "21:28", "04:50", "06:36", "13:17", "16:50", "19:48", "21:26", "04:52", "06:37", "13:16", "16:49", "19:46", "21:24", "04:54", "06:38", "13:16", "16:48", "19:43", "21:21", "04:56", "06:40", "13:16", "16:46", "19:41", "21:19", "04:57", "06:41", "13:15", "16:45", "19:39", "21:16", "04:59", "06:42", "13:15", "16:44", "19:37", "21:14", "05:01", "06:44", "13:14", "16:42", "19:35", "21:11", "05:03", "06:45", "13:14", "16:41", "19:33", "21:09", "05:04", "06:47", "13:14", "16:39", "19:31", "21:07", "05:06", "06:48", "13:13", "16:38", "19:29", "21:04", "05:08", "06:49", "13:13", "16:36", "19:27", "21:02", "05:09", "06:51", "13:13", "16:35", "19:25", "21:00", "05:11", "06:52", "13:12", "16:34", "19:23", "20:57", "05:13", "06:53", "13:12", "16:32", "19:21", "20:55", "05:14", "06:55", "13:12", "16:31", "19:18", "20:53", "05:16", "06:56", "13:11", "16:29", "19:16", "20:50", "05:17", "06:58", "13:11", "16:28", "19:14", "20:48", "05:19", "06:59", "13:11", "16:26", "19:12", "20:46", "05:21", "07:00", "13:10", "16:25", "19:10", "20:44", "05:22", "07:02", "13:10", "16:23", "19:08", "20:41", "05:24", "07:03", "13:10", "16:22", "19:06", "20:39", "05:25", "07:05", "13:09", "16:20", "19:04", "20:37", "05:27", "07:06", "13:09", "16:19", "19:02", "20:35"},
	{"05:28", "07:07", "13:09", "16:17", "19:00", "20:32", "05:30", "07:09", "13:08", "16:15", "18:58", "20:30", "05:31", "07:10", "13:08", "16:14", "18:56", "20:28", "05:33", "07:12", "13:08", "16:12", "18:54", "20:26", "05:34", "07:13", "13:07", "16:11", "18:52", "20:24", "05:36", "07:15", "13:07", "16:09", "18:49", "20:22", "05:37", "07:16", "13:07", "16:08", "18:47", "20:20", "05:39", "07:17", "13:06", "16:06", "18:45", "20:18", "05:40", "07:19", "13:06", "16:05", "18:43", "20:16", "05:42", "07:20", "13:06", "16:03", "18:41", "20:14", "05:43", "07:22", "13:06", "16:02", "18:39", "20:12", "05:45", "07:23", "13:05", "16:00", "18:38", "20:10", "05:46", "07:25", "13:05", "15:59", "18:36", "20:08", "05:48", "07:26", "13:05", "15:57", "18:34", "20:06", "05:49", "07:28", "13:05", "15:56", "18:32", "20:04", "05:51", "07:29", "13:04", "15:54", "18:30", "20:02", "05:52", "07:31", "13:04", "15:53", "18:28", "20:00", "05:53", "07:32", "13:04", "15:51", "18:26", "19:59", "05:55", "07:34", "13:04", "15:50", "18:24", "19:57", "05:56", "07:35", "13:04", "15:48", "18:22", "19:55", "05:58", "07:37", "13:03", "15:47", "18:20", "19:53", "05:59", "07:38", "13:03", "15:45", "18:19", "19:51", "06:01", "07:40", "13:03", "15:44", "18:17", "19:50", "06:02", "07:41", "13:03", "15:43", "18:15", "19:48", "06:03", "07:43", "13:03", "15:41", "18:13", "19:46", "06:05", "07:44", "13:03", "15:40", "18:11", "19:45", "06:06", "07:46", "13:03", "15:38", "18:10", "19:43", "06:07", "07:47", "13:03", "15:37", "18:08", "19:42", "06:09", "07:49", "13:02", "15:36", "18:06", "19:40", "05:10", "06:50", "12:02", "14:34", "17:05", "18:39", "05:12", "06:52", "12:00", "14:33", "17:03", "18:37"},
	{"05:13", "06:53", "12:02", "14:32", "17:01", "18:36", "05:14", "06:55", "12:02", "14:31", "17:00", "18:34", "05:16", "06:56", "12:02", "14:29", "16:58", "18:33", "05:17", "06:58", "12:02", "14:28", "16:57", "18:32", "05:18", "07:00", "12:02", "14:27", "16:55", "18:30", "05:20", "07:01", "12:02", "14:26", "16:54", "18:29", "05:21", "07:03", "12:02", "14:25", "16:52", "18:28", "05:22", "07:04", "12:02", "14:23", "16:51", "18:26", "05:24", "07:06", "12:03", "14:22", "16:49", "18:25", "05:25", "07:07", "12:03", "14:21", "16:48", "18:24", "05:26", "07:09", "12:03", "14:20", "16:47", "18:23", "05:28", "07:10", "12:03", "14:19", "16:45", "18:22", "05:29", "07:12", "12:03", "14:18", "16:44", "18:21", "05:30", "07:13", "12:03", "14:17", "16:43", "18:20", "05:32", "07:15", "12:03", "14:16", "16:42", "18:19", "05:33", "07:16", "12:03", "14:15", "16:40", "18:18", "05:34", "07:18", "12:04", "14:14", "16:39", "18:17", "05:35", "07:19", "12:04", "14:14", "16:38", "18:16", "05:37", "07:21", "12:04", "14:13", "16:37", "18:15", "05:38", "07:22", "12:04", "14:12", "16:36", "18:15", "05:39", "07:24", "12:04", "14:11", "16:35", "18:14", "05:40", "07:25", "12:05", "14:11", "16:34", "18:13", "05:41", "07:27", "12:05", "14:10", "16:33", "18:12", "05:43", "07:28", "12:05", "14:09", "16:32", "18:12", "05:44", "07:29", "12:06", "14:09", "16:32", "18:11", "05:45", "07:31", "12:06", "14:08", "16:31", "18:11", "05:46", "07:32", "12:06", "14:08", "16:30", "18:10", "05:47", "07:34", "12:07", "14:07", "16:29", "18:10", "05:48", "07:35", "12:07", "14:07", "16:29", "18:09", "05:49", "07:36", "12:07", "14:06", "16:28", "18:09"},
	{"05:50", "07:37", "12:08", "14:06", "16:28", "18:09", "05:51", "07:39", "12:08", "14:06", "16:27", "18:08", "0:52", "07:40", "12:08", "14:05", "16:27", "18:08", "05:53", "07:41", "12:08", "14:05", "16:26", "18:08", "05:54", "07:42", "12:09", "14:05", "16:26", "18:08", "05:55", "07:43", "12:10", "14:05", "16:26", "18:07", "05:56", "07:45", "12:10", "14:05", "16:25", "18:07", "05:57", "07:46", "12:10", "14:04", "16:25", "18:07", "05:58", "07:47", "12:11", "14:04", "16:25", "18:07", "05:59", "07:48", "12:11", "14:04", "16:25", "18:07", "06:00", "07:49", "12:12", "14:04", "16:25", "18:07", "06:01", "07:50", "12:12", "14:05", "16:25", "18:07", "06:02", "07:50", "12:13", "14:05", "16:25", "18:08", "06:02", "07:51", "12:13", "14:05", "16:25", "18:08", "06:03", "07:52", "12:14", "14:05", "16:25", "18:08", "06:04", "07:53", "12:14", "14:05", "16:25", "18:08", "06:04", "07:54", "12:15", "14:06", "16:26", "18:09", "06:05", "07:54", "12:15", "14:06", "16:26", "18:09", "06:06", "07:55", "12:16", "14:06", "16:26", "18:09", "06:06", "07:56", "12:16", "14:07", "16:27", "18:10", "06:07", "07:56", "12:17", "14:07", "16:27", "18:10", "06:07", "07:57", "12:17", "14:08", "16:27", "18:11", "06:08", "07:57", "12:18", "14:08", "16:28", "18:11", "06:08", "07:58", "12:18", "14:09", "16:29", "18:12", "06:09", "07:58", "12:19", "14:09", "16:29", "18:12", "06:09", "07:58", "12:19", "14:10", "16:30", "18:13", "06:09", "07:59", "12:20", "14:11", "16:31", "18:14", "06:10", "07:59", "12:20", "14:11", "16:31", "18:14", "06:10", "07:59", "12:21", "14:12", "16:32", "18:15", "06:10", "07:59", "12:21", "14:13", "16:33", "18:16", "06:10", "07:59", "12:21", "14:14", "16:34", "18:16"},
}
