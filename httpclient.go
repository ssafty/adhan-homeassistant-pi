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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// clientInterface is used for mocking Do().
type clientInterface interface {
	Do(req *http.Request) (*http.Response, error)
}

type httpclient struct {
	client clientInterface
	token  string
}

func NewHTTPClient(token string) *httpclient {
	return &httpclient{
		token: token,
	}
}

func (c *httpclient) sendReq(req *http.Request, token string) (string, int, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("received an error on response for req %v: %w", req, err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", 0, fmt.Errorf("error while reading the response bytes for %v: %w", resp.Body, err)
	}

	return string([]byte(body)), resp.StatusCode, nil
}

func (c *httpclient) Get(addr string) (string, int, error) {
	// Create a new request using http
	req, err := http.NewRequest(http.MethodGet, addr, nil)
	if err != nil {
		return "", 0, fmt.Errorf("error on Get NewRequest: %w", err)
	}

	// Send req using http Client
	resp, statusCode, err := c.sendReq(req, c.token)
	if err != nil {
		return "", 0, fmt.Errorf("error while sending Get req %v to the httpClient: %w", req, err)
	}

	return resp, statusCode, nil
}

func (c *httpclient) Post(addr string, payload map[string]string) (string, int, error) {
	jsonload, err := json.Marshal(payload)
	if err != nil {
		return "", 0, fmt.Errorf("error on payload json marshal: %w", err)
	}

	// Create a new request using http
	req, err := http.NewRequest(http.MethodPost, addr, strings.NewReader(string(jsonload)))
	if err != nil {
		return "", 0, fmt.Errorf("error on Post NewRequest: %w", err)
	}

	// Send req using http Client
	resp, statusCode, err := c.sendReq(req, c.token)
	if err != nil {
		return "", 0, fmt.Errorf("error while sending req %v to the httpClient: %w", req, err)
	}

	return resp, statusCode, nil
}
