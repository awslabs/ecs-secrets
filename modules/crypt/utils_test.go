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

import "testing"

func TestGetCryptoKey(t *testing.T) {
	key := []byte("super-awesome-aes-key-so-secure?")
	cryptoKey, err := getCryptoKey(key)
	if err != nil {
		t.Errorf("Error converting to crypto key: %v", err)
	}
	if len(*cryptoKey) != 32 {
		t.Error("Incorrect length returned for crypto key")
	}
}
