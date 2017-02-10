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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/awslabs/ecs-secrets/modules/cache"
	"github.com/awslabs/ecs-secrets/modules/dao"
	"github.com/awslabs/ecs-secrets/modules/kms/client"
	"github.com/awslabs/ecs-secrets/modules/kms/utils"
)

//go:generate mockgen.sh github.com/awslabs/ecs-secrets/modules/crypt Crypter mock/crypt_mock.go

const (
	defaultKeySpec = "AES_256"
)

// Crypter defines the interface to encrypt and decrypt secret records
type Crypter interface {
	EncryptSecret(*dao.SecretRecord, string) (*dao.SecretRecord, error)
	DecryptSecret(*dao.SecretRecord) (*string, error)
}

// kmsCrypter implements the Crypter interface to encrypt and decrupt secret records
// using AWS KMS
type kmsCrypter struct {
	kmsClient client.Client
	keyCache  cache.Cache
	appName   string
}

// NewCrypter creates a new Crypter object
func NewCrypter(kmsClient client.Client, keyCache cache.Cache, appName string) Crypter {
	return &kmsCrypter{
		kmsClient: kmsClient,
		keyCache:  keyCache,
		appName:   appName,
	}
}

// EncryptSecret encrypts a secret record using a KMS data key
func (crypter *kmsCrypter) EncryptSecret(secretRecord *dao.SecretRecord, secret string) (*dao.SecretRecord, error) {
	// call kms to get datakey
	result, err := crypter.generateDataKey()
	if err != nil {
		return nil, err
	}

	// encrypt secret with that datakey
	encryptedBlob, err := encrypt(secret, result.Plaintext)
	if err != nil {
		return nil, err
	}

	// store the encrypted datakey and the encrypted data
	secretRecord.EncryptedData = base64Encode(encryptedBlob)
	secretRecord.EncryptedDataKey = base64Encode(result.CiphertextBlob)
	return secretRecord, nil
}

// DecryptSecret decrypts a secret record using a KMS data key
func (crypter *kmsCrypter) DecryptSecret(secretRecord *dao.SecretRecord) (*string, error) {
	// decrypt the datakey
	dataKey, err := crypter.fetchDataKey(secretRecord)
	if err != nil {
		return nil, err
	}

	decodedData, err := base64Decode(secretRecord.EncryptedData)
	if err != nil {
		return nil, err
	}

	// decrypt the data
	decryptedData, err := decrypt(decodedData, dataKey)
	if err != nil {
		return nil, err
	}

	secretData := string(decryptedData)
	return &secretData, nil
}

func (crypter *kmsCrypter) generateDataKey() (*kms.GenerateDataKeyOutput, error) {
	return crypter.kmsClient.GenerateDataKey(&kms.GenerateDataKeyInput{
		KeySpec: aws.String(defaultKeySpec),
		KeyId:   aws.String(utils.GetCMKAlias(crypter.appName)),
	})
}

func (crypter *kmsCrypter) fetchDataKey(loadedSecret *dao.SecretRecord) ([]byte, error) {
	if dataKey, ok := crypter.keyCache.Get(loadedSecret.EncryptedDataKey); ok {
		return dataKey.([]byte), nil
	}

	decodedKey, err := base64Decode(loadedSecret.EncryptedDataKey)
	if err != nil {
		return nil, err
	}

	result, err := crypter.kmsClient.Decrypt(&kms.DecryptInput{
		CiphertextBlob: decodedKey,
	})
	if err != nil {
		return nil, err
	}

	crypter.keyCache.Set(loadedSecret.EncryptedDataKey, result.Plaintext)

	return result.Plaintext, nil
}
