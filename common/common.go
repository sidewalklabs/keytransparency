// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// This package contains common type definitions and functions used by other
// packages. Types that can cause circular import should be added here.
package common

import (
	"encoding/binary"
	"crypto/hmac"
	"crypto/sha256"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

const (
	// HashSize contains the blocksize of the used hash function in bytes.
	HashSize = sha256.Size
)

var (
	// TreeNonce is a constant value used as a salt in all leaf node calculations.
	// The TreeNonce prevents different realms from producing collisions.
	TreeNonce = []byte{241, 71, 100, 55, 62, 119, 69, 16, 150, 179, 228, 81, 34, 200, 144, 6}
	// LeafIdentifier is the data used to indicate a leaf node.
	LeafIdentifier = []byte("L")
	// EmptyIdentifier is used while calculating the data of nil sub branches.
	EmptyIdentifier = []byte("E")
)

// GenerateProfileCommitment calculates and returns the profile commitment based
// on the provided nonce. Commitment is HMAC(profile, nonce).
func GenerateProfileCommitment(nonce []byte, profile []byte) ([]byte, error) {
	mac := hmac.New(sha256.New, nonce)
	if _, err := mac.Write(profile); err != nil {
		return nil, grpc.Errorf(codes.Internal, "Error while generating profile commitment: %v", err)
	}
	return mac.Sum(nil), nil
}

// VerifyProfileCommitment returns nil if the profile commitment using the
// nonce matches the provided commitment, and error otherwise.
func VerifyProfileCommitment(nonce []byte, profile []byte, commitment []byte) error {
	expectedCommitment, err := GenerateProfileCommitment(nonce, profile)
	if err != nil {
		return err
	}
	if !hmac.Equal(expectedCommitment, commitment) {
		return grpc.Errorf(codes.InvalidArgument, "Invalid profile commitment")
	}
	return nil
}

// HashLeaf calculate the merkle tree leaf node value. This is computed as
// H(TreeNonce || Identifier || depth || index || dataHash), where TreeNonce,
// Identifier, depth, and index are fixed-length.
func HashLeaf(identifier []byte, depth int, index []byte, dataHash []byte) []byte {
	bdepth := make([]byte, 4)
	binary.BigEndian.PutUint32(bdepth, uint32(depth))

	h := sha256.New()
	h.Write(TreeNonce)
	h.Write(identifier)
	h.Write(bdepth)
	h.Write(index)
	h.Write(dataHash)
	return h.Sum(nil)
}

// HashIntermediateNode calculates an interior node's value by H(left || right)
func HashIntermediateNode(left []byte, right []byte) []byte {
	h := sha256.New()
	h.Write(left)
	h.Write(right)
	return h.Sum(nil)
}

// EmptyLeafValue computes the value of an empty leaf as
// H(TreeNonce || EmptyIdentifier || depth || index), where TreeNonce,
// EmptyIdentifier, depth, and index are fixed-length.
func EmptyLeafValue(prefix string) []byte {
	return HashLeaf(EmptyIdentifier, len(prefix), []byte(prefix), nil)
}

// Hash calculates the hash of the given data.
func Hash(data []byte) []byte {
	dataHash := sha256.Sum256(data)
	return dataHash[:]
}
