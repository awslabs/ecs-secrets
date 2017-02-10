// Copyright 2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//	http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/awslabs/ecs-secrets/modules/api"
	"github.com/awslabs/ecs-secrets/modules/store/mock"
	"github.com/awslabs/ecs-secrets/modules/version"
	"github.com/golang/mock/gomock"
)

func TestVersion(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mock_store.NewMockStore(ctrl)
	s := NewServer(mockStore)
	router := s.Router()
	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/latest/version", nil)
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("Incorrect http status: %v", recorder.Code)
	}
	var response versionResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Error decoding response: %v", err)
	}
	expectedResponse := versionResponse{
		application:        version.AppName,
		applicationVersion: version.Version,
		apiVersion:         version.ApiVersion,
	}
	if reflect.DeepEqual(response, expectedResponse) {
		t.Errorf("Incorrect response. %v != %v", response, expectedResponse)
	}
}

func TestFetchSecretError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mock_store.NewMockStore(ctrl)

	mockStore.EXPECT().Get("foo", "").Return(nil, fmt.Errorf("its already leaked"))
	s := NewServer(mockStore)
	router := s.Router()
	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/latest/secrets/foo", nil)
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("Incorrect http status: %v", recorder.Code)
	}
}

func TestFetchSecret(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mock_store.NewMockStore(ctrl)

	mockStore.EXPECT().Get("foo", "").Return(&api.SecretRecord{Name: "foo"}, nil)
	s := NewServer(mockStore)
	router := s.Router()
	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/latest/secrets/foo", nil)
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("Incorrect http status: %v", recorder.Code)
	}
	var response api.SecretRecord
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Error decoding response: %v", err)
	}
	expectedResponse := api.SecretRecord{
		Name: "foo",
	}
	if !reflect.DeepEqual(response, expectedResponse) {
		t.Errorf("Incorrect response. %v != %v", response, expectedResponse)
	}
}

func TestRevokeSecretsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mock_store.NewMockStore(ctrl)

	mockStore.EXPECT().Revoke("foo", "1").Return(fmt.Errorf("secret not found"))
	s := NewServer(mockStore)
	router := s.Router()
	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/latest/revoke/foo/1", nil)
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("Incorrect http status: %v", recorder.Code)
	}
}

func TestRevokeSecrets(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mock_store.NewMockStore(ctrl)

	mockStore.EXPECT().Revoke("foo", "1").Return(nil)
	s := NewServer(mockStore)
	router := s.Router()
	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/latest/revoke/foo/1", nil)
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("Incorrect http status: %v", recorder.Code)
	}
}

func TestCreateSecretsEmptyPayload(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mock_store.NewMockStore(ctrl)

	s := NewServer(mockStore)
	router := s.Router()
	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/latest/secrets/foo", nil)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("Incorrect http status: %v", recorder.Code)
	}
}

func TestCreateSecretsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mock_store.NewMockStore(ctrl)

	secret := &api.SecretRecord{
		Name:    "foo",
		Serial:  int64(1),
		Payload: "bar",
		Active:  true,
	}
	mockStore.EXPECT().Save(secret).Return(nil, fmt.Errorf("secret not found"))
	s := NewServer(mockStore)
	router := s.Router()
	recorder := httptest.NewRecorder()
	data := api.SecretPayload{
		Payload: "bar",
	}
	dataBytes, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", "/latest/secrets/foo", bytes.NewBuffer(dataBytes))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("Incorrect http status: %v", recorder.Code)
	}
}

func TestCreateSecrets(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mock_store.NewMockStore(ctrl)

	secret := &api.SecretRecord{
		Name:    "foo",
		Serial:  int64(1),
		Payload: "bar",
		Active:  true,
	}
	mockStore.EXPECT().Save(secret).Return(secret, nil)
	s := NewServer(mockStore)
	router := s.Router()
	recorder := httptest.NewRecorder()
	data := api.SecretPayload{
		Payload: "bar",
	}
	dataBytes, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", "/latest/secrets/foo", bytes.NewBuffer(dataBytes))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("Incorrect http status: %v", recorder.Code)
	}
}
