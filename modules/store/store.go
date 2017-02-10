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
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	log "github.com/cihub/seelog"

	"github.com/awslabs/ecs-secrets/modules/api"

	"github.com/awslabs/ecs-secrets/modules/crypt"
	"github.com/awslabs/ecs-secrets/modules/dao"
)

//go:generate mockgen.sh github.com/awslabs/ecs-secrets/modules/store Store mock/store_mock.go

// Store defines the secret store interface
type Store interface {
	Get(string, string) (*api.SecretRecord, error)
	Save(*api.SecretRecord) (*api.SecretRecord, error)
	Revoke(string, string) error
}

type store struct {
	dao     dao.DAO
	crypter crypt.Crypter
}

// NewStore creates a new secret store backed by DynamoDB
func NewStore(appName string, dao dao.DAO, crypter crypt.Crypter) Store {
	return &store{
		dao:     dao,
		crypter: crypter,
	}
}

// Get gets a secret from the store
func (s *store) Get(name string, serial string) (*api.SecretRecord, error) {
	var loadedSecret *dao.SecretRecord
	var err error
	if serial == "" {
		loadedSecret, err = s.dao.GetLatestVersion(name)
	} else {
		serialInt, err := strconv.Atoi(serial)
		if err != nil {
			return nil, err
		}
		loadedSecret, err = s.dao.GetSecretRecord(name, int64(serialInt))
	}

	if err != nil {
		log.Errorf("Error getting secret record for: %s, %v", name, err)
		return nil, err
	}

	if loadedSecret == nil {
		return nil, fmt.Errorf("Secret with name '%s' not found", name)
	}

	secretRecord := &api.SecretRecord{
		Name:   loadedSecret.Name,
		Serial: loadedSecret.Serial,
		Active: loadedSecret.Active,
	}
	if !loadedSecret.Active {
		log.Debugf("Returning inactive secret; name: %s, serial: %d", secretRecord.Name, secretRecord.Serial)
		return secretRecord, nil
	}

	decryptedSecret, err := s.crypter.DecryptSecret(loadedSecret)
	if err != nil {
		log.Errorf("Error decrypting secret for: %s, %v", name, err)
		return nil, err
	}
	secretRecord.Payload = aws.StringValue(decryptedSecret)

	return secretRecord, nil
}

// Revoke revokes the secret from the store
func (s *store) Revoke(name string, serial string) error {
	serialInt, err := strconv.Atoi(serial)
	if err != nil {
		return err
	}
	return s.dao.RevokeSecretRecord(name, int64(serialInt))
}

// Save saves the secret into the store
func (s *store) Save(passedSecret *api.SecretRecord) (*api.SecretRecord, error) {
	// get latest revision, increment serial by 1
	latestSecret, err := s.dao.GetLatestVersion(passedSecret.Name)
	if err != nil {
		return nil, err
	}
	if latestSecret != nil {
		passedSecret.Serial = latestSecret.Serial + 1
	}

	newSecret := &dao.SecretRecord{
		Name:   passedSecret.Name,
		Serial: passedSecret.Serial,
		Active: passedSecret.Active,
	}

	_, err = s.crypter.EncryptSecret(newSecret, passedSecret.Payload)
	if err != nil {
		log.Errorf("Error encrypting secret record for: %s, %v", passedSecret.Name, err)
		return nil, err
	}
	err = s.dao.PutSecretRecord(newSecret)
	return passedSecret, err
}
