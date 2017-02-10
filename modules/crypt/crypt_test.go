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

package crypt

import (
	"fmt"
	"testing"

	"github.com/awslabs/ecs-secrets/modules/cache/mock"
	"github.com/awslabs/ecs-secrets/modules/dao"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/awslabs/ecs-secrets/modules/kms/client/mock"

	"github.com/golang/mock/gomock"
)

// aesKey is the 32 byte long key used for testing
const aesKey = "super-awesome-aes-key-so-secure?"

func TestEncryptSecret(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	kmsClient := mock_client.NewMockClient(ctrl)
	kmsClient.EXPECT().GenerateDataKey(&kms.GenerateDataKeyInput{
		KeySpec: aws.String("AES_256"),
		KeyId:   aws.String("alias/ECSSecretsMaskerKey-myapp"),
	}).Return(&kms.GenerateDataKeyOutput{
		Plaintext:      []byte(aesKey),
		CiphertextBlob: []byte(""),
	}, nil)

	cache := mock_cache.NewMockCache(ctrl)
	crypter := NewCrypter(kmsClient, cache, "myapp")
	secret := &dao.SecretRecord{}
	_, err := crypter.EncryptSecret(secret, "mysecret")
	if err != nil {
		t.Errorf("Error encrypting secret: %v", err)
	}
}

func TestEncryptSecretOnGenerateDataKeyError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	kmsClient := mock_client.NewMockClient(ctrl)
	kmsClient.EXPECT().GenerateDataKey(&kms.GenerateDataKeyInput{
		KeySpec: aws.String("AES_256"),
		KeyId:   aws.String("alias/ECSSecretsMaskerKey-myapp"),
	}).Return(nil, fmt.Errorf("no key for you"))

	cache := mock_cache.NewMockCache(ctrl)
	crypter := NewCrypter(kmsClient, cache, "myapp")
	secret := &dao.SecretRecord{}
	_, err := crypter.EncryptSecret(secret, "mysecret")
	if err == nil {
		t.Error("Expected error encrypting secret")
	}
}

func TestEncryptSecretOnAESEncryptError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	kmsClient := mock_client.NewMockClient(ctrl)
	// Return a blank plaintext key. This will cause AES Encryption
	// to fail as key size == 0
	kmsClient.EXPECT().GenerateDataKey(&kms.GenerateDataKeyInput{
		KeySpec: aws.String("AES_256"),
		KeyId:   aws.String("alias/ECSSecretsMaskerKey-myapp"),
	}).Return(&kms.GenerateDataKeyOutput{
		Plaintext:      []byte(""),
		CiphertextBlob: []byte(""),
	}, nil)

	cache := mock_cache.NewMockCache(ctrl)
	crypter := NewCrypter(kmsClient, cache, "myapp")
	secret := &dao.SecretRecord{
		EncryptedData: "",
	}
	_, err := crypter.EncryptSecret(secret, "mysecret")
	if err == nil {
		t.Error("Expected error encrypting secret")
	}
}

func TestDecryptSecretCacheHit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	kmsClient := mock_client.NewMockClient(ctrl)
	cache := mock_cache.NewMockCache(ctrl)

	payload := "one divided by zero is infinity"
	encrypted, err := encrypt(payload, []byte(aesKey))
	if err != nil {
		t.Fatal("Error encrypting data: %v", err)
	}
	b64Encrypted := base64Encode(encrypted)

	secret := &dao.SecretRecord{
		EncryptedDataKey: b64Encrypted,
		EncryptedData:    b64Encrypted,
	}

	cache.EXPECT().Get(secret.EncryptedDataKey).Return([]byte(aesKey), true)

	crypter := NewCrypter(kmsClient, cache, "myapp")
	_, err = crypter.DecryptSecret(secret)
	if err != nil {
		t.Errorf("Error decrypting secret: %v", err)
	}
}

func TestDecryptSecretCacheMiss(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	kmsClient := mock_client.NewMockClient(ctrl)
	cache := mock_cache.NewMockCache(ctrl)

	payload := "one divided by zero is infinity"
	encrypted, err := encrypt(payload, []byte(aesKey))
	if err != nil {
		t.Fatal("Error encrypting data: %v", err)
	}
	b64Encrypted := base64Encode(encrypted)

	secret := &dao.SecretRecord{
		EncryptedDataKey: b64Encrypted,
		EncryptedData:    b64Encrypted,
	}

	gomock.InOrder(
		cache.EXPECT().Get(secret.EncryptedDataKey).Return(nil, false),
		kmsClient.EXPECT().Decrypt(&kms.DecryptInput{
			CiphertextBlob: encrypted,
		}).Return(&kms.DecryptOutput{
			Plaintext: []byte(aesKey),
		}, nil),
		cache.EXPECT().Set(secret.EncryptedDataKey, []byte(aesKey)),
	)

	crypter := NewCrypter(kmsClient, cache, "myapp")
	_, err = crypter.DecryptSecret(secret)
	if err != nil {
		t.Errorf("Error decrypting secret: %v", err)
	}
}

func TestDecryptSecretCacheMissOnKMSDecryptError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	kmsClient := mock_client.NewMockClient(ctrl)
	cache := mock_cache.NewMockCache(ctrl)

	payload := "one divided by zero is infinity"
	encrypted, err := encrypt(payload, []byte(aesKey))
	if err != nil {
		t.Fatal("Error encrypting data: %v", err)
	}
	b64Encrypted := base64Encode(encrypted)

	secret := &dao.SecretRecord{
		EncryptedDataKey: b64Encrypted,
		EncryptedData:    b64Encrypted,
	}

	gomock.InOrder(
		cache.EXPECT().Get(secret.EncryptedDataKey).Return(nil, false),
		kmsClient.EXPECT().Decrypt(&kms.DecryptInput{
			CiphertextBlob: encrypted,
		}).Return(nil, fmt.Errorf("world is not ready for this")),
	)

	crypter := NewCrypter(kmsClient, cache, "myapp")
	_, err = crypter.DecryptSecret(secret)
	if err == nil {
		t.Error("Expected error decrypting secret")
	}
}
