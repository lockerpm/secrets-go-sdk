package locker

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"sdk-test/types"
)

func validateMac(encString string, macKey []byte) (bool, error) {
	iv, data, macCode, err := parseEncString(encString)
	if err != nil {
		return false, err
	}

	ivBytes, err := base64.StdEncoding.DecodeString(iv)
	if err != nil {
		types.CURRENT_ERR = types.ERR_DATA
		return false, fmt.Errorf("checking MAC: Invalid base64 iv string")
	}

	dataBytes, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		types.CURRENT_ERR = types.ERR_DATA
		return false, fmt.Errorf("checking MAC: Invalid base64 data string")
	}

	macCodeBytes, err := base64.StdEncoding.DecodeString(macCode)
	if err != nil {
		types.CURRENT_ERR = types.ERR_INPUT
		return false, fmt.Errorf("checking MAC: Invalid base64 mac code string")
	}

	dataToValidate := append(ivBytes, dataBytes...)

	mac := hmac.New(sha256.New, macKey)
	_, err = mac.Write(dataToValidate)
	if err != nil {
		types.CURRENT_ERR = types.ERR_FUNC
		return false, fmt.Errorf("checking MAC: Error calculating mac code")
	}
	calculatedMac := mac.Sum(nil)

	return hmac.Equal(macCodeBytes, calculatedMac), nil
}
