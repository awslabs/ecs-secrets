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

package store

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/awslabs/ecs-secrets/modules/api"

	"github.com/awslabs/ecs-secrets/modules/crypt/mock"
	"github.com/awslabs/ecs-secrets/modules/dao"
	"github.com/awslabs/ecs-secrets/modules/dao/mock"
	"github.com/golang/mock/gomock"
)

func TestGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDAO := mock_dao.NewMockDAO(ctrl)
	crypter := mock_crypt.NewMockCrypter(ctrl)

	loadedSecret := &dao.SecretRecord{
		Name:   "bar",
		Active: true,
	}
	gomock.InOrder(
		mockDAO.EXPECT().GetSecretRecord("foo", int64(1)).Return(loadedSecret, nil),
		crypter.EXPECT().DecryptSecret(loadedSecret).Return(aws.String("foobar"), nil),
	)

	secretStore := NewStore("myapp", mockDAO, crypter)
	secret, err := secretStore.Get("foo", "1")
	if err != nil {
		t.Fatalf("Error getting secret: %v", err)
	}
	expectedSecret := &api.SecretRecord{
		Name:    "bar",
		Active:  true,
		Payload: "foobar",
	}
	if !reflect.DeepEqual(secret, expectedSecret) {
		t.Errorf("Mismatch between expected and retrieved secret: %v != %v", secret, expectedSecret)
	}
}

func TestGetNoSerial(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDAO := mock_dao.NewMockDAO(ctrl)
	crypter := mock_crypt.NewMockCrypter(ctrl)

	loadedSecret := &dao.SecretRecord{
		Name:   "bar",
		Active: true,
	}
	gomock.InOrder(
		mockDAO.EXPECT().GetLatestVersion("foo").Return(loadedSecret, nil),
		crypter.EXPECT().DecryptSecret(loadedSecret).Return(aws.String("foobar"), nil),
	)

	secretStore := NewStore("myapp", mockDAO, crypter)
	secret, err := secretStore.Get("foo", "")
	if err != nil {
		t.Fatalf("Error getting secret: %v", err)
	}
	expectedSecret := &api.SecretRecord{
		Name:    "bar",
		Active:  true,
		Payload: "foobar",
	}
	if !reflect.DeepEqual(secret, expectedSecret) {
		t.Errorf("Mismatch between expected and retrieved secret: %v != %v", secret, expectedSecret)
	}
}

func TestGetInactiveSecret(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDAO := mock_dao.NewMockDAO(ctrl)
	crypter := mock_crypt.NewMockCrypter(ctrl)

	loadedSecret := &dao.SecretRecord{
		Name:          "bar",
		Active:        false,
		EncryptedData: "foobar",
	}

	mockDAO.EXPECT().GetLatestVersion("foo").Return(loadedSecret, nil)

	secretStore := NewStore("myapp", mockDAO, crypter)
	fetchedSecret, err := secretStore.Get("foo", "")
	if err != nil {
		t.Errorf("Error getting inactive secret: %v", err)
	}
	if fetchedSecret.Payload != "" {
		t.Error("Expected payload to be empty for inactive secret")
	}
}

func TestGetDecryptSecretError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDAO := mock_dao.NewMockDAO(ctrl)
	crypter := mock_crypt.NewMockCrypter(ctrl)

	loadedSecret := &dao.SecretRecord{
		Name:   "bar",
		Active: true,
	}
	gomock.InOrder(
		mockDAO.EXPECT().GetLatestVersion("foo").Return(loadedSecret, nil),
		crypter.EXPECT().DecryptSecret(loadedSecret).Return(nil, fmt.Errorf("denied")),
	)

	secretStore := NewStore("myapp", mockDAO, crypter)
	_, err := secretStore.Get("foo", "")
	if err == nil {
		t.Errorf("Expected error getting secret")
	}
}

func TestGetSecretDoesNotExist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDAO := mock_dao.NewMockDAO(ctrl)
	crypter := mock_crypt.NewMockCrypter(ctrl)

	mockDAO.EXPECT().GetLatestVersion("foo").Return(nil, nil)

	secretStore := NewStore("myapp", mockDAO, crypter)
	_, err := secretStore.Get("foo", "")
	if err == nil {
		t.Error("Expected error getting non existent secret")
	}
}

