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

// MP3 files handler. Plays the adhan and rewinding it to prepare it for
// the next prayer.

package main

<<<<<<< HEAD:azanapp.go
func main() {}
=======
import "errors"

type adhanPlayer struct{}

func NewAdhanPlayer() (*adhanPlayer, error) {
	return nil, errors.New("Unimplemented.")
}

func (a *adhanPlayer) Play() error {
	// TODO(ssafty): Playing an MP3 file requires rewinding before playing it again.
	return errors.New("Unimplemented.")
}

func (a *adhanPlayer) IsPlaying() bool {
	return false
}
>>>>>>> d137cdea57e1a2c563afe05d727b4751b1223c19:player.go
