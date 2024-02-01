// Copyright 2020 the Pinniped contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package pkce

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
)

// Generate generates a new random PKCE code.
func Generate() (Code, error) { return generate(rand.Reader) }

func generate(rand io.Reader) (Code, error) {
	// From https://tools.ietf.org/html/rfc7636#section-4.1:
	//   code_verifier = high-entropy cryptographic random STRING using the
	//   unreserved characters [A-Z] / [a-z] / [0-9] / "-" / "." / "_" / "~"
	//   from Section 2.3 of [RFC3986], with a minimum length of 43 characters
	//   and a maximum length of 128 characters.
	var buf [32]byte
	if _, err := io.ReadFull(rand, buf[:]); err != nil {
		return "", fmt.Errorf("could not generate PKCE code: %w", err)
	}
	return Code(hex.EncodeToString(buf[:])), nil
}

// Code implements the basic options required for RFC 7636: Proof Key for Code Exchange (PKCE).
type Code string

// Challenge returns the URL encoded PKCE code challenge hash.
func (p *Code) Challenge() string {
	b := sha256.Sum256([]byte(*p))
	return base64.RawURLEncoding.EncodeToString(b[:])
}

// Method returns the PKCE code challenge method.
func (p *Code) Method() string {
	return "S256"
}

// Verifier returns the PKCE code verifier.
func (p *Code) Verifier() string {
	return string(*p)
}
