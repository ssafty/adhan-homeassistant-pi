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

type httpclientMock struct {
	err bool

	statusCode int
	resp       string
}

func (c *httpclientMock) Do(req *http.Request) (*http.Response, error) {
	if c.err == true {
		return nil, errors.New("error")
	}

	reader := ioutil.NopCloser(bytes.NewReader([]byte(c.resp)))

	return &http.Response{
		StatusCode: c.statusCode,
		Body:       reader,
	}, nil
}

func TestInvalidSendReq(t *testing.T) {
	httpClient := &httpclientMock{err: true}
	c := &httpclient{
		client: httpClient,
		token:  "test-token",
	}
	if _, _, err := c.Get("test-address"); err == nil {
		t.Errorf("Server side errors should be propagated. want err, got nil")
	}
	if _, _, err := c.Post("test-address", map[string]string{"test": "test"}); err == nil {
		t.Errorf("Server side errors should be propagated. want err, got nil")
	}
}

func TestValidSendReq(t *testing.T) {
	httpClient := &httpclientMock{statusCode: 200, resp: `{"key": "value"}`}
	c := &httpclient{
		client: httpClient,
		token:  "test-token",
	}
	switch _, code, err := c.Get("test-address"); {
	case err != nil:
		t.Fatalf("httpClient Get request should succeed. got err: %v", err)

	case code != 200:
		t.Errorf("httpClient Get request didn't return correct statusCode. got %v want %v", code, 200)
	}

	switch _, code, err := c.Post("test-address", map[string]string{"test": "test"}); {
	case err != nil:
		t.Fatalf("httpClient Post request should succeed. got err: %v", err)
	case code != 200:
		t.Errorf("httpClient Post request didn't return correct statusCode. got %v want %v", code, 200)
	}
}
