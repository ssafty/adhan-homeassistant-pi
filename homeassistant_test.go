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
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"
)

const (
	validIp          = "vIp"
	validSwitchId    = "vswitchId"
	validAuthToken   = "vauthtoken"
	invalidIp        = "ivIp"
	invalidSwitchId  = "ivswitchId"
	invalidAuthToken = "ivauthtoken"
)

type homeassistantHttpClientMock struct {
	ip        string
	authToken string
	switchId  string
}

func (c *homeassistantHttpClientMock) Do(req *http.Request) (*http.Response, error) {
	reader := ioutil.NopCloser(bytes.NewReader([]byte("resp")))

	statusCode := 200
	switch {
	case c.ip != validIp:
		return nil, errors.New("Connection timeout")
	case c.switchId != validSwitchId:
		statusCode = 404
	case c.authToken != validAuthToken:
		statusCode = 401
	}

	return &http.Response{
		StatusCode: statusCode,
		Body:       reader,
	}, nil
}

func TestValidNewHomeAssistant(t *testing.T) {
	h, err := NewHomeAssistant(
		HTTPClient(&httpclient{
			client: &homeassistantHttpClientMock{
				ip:        validIp,
				authToken: validAuthToken,
				switchId:  validSwitchId,
			},
			token: validAuthToken,
		}),
		IPAddress(validIp),
		SwitchID(validSwitchId))
	if err != nil {
		t.Fatalf("NewHomeAssistant with valid arguments should raise no errors. Got %v", err)
	}
	if _, err := h.TurnSwitchOff(); err != nil {
		t.Fatalf("NewHomeAssistant turn switch off action expect no errors. Got %v", err)
	}
	if _, err := h.TurnSwitchOn(); err != nil {
		t.Fatalf("NewHomeAssistant turn switch on action expect no errors. Got %v", err)
	}
}

func TestInvalidNewHomeAssistant(t *testing.T) {
	for _, test := range []struct {
		description string
		ip          string
		authToken   string
		switchId    string
	}{
		{
			description: "Auth/Switch/IP are not provided",
		},
		{
			description: "Switch/IP are not provided",
			authToken:   validAuthToken,
		},
		{
			description: "IP is not provided",
			authToken:   validAuthToken,
			switchId:    validSwitchId,
		},
		{
			description: "Auth/Switch/IP are provided, but invalid Auth",
			authToken:   invalidAuthToken,
			switchId:    validSwitchId,
			ip:          validIp,
		},
		{
			description: "Auth/Switch/IP are provided, but invalid switch Id",
			authToken:   validAuthToken,
			switchId:    invalidSwitchId,
			ip:          validIp,
		},
		{
			description: "Auth/Switch/IP are provided, but invalid ip",
			authToken:   validAuthToken,
			switchId:    validSwitchId,
			ip:          invalidIp,
		},
	} {
		t.Run(test.description, func(t *testing.T) {
			_, err := NewHomeAssistant(
				HTTPClient(&httpclient{
					client: &homeassistantHttpClientMock{
						ip:        test.ip,
						authToken: test.authToken,
						switchId:  test.switchId,
					},
					token: test.authToken,
				}),
				IPAddress(test.ip),
				SwitchID(test.switchId))

			if err == nil {
				t.Errorf("NewHomeAssistant should raise an input validation error. Got none.")
			}
		})
	}
}
