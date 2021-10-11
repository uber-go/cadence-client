// Copyright (c) 2021 Uber Technologies, Inc.
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

package internal

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/cristalhq/jwt/v3"
	"github.com/stretchr/testify/suite"
	"go.uber.org/cadence/internal/common/auth"
	"go.uber.org/cadence/internal/common/util"
)

type (
	jwtAuthSuite struct {
		suite.Suite
		key []byte
	}
)

func TestJWTAuthSuite(t *testing.T) {
	suite.Run(t, new(jwtAuthSuite))
}

func (s *jwtAuthSuite) SetupTest() {
	var err error
	s.key, err = ioutil.ReadFile("./common/auth/credentials/keytest")
	s.NoError(err)
}

func (s *jwtAuthSuite) TearDownTest() {
}

func (s *jwtAuthSuite) TestCorrectTokenCreation() {
	authorizer := NewAdminJwtAuthorizationProvider(s.key)
	authToken, err := authorizer.GetAuthToken()
	s.NoError(err)

	// Decrypting token, it should be enough to make sure authtoken is not empty, this is one steap ahead of that
	publicKeyStr, err := ioutil.ReadFile("./common/auth/credentials/keytest.pub")
	s.NoError(err)
	publicKey, err := util.LoadRSAPublicKey(publicKeyStr)
	s.NoError(err)
	verifier, err := jwt.NewVerifierRS(jwt.RS256, publicKey)
	s.NoError(err)
	token, err := jwt.ParseAndVerifyString(string(authToken), verifier)
	s.NoError(err)
	var claims auth.JWTClaims
	_ = json.Unmarshal(token.RawClaims(), &claims)
	s.Equal(claims.Admin, true)
	s.Equal(claims.Groups, "")
	s.Equal(claims.TTL, int64(60*10))
}

func (s *jwtAuthSuite) TestIncorrectPrivateKeyForTokenCreation() {
	authorizer := NewAdminJwtAuthorizationProvider([]byte{})
	_, err := authorizer.GetAuthToken()
	s.EqualError(err, "failed to parse PEM block containing the private key")
}
