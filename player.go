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

// MP3 files handler. Reads the Adhan mp3 file and Rewinds/Plays it.

package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/hajimehoshi/go-mp3"
	"github.com/hajimehoshi/oto/v2"
)

type adhanPlayer struct {
	player oto.Player

	filePath      string
	samplingRate  *int
	numChannels   *int
	audioBitDepth *int
}

type adhanPlayerOpt func(*adhanPlayer)

func FilePath(f string) adhanPlayerOpt {
	return func(a *adhanPlayer) {
		a.filePath = f
	}
}

func SamplingRate(r int) adhanPlayerOpt {
	return func(a *adhanPlayer) {
		a.samplingRate = &r
	}
}

func NumChannels(n int) adhanPlayerOpt {
	return func(a *adhanPlayer) {
		a.numChannels = &n
	}
}

func AudioBitDepth(d int) adhanPlayerOpt {
	return func(a *adhanPlayer) {
		a.audioBitDepth = &d
	}
}

func NewAdhanPlayer(opts ...adhanPlayerOpt) (*adhanPlayer, error) {
	ap := &adhanPlayer{}

	for _, opt := range opts {
		opt(ap)
	}

	switch {
	case ap.filePath == "":
		return nil, errors.New("NewAdhanPlayer's file path is not specified.")
	case ap.audioBitDepth == nil:
		return nil, errors.New("NewAdhanPlayer's AudioBitDepth is not specified.")
	case ap.numChannels == nil:
		return nil, errors.New("NewAdhanPlayer's NumChannels is not specified")
	case ap.samplingRate == nil:
		return nil, errors.New("NewAdhanPlayer's audioBitDepth is not specified")
	}

	// Load file into memory.
	audioBytes, err := os.ReadFile(ap.filePath)
	if err != nil {
		return nil, fmt.Errorf("NewAdhanPlayer reading %s to bytes failed: %w", ap.filePath, err)
	}

	decoded, err := mp3.NewDecoder(bytes.NewReader(audioBytes))
	if err != nil {
		return nil, fmt.Errorf("NewAdhanPlayer decoding Audio Byte failed: %w", err)
	}

	otoCtx, readyChan, err := oto.NewContext(*ap.samplingRate, *ap.numChannels, *ap.audioBitDepth)
	if err != nil {
		return nil, fmt.Errorf("NewAdhanPlayer oto.NewContext creation failed: %w", err)
	}

	<-readyChan

	ap.player = otoCtx.NewPlayer(decoded)
	return ap, nil
}

func (a *adhanPlayer) Play() error {
	_, err := a.player.(io.Seeker).Seek(0, io.SeekStart)
	if err != nil {
		return fmt.Errorf("AdhanPlayer rewind failed: %w", err)
	}
	log.Println("Playing Adhan.")
	a.player.Play()
	return nil
}

func (a *adhanPlayer) IsPlaying() bool {
	if ip := a.player.IsPlaying(); ip {
		log.Println("AdhanPlayer is currently playing.")
		return true
	}
	return false
}

func (a *adhanPlayer) Stop() error {
	err := a.player.Close()
	if err != nil {
		return fmt.Errorf("AdhanPlayer closing failed: %w", err)
	}
	return nil
}
