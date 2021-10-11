// Copyright (c) 2021 Uber Technologies Inc.
// Portions of the Software are attributed to Copyright (c) 2020 Temporal Technologies Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package util

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
)

type KeyType string

const (
	KeyTypePrivate KeyType = "private key"

	KeyTypePublic KeyType = "public key"
)

func loadRSAKey(keyString []byte, keyType KeyType) (interface{}, error) {
	block, _ := pem.Decode(keyString)
	if block == nil || strings.ToLower(block.Type) != strings.ToLower(string(keyType)) {
		return nil, fmt.Errorf("failed to parse PEM block containing the %s", keyType)
	}

	switch keyType {
	case KeyTypePrivate:
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse DER encoded %s: %s", keyType, err.Error())
		}
		return key, nil
	case KeyTypePublic:
		key, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse DER encoded %s: %s", keyType, err.Error())
		}
		return key, nil
	default:
		return nil, fmt.Errorf("invalid Key Type")
	}
}

func LoadRSAPublicKey(key []byte) (*rsa.PublicKey, error) {
	rsaKey, err := loadRSAKey(key, KeyTypePublic)
	if err != nil {
		return nil, err
	}
	return rsaKey.(*rsa.PublicKey), err
}

func LoadRSAPrivateKey(key []byte) (*rsa.PrivateKey, error) {
	rsaKey, err := loadRSAKey(key, KeyTypePrivate)
	if err != nil {
		return nil, err
	}
	return rsaKey.(*rsa.PrivateKey), err
}
