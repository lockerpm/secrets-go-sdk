package locker

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/lockerpm/secrets-sdk-go/types"

	"strconv"
)

func (locker *Locker) getHash(plainKey string) (string, error) {
	var profile types.Profile
	result := locker.dBConn.First(&profile)
	if result.Error != nil {
		types.CURRENT_ERR = types.ERR_DB
		return "", fmt.Errorf("error querying profile: %w", result.Error)
	}

	salt := profile.ProjectID

	projectIDBytes := []byte(strconv.Itoa(salt))
	plainKeyBytes := []byte(plainKey)

	dataBytes := append(projectIDBytes, plainKeyBytes...)
	hasher := sha256.New()
	hasher.Write([]byte(dataBytes))

	return base64.RawURLEncoding.EncodeToString(hasher.Sum(nil)), nil
}
