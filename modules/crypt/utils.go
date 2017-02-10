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
	"bytes"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/gtank/cryptopasta"
)

func base64Encode(input []byte) string {
	return base64.StdEncoding.EncodeToString(input)
}

func base64Decode(input string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(input)
}

func getCryptoKey(input []byte) (*[32]byte, error) {
	reader := bytes.NewReader(input)
	key := [32]byte{}
	_, err := io.ReadFull(reader, key[:])
	if err != nil {
		return nil, fmt.Errorf("Error reading data key: %v", err)
	}
	return &key, nil
}

func encrypt(secret string, key []byte) ([]byte, error) {
	cryptoKey, err := getCryptoKey(key)
	if err != nil {
		return nil, err
	}
	encryptedBlob, err := cryptopasta.Encrypt([]byte(secret), cryptoKey)
	if err != nil {
		return nil, fmt.Errorf("Error encrypting secret: %v", err)
	}

	return encryptedBlob, nil
}

func decrypt(data, key []byte) ([]byte, error) {
	cryptoKey, err := getCryptoKey(key)
	if err != nil {
		return nil, err
	}
	decryptedData, err := cryptopasta.Decrypt(data, cryptoKey)
	if err != nil {
		return nil, fmt.Errorf("Error decrypting secret: %v", err)
	}

	return decryptedData, nil
}
