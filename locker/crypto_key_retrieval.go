package locker

import (
	"crypto/sha256"
	"fmt"
	"io"
	"sdk-test/types"

	"golang.org/x/crypto/hkdf"
)

/*========================================================================================================================*/
/* GenerateKey() takes in:
/* 	- An access key from the --access-key flag (byte array)
/* and returns:
/*	- An stretched encryption key (byte array)
/*	- A Mac key (byte array)
/*	- An error (error) or nil
/* Intended to be used to create a stretched key/Mac key pair to decrypt the encrypted symmetric key
/*
/* 1. Calls StretchKey() to generate the stretched encryption key
/* 2. Calls StretchKey() to generate the Mac key
/*========================================================================================================================*/
func generateKey(accessKey []byte) ([]byte, []byte, error) {
	stretchedEncKey, err := stretchKey(accessKey, "enc")
	if err != nil {
		return nil, nil, err
	}
	macKey, err := stretchKey(accessKey, "mac")
	if err != nil {
		return nil, nil, err
	}

	return stretchedEncKey, macKey, nil
}

/*========================================================================================================================*/
/* StretchKey() takes in:
/* 	- An key to be stretched (byte array)
/* and returns:
/*	- A hkdf-stretched key (byte array)
/*	- An error (error) or nil
/*
/* 1. Uses hkdf to stretch the input key to a 32-bytes key
/*========================================================================================================================*/
func stretchKey(keyToBeStretched []byte, param string) ([]byte, error) {
	hkdfReader := hkdf.Expand(sha256.New, keyToBeStretched, []byte(param))
	stretchedKey := make([]byte, 32)
	_, err := io.ReadFull(hkdfReader, stretchedKey)
	if err != nil {
		types.CURRENT_ERR = types.ERR_FUNC
		return nil, fmt.Errorf("error stretching key: %w", err)
	}
	return stretchedKey, nil
}

/*========================================================================================================================*/
/* GetSymKey() takes in:
/* 	- The encrypted symmetric key (string)
/* 	- The stretched encryption key (byte array)
/* 	- The MAC key (byte array)
/* and returns:
/*	- The main symmetric key (byte array)
/*	- The main MAC key (byte array)
/*	- An error (error) or nil
/*
/* 1. Calls ValidateMac() to validate the MAC code of the encrypted symmetric key string
/* 	1.1. If the chekc passes, calls AES256Decrypt() to decrypt the metric key string
/* 	1.2. Retrieves the main symmetric key from the first 32 bytes of the decrypted byte array
/* 	1.3. Retrieves the main MAC key from the second 32 bytes of the decrypted byte array
/*========================================================================================================================*/
func getSymKey(encString string, stretchedEncKey []byte, macKey []byte) ([]byte, []byte, error) {
	macRes, err := validateMac(encString, macKey)
	if err != nil {
		return nil, nil, err
	}

	if !macRes {
		types.CURRENT_ERR = types.ERR_FUNC
		return nil, nil, fmt.Errorf("ValidateMac() failed")
	}

	key, err := aes256Decrypt(encString, stretchedEncKey)
	if err != nil {
		return nil, nil, err
	}
	if len(key) < 64 {
		return nil, nil, fmt.Errorf("decrypted key is of invalid length")
	}
	symmetricKey := key[:32]
	symmetricMacKey := key[32:64]
	return symmetricKey, symmetricMacKey, nil
}
