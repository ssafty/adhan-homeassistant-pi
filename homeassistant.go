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
	"log"
	"strings"
)

type SwitchAction string

const (
	TURNON  SwitchAction = "/api/services/switch/turn_on"
	TURNOFF SwitchAction = "/api/services/switch/turn_off"
	STATUS  SwitchAction = "/api/states/"
)

type homeassistant struct {
	client *httpclient

	switchID string
	ipAddr   string
}

type homeassistantOpt func(*homeassistant)

func IPAddress(ip string) homeassistantOpt {
	url := strings.TrimRight(ip, "/")

	return func(h *homeassistant) {
		h.ipAddr = url
	}
}

func SwitchID(se string) homeassistantOpt {
	return func(h *homeassistant) {
		h.switchID = se
	}
}

func HTTPClient(c *httpclient) homeassistantOpt {
	return func(h *homeassistant) {
		h.client = c
	}
}

// Initializes HomeAssistant instance with a specific switch. NewHomeAssistant sends
// on creation a GET request to homeassistant to verify that the token/ip are correct.
func NewHomeAssistant(opts ...homeassistantOpt) (*homeassistant, error) {
	ha := &homeassistant{}

	for _, opt := range opts {
		opt(ha)
	}

	switch {
	case ha.client == nil || ha.client.token == "":
		return nil, errors.New("Httpclient with GET/POST features is not specified.")
	case ha.switchID == "":
		return nil, errors.New("NewHomeAssistant's switch id/entity is not specified.")
	case ha.ipAddr == "":
		return nil, errors.New("NewHomeAssistant's IP address is not specified.")
	}

	if _, err := ha.getSwitchStatus(); err != nil {
		// This validation check may fail if the AuthToken, IP address or switchID are incorrect.
		return nil, fmt.Errorf("NewHomeAssistant status validation check failed: %w", err)
	}

	return ha, nil
}

// makeSwitchAction is a private function that builds and sends the POST request
// to home assistant to turn the switch on or off.
func (h *homeassistant) makeSwitchAction(action SwitchAction) (string, error) {
	url := h.ipAddr + string(action)
	payload := map[string]string{
		"entity_id": h.switchID,
	}

	body, statusCode, err := h.client.Post(url, payload)
	if err != nil {
		return "", fmt.Errorf("encountered error from POST(%s, %v) request: %w", url, payload, err)
	}
	if statusCode != 200 {
		return "", fmt.Errorf("unsuccessful response status code. Received statusCode: %d for POST(%s, %v): %v", statusCode, url, payload, body)
	}

	log.Printf("Speaker Action succeeded: %v", action)
	return body, nil
}

// getStatus query the status of the home automation entity that homeassistant
// struct is initialized with i.e. h.switchID.
func (h *homeassistant) getSwitchStatus() (string, error) {
	url := h.ipAddr + string(STATUS) + h.switchID

	body, statusCode, err := h.client.Get(url)
	if err != nil {
		return "", fmt.Errorf("encountered error from Get(%s) request: %w", url, err)
	}
	if statusCode != 200 {
		return "", fmt.Errorf("unsuccessful response status code. Received statusCode: %d for Get(%s): %v", statusCode, url, body)
	}

	log.Printf("Speaker Action succeeded: %v", url)
	return body, nil
}

// TurnSwitchOn turns the switch on.
func (h *homeassistant) TurnSwitchOn() (string, error) {
	resp, err := h.makeSwitchAction(TURNON)
	if err != nil {
		return "", fmt.Errorf("error switching on %v: %w", h.switchID, err)
	}
	return resp, nil
}

// TurnSwitchOff turns the switch off.
func (h *homeassistant) TurnSwitchOff() (string, error) {
	resp, err := h.makeSwitchAction(TURNOFF)
	if err != nil {
		return "", fmt.Errorf("error switching off %v: %w", h.switchID, err)
	}
	return resp, nil
}
