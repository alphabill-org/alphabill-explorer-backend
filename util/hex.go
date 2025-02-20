package util

import (
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

// Decode decodes a hex string with 0x prefix.
func Decode(input string) ([]byte, error) {
	if len(input) == 0 {
		return nil, errors.New("empty hex string")
	}
	if !has0xPrefix(input) {
		return nil, errors.New("hex string without 0x prefix")
	}
	return hex.DecodeString(input[2:])
}

func has0xPrefix(input string) bool {
	return len(input) >= 2 && input[0] == '0' && (input[1] == 'x' || input[1] == 'X')
}
