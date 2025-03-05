package locker

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/lockerpm/secrets-sdk-go/types"
)

/*========================================================================================================================*/
/* AES256Decrypt() takes in:
/* 	- An encrypted string (string) of the form: 2.{iv}|{cipher text}|{MAC code}
/*	- A symmetric key (byte array)
/* and returns:
/*	- A clear text (byte array)
/*	- An error (error) or nil
/*
/* 1. Calls ParseEncString() to decompse the encrypted string to components (iv and cipher text)
/* 2. Decodes the components (which are in base64) to standard encoding
/* 3. Checks if the cipher text's length is divisible by the aes's block size
/* 	3.1. If yes, decrypts the cipher text with AES-256-CBC
/*	3.2. return the clear text
/*========================================================================================================================*/
func aes256Decrypt(encString string, deKey []byte) ([]byte, error) {
	iv, cipherText, _, err := parseEncString(encString)
	if err != nil {
		return nil, err
	}

	cipherTextBytes, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		types.CURRENT_ERR = types.ERR_DATA
		return nil, fmt.Errorf("decrypting: Invalid base64 encrypted string")
	}

	ivBytes, err := base64.StdEncoding.DecodeString(iv)
	if err != nil {
		types.CURRENT_ERR = types.ERR_DATA
		return nil, fmt.Errorf("decrypting: Invalid base64 iv string")
	}

	if len(cipherTextBytes)%aes.BlockSize != 0 {
		types.CURRENT_ERR = types.ERR_DATA
		return nil, fmt.Errorf("decrypting: Encrypted string is not a multiple of the block size")
	}

	block, err := aes.NewCipher(deKey)
	if err != nil {
		types.CURRENT_ERR = types.ERR_FUNC
		return nil, fmt.Errorf("decrypting: Error creating new cipher")
	}

	mode := cipher.NewCBCDecrypter(block, ivBytes)
	mode.CryptBlocks(cipherTextBytes, cipherTextBytes)

	return cipherTextBytes, nil
}

/*========================================================================================================================*/
/* AES256Encrypt() takes in:
/* 	- A clear text (byte array)
/*	- A symmetric key (byte array)
/* 	- An iv (byte array)
/* and returns:
/*	- An cipher text (byte array)
/*	- An error (error) or nil
/*
/* 1. Checks if the clear text's length is divisible by the aes's block size
/*	1.1. If not, pad the clear text with 0x09
/* 2. Encrypts the clear text with AES-256-CBC
/* 	2.1. Returns the cipher text
/*========================================================================================================================*/
func aes256Encrypt(clearText []byte, encKey []byte, iv []byte) ([]byte, error) {
	var charToPad byte

	if len(clearText)%aes.BlockSize != 0 {
		charToPad = byte(aes.BlockSize - len(clearText)%aes.BlockSize)
		for len(clearText)%aes.BlockSize != 0 {
			clearText = append(clearText, charToPad)
		}
	} else {
		charToPad = byte(16)
		targetLen := len(clearText) + 16
		for len(clearText) < targetLen {
			clearText = append(clearText, charToPad)
		}
	}

	block, err := aes.NewCipher(encKey)
	if err != nil {
		types.CURRENT_ERR = types.ERR_DATA
		return nil, fmt.Errorf("encrypting: Error creating new cipher")
	}

	encryptedData := make([]byte, len(clearText))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(encryptedData, clearText)
	return encryptedData, nil
}

func aes256DecryptToString(encData string, symKey []byte, macKey []byte) (string, error) {
	if encData == "" {
		return "", nil
	}

	decData := ""
	macRes, err := validateMac(encData, macKey)
	if err != nil {
		if strings.Contains(err.Error(), "Empty") {
			return "", nil
		}
		return "", err
	}

	if !macRes {
		types.CURRENT_ERR = types.ERR_INPUT
		return "", fmt.Errorf("decrypting: MAC check failed")
	}

	decRes, err := aes256Decrypt(encData, symKey)
	if err != nil {
		return "", err
	}

	if len(decRes) > 0 {
		paddingTrim := len(decRes) - int(decRes[len(decRes)-1])
		if paddingTrim >= 0 {
			decRes = decRes[:paddingTrim]
		}
	}

	decData = string(decRes)

	return decData, nil
}

/*========================================================================================================================*/
/* AES256EncryptToString() takes in:
/* 	- A clear text (byte array)
/*	- A symmetric key (byte array)
/* 	- A MAC key (byte array)
/* and returns:
/*	- An encrypted text (string) of the form: 2.{iv}|{cipher text}|{MAC code}
/*	- An error (error) or nil
/*
/* 1. Makes a random iv
/* 2. Calls AES256Encrypt() to encrypt the clear text with the above iv
/* 3. Makes a MAC code with the MAC key, the created cipher text and the iv
/* 4. Encodes the MAC code, the cipher text and the iv in base 64
/* 5. Concatenates them in the form 2.{iv}|{cipher text}|{MAC code} and returns this string
/*========================================================================================================================*/
func aes256EncryptToString(clearText []byte, encKey []byte, macKey []byte) (string, error) {
	iv := make([]byte, 16)
	_, err := rand.Read(iv)
	if err != nil {
		types.CURRENT_ERR = types.ERR_FUNC
		return "", fmt.Errorf("encrypting: Error reading iv to buffer")
	}

	cipherText, err := aes256Encrypt(clearText, encKey, iv)
	if err != nil {
		return "", err
	}

	macData := append(iv, cipherText...)
	keyMac := hmac.New(sha256.New, macKey)
	keyMac.Write(macData)
	macCode := keyMac.Sum(nil)

	ivB64 := base64.StdEncoding.EncodeToString(iv)
	ciphertextB64 := base64.StdEncoding.EncodeToString(cipherText)
	macCodeB64 := base64.StdEncoding.EncodeToString(macCode)

	encString := "2" + "." + ivB64 + "|" + ciphertextB64 + "|" + macCodeB64
	return encString, nil
}

func parseEncString(encString string) (string, string, string, error) {
	if encString == "" {
		types.CURRENT_ERR = types.ERR_INPUT
		return "", "", "", fmt.Errorf("empty data")
	}
	encStringSplited := strings.Split(encString, ".")
	if len(encStringSplited) < 2 {
		types.CURRENT_ERR = types.ERR_INPUT
		return "", "", "", fmt.Errorf("malformed data: missing \".\", expecting {int}.{iv: base64 string}|{data: base64 string}|{mac: base64 string}")
	}
	dataChunk := encStringSplited[1]

	retDataArr := strings.Split(dataChunk, "|")
	if len(retDataArr) < 3 {
		types.CURRENT_ERR = types.ERR_INPUT
		return "", "", "", fmt.Errorf("parseEncString: missing \"|\", expecting {int}.{iv: base64 string}|{data: base64 string}|{mac: base64 string}")
	}
	iv := retDataArr[0]
	data := retDataArr[1]
	macCode := retDataArr[2]

	return iv, data, macCode, nil
}
