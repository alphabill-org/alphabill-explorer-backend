package util

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
)

const (
	PubKeyBytesLength     = 33
	PubKeyHashBytesLength = 32
)

func ToHex(src []byte) []byte {
	if len(src) == 0 {
		return nil
	}
	dst := make([]byte, len(src)*2+2)
	copy(dst, `0x`)
	hex.Encode(dst[2:], src)
	return dst
}

func FromHex(src []byte) ([]byte, error) {
	src, err := CheckHex(src)
	if err != nil {
		return nil, err
	}
	if len(src) == 0 {
		return nil, nil
	}
	dst := make([]byte, len(src)/2)
	_, err = hex.Decode(dst, src)
	if err != nil {
		return nil, err
	}
	return dst, nil
}

func CheckHex(input []byte) ([]byte, error) {
	if len(input) == 0 {
		return nil, nil
	}
	if len(input) >= 2 && input[0] == '0' && (input[1] == 'x' || input[1] == 'X') {
		input = input[2:]
	} else {
		return nil, fmt.Errorf("hex string without 0x prefix")
	}
	if len(input)%2 != 0 {
		return nil, fmt.Errorf("hex string of odd length")
	}
	return input, nil
}

// PubKeyHash returns the hash of the given hex-encoded public key
// If the input is already a hash of the public key, it just returns the hash
func PubKeyHash(pubKeyHex string) ([]byte, error) {
	bytes, err := DecodeHex(pubKeyHex)
	if err != nil {
		return nil, err
	}
	var pubKeyHash []byte
	if len(bytes) == PubKeyBytesLength {
		hash := sha256.Sum256(bytes)
		pubKeyHash = hash[:]
	} else if len(bytes) == PubKeyHashBytesLength {
		pubKeyHash = bytes
	} else {
		return nil, errors.New("hex string is not public key or it's hash")
	}
	return pubKeyHash, nil
}

// DecodeHex decodes a hex string with optional 0x prefix
func DecodeHex(input string) ([]byte, error) {
	if len(input) == 0 {
		return nil, errors.New("empty hex string")
	}
	if !has0xPrefix(input) {
		return hex.DecodeString(input)
	}
	return hex.DecodeString(input[2:])
}

func has0xPrefix(input string) bool {
	return len(input) >= 2 && input[0] == '0' && (input[1] == 'x' || input[1] == 'X')
}
