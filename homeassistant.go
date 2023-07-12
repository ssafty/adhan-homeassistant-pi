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

// Home assistant API handler. Talks to Home assistant with REST commands to
// control specific entities. An entity example is a Zigbee switch or a lightbulb.

package main

import (
	"errors"
	"fmt"
)

type homeassistant struct {
	switchID string
	ipAddr   string
}

// Initializes HomeAssistant instance with a specific switch.
func NewHomeAssistant() (*homeassistant, error) {
	return nil, errors.New("Unimplemented.")
}

// makeSwitchAction is a private function that builds and sends the POST request
// to home assistant to turn the switch on or off.
func (h *homeassistant) makeSwitchAction(action string) (string, error) {
	return "", errors.New("Unimplemented.")
}

// TurnSwitchOn turns the switch on.
func (h *homeassistant) TurnSwitchOn() (string, error) {
	resp, err := h.makeSwitchAction("TURN_ON")
	if err != nil {
		return "", fmt.Errorf("Unimplemented error response: %w", err)
	}
	return resp, nil
}

// TurnSwitchOff turns the switch off.
func (h *homeassistant) TurnSwitchOff() (string, error) {
	resp, err := h.makeSwitchAction("TURN_OFF")
	if err != nil {
		return "", fmt.Errorf("Unimplemented error response: %w", err)
	}
	return resp, nil
}