func TestGetDAOError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDAO := mock_dao.NewMockDAO(ctrl)
	crypter := mock_crypt.NewMockCrypter(ctrl)

	mockDAO.EXPECT().GetLatestVersion("foo").Return(nil, fmt.Errorf("come back in a thousand years"))

	secretStore := NewStore("myapp", mockDAO, crypter)
	_, err := secretStore.Get("foo", "")
	if err == nil {
		t.Error("Expected error getting inactive secret")
	}
}

func TestRevoke(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDAO := mock_dao.NewMockDAO(ctrl)
	crypter := mock_crypt.NewMockCrypter(ctrl)

	mockDAO.EXPECT().RevokeSecretRecord("foo", int64(1)).Return(nil)

	secretStore := NewStore("myapp", mockDAO, crypter)
	err := secretStore.Revoke("foo", "1")
	if err != nil {
		t.Errorf("Error revoking secret: %v", err)
	}
}

func TestSave(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDAO := mock_dao.NewMockDAO(ctrl)
	crypter := mock_crypt.NewMockCrypter(ctrl)

	loadedSecret := &dao.SecretRecord{
		Name:   "bar",
		Active: true,
		Serial: 1,
	}

	newSecret := &dao.SecretRecord{
		Name:   "bar",
		Active: true,
		Serial: 2,
	}
	gomock.InOrder(
		mockDAO.EXPECT().GetLatestVersion("bar").Return(loadedSecret, nil),
		crypter.EXPECT().EncryptSecret(newSecret, "foobar").Return(nil, nil),
		mockDAO.EXPECT().PutSecretRecord(newSecret).Return(nil),
	)

	secretStore := NewStore("myapp", mockDAO, crypter)

	apiSecret := &api.SecretRecord{
		Name:    "bar",
		Active:  true,
		Serial:  1,
		Payload: "foobar",
	}
	secret, err := secretStore.Save(apiSecret)
	if err != nil {
		t.Fatalf("Error saving secret: %v", err)
	}
	expectedSecret := &api.SecretRecord{
		Name:    "bar",
		Active:  true,
		Payload: "foobar",
		Serial:  2,
	}
	if !reflect.DeepEqual(secret, expectedSecret) {
		t.Errorf("Mismatch between expected and saved secret: %v != %v", secret, expectedSecret)
	}
}

func TestSaveNoLatestVersion(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDAO := mock_dao.NewMockDAO(ctrl)
	crypter := mock_crypt.NewMockCrypter(ctrl)

	newSecret := &dao.SecretRecord{
		Name:   "bar",
		Active: true,
		Serial: 1,
	}
	gomock.InOrder(
		mockDAO.EXPECT().GetLatestVersion("bar").Return(nil, nil),
		crypter.EXPECT().EncryptSecret(newSecret, "foobar").Return(nil, nil),
		mockDAO.EXPECT().PutSecretRecord(newSecret).Return(nil),
	)

	secretStore := NewStore("myapp", mockDAO, crypter)

	apiSecret := &api.SecretRecord{
		Name:    "bar",
		Active:  true,
		Serial:  1,
		Payload: "foobar",
	}
	secret, err := secretStore.Save(apiSecret)
	if err != nil {
		t.Fatalf("Error saving secret: %v", err)
	}
	if !reflect.DeepEqual(secret, apiSecret) {
		t.Errorf("Mismatch between expected and saved secret: %v != %v", secret, apiSecret)
	}
}

func TestSaveOnEncryptError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDAO := mock_dao.NewMockDAO(ctrl)
	crypter := mock_crypt.NewMockCrypter(ctrl)

	newSecret := &dao.SecretRecord{
		Name:   "bar",
		Active: true,
		Serial: 1,
	}
	gomock.InOrder(
		mockDAO.EXPECT().GetLatestVersion("bar").Return(nil, nil),
		crypter.EXPECT().EncryptSecret(newSecret, "foobar").Return(nil, fmt.Errorf("what's the point")),
	)

	secretStore := NewStore("myapp", mockDAO, crypter)

	apiSecret := &api.SecretRecord{
		Name:    "bar",
		Active:  true,
		Serial:  1,
		Payload: "foobar",
	}
	_, err := secretStore.Save(apiSecret)
	if err == nil {
		t.Error("Expected error saving secret")
	}
}
