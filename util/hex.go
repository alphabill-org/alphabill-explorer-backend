package util

import (
	"encoding/hex"
	"fmt"
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
