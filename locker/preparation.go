package locker

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"

	"github.com/lockerpm/secrets-sdk-go/types"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (locker *Locker) prepareData(hash, kind string) error {
	var err error
	var revDate types.RevisionDate
	var fetch bool

	// get revdate, if server date is newer, set fetch to true to force fetch server's data
	revDate, fetch, err = locker.evaluateFetch(hash, kind)
	if err != nil {
		return err
	}

	// if list, evaluate delete date to see if it's necessary to re-sync all data and override fetch
	if hash == "" {
		fetch, err = locker.evaluateDeletedDate()
		if err != nil {
			return err
		}
	}

	if fetch || hash == "" {
		err = locker.fetchDataFromServer(hash, revDate, kind)
		if err != nil {
			return err
		}
	}

	return nil
}

func (locker *Locker) prepareProfile() error {
	var profile types.Profile
	result := locker.dBConn.First(&profile)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			err := locker.fetchDataFromServer("", types.RevisionDate{}, types.FETCH_KIND_PROFILE)
			if err != nil {
				return err
			}

			return nil
		}
		types.CURRENT_ERR = types.ERR_DB
		return fmt.Errorf("error querying profile: %w", result.Error)
	}

	return nil
}

func (locker *Locker) prepareHash(input string) (string, error) {
	var hash string
	var err error

	if input != "" {
		hash, err = locker.getHash(input)
		if err != nil {
			return "", err
		}
	}

	return hash, nil
}

func (locker *Locker) prepareKey() ([]byte, []byte, error) {
	// create database and relevant tables
	err := locker.initDB()
	if err != nil {
		types.CURRENT_ERR = types.ERR_DB
		return nil, nil, err
	}

	accessKey, err := base64.StdEncoding.DecodeString(locker.SecretAccessKey)
	if err != nil {
		types.CURRENT_ERR = types.ERR_INPUT_KEY
		return nil, nil, fmt.Errorf("invalid Secret Access Key string")
	}
	stretchedKey, symMacKey, err := generateKey(accessKey)
	if err != nil {
		return nil, nil, err
	}

	var profile types.Profile
	result := locker.dBConn.First(&profile)
	if result.Error != nil {
		types.CURRENT_ERR = types.ERR_DB
		return nil, nil, fmt.Errorf("error querying profile: %w", result.Error)
	}

	encSymKeyStr := profile.Key

	symKey, macKey, err := getSymKey(encSymKeyStr, stretchedKey, symMacKey)
	if err != nil {
		return nil, nil, err
	}

	return symKey, macKey, nil
}

func (locker *Locker) prepare(input, dataType string) error {
	var err error
	locker.hash, err = locker.prepareHash(input)
	if err != nil {
		return err
	}

	err = locker.prepareData(locker.hash, dataType)
	if err != nil {
		return err
	}

	locker.symKey, locker.macKey, err = locker.prepareKey()
	if err != nil {
		return err
	}

	return nil
}

func (locker *Locker) evaluateFetch(hash, kind string) (types.RevisionDate, bool, error) {
	var isFetchFromServer bool = locker.Fetch

	// evaluate last revision date
	revDate, isDateNewer, err := locker.evaluateDate(kind)
	if err != nil {
		return types.RevisionDate{}, false, err
	}

	// if table count is mismatched with server, force RevDate 0 and fetch
	var localCount int64
	var result *gorm.DB
	switch kind {
	case types.FETCH_KIND_SEC, types.FETCH_KIND_RUN:
		if hash != "" {
			result = locker.dBConn.Model(&types.Secret{}).Where("secret_hash = ?", hash).Count(&localCount)
		} else {
			result = locker.dBConn.Model(&types.Secret{}).Count(&localCount)
		}
		if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return types.RevisionDate{}, false, result.Error
		}
	case types.FETCH_KIND_ENV:
		if hash != "" {
			result = locker.dBConn.Model(&types.Environment{}).Where("hash = ?", hash).Count(&localCount)
		} else {
			result = locker.dBConn.Model(&types.Environment{}).Count(&localCount)
		}
		if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return types.RevisionDate{}, false, result.Error
		}
	}

	serverCount, err := locker.fetchCount(kind)
	if err != nil {
		return types.RevisionDate{}, false, err
	}

	if serverCount != localCount {
		return types.RevisionDate{}, true, nil
	}

	// if newer date from server, return true to force fetch
	if isDateNewer {
		return revDate, true, nil
	}

	if isFetchFromServer {
		return revDate, true, nil
	}

	return revDate, false, nil
}

func (locker *Locker) evaluateDate(kind string) (types.RevisionDate, bool, error) {
	// fetch latest revision date from server
	revDateServer, local, err := locker.fetchRevisionDate(kind)
	if err != nil {
		return types.RevisionDate{}, false, err
	}

	// if revDateServer is from local, don't bother checking further
	if local {
		locker.SetGettingFromLocal(true)
		return revDateServer, false, nil
	}

	// otherwise, query and evaluate
	var revDateLocal types.RevisionDate
	result := locker.dBConn.First(&revDateLocal)
	if result.Error != nil {
		// if NoRows, insert, then return true
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			err := locker.upsertRevisionDate(revDateServer)
			if err != nil {
				return types.RevisionDate{}, false, err
			}

			return revDateServer, true, nil
		}
		return types.RevisionDate{}, false, err
	}

	// if server date > local date or last call = 0, update db and return true
	if revDateServer.RevisionDate > revDateLocal.RevisionDate || revDateLocal.LastCallSec == 0 || revDateLocal.LastCallEnv == 0 {
		// update rev date in local
		err := locker.upsertRevisionDate(revDateServer)
		if err != nil {
			return types.RevisionDate{}, false, err
		}
		return revDateLocal, true, nil
	}

	// otherwise, use local date and return false
	return revDateLocal, false, nil
}

func (locker *Locker) evaluateDeletedDate() (bool, error) {
	delDateObj, err := locker.queryDeletionDate()
	if err != nil {
		return false, nil
	}

	localDelDate := delDateObj.DeletionDate

	dataEndpoint := locker.APIBase + "/v1/sync/deleted_item_date"
	// skip error handling (this is explicitly by design) (for now)
	resBody, err := locker.httpActionIn(dataEndpoint)
	if err != nil {
		return false, nil
	}

	fetchedDelDate, err := strconv.ParseFloat(string(resBody), 64)
	if err != nil {
		return false, fmt.Errorf("error parsing deleted date: %w", err)
	}

	delDateObj.DeletionDate = fetchedDelDate

	result := locker.dBConn.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&delDateObj)

	if result.Error != nil {
		return false, result.Error
	}

	// if server's deleted date is greater, drop secret and env tables, then force fetch
	if fetchedDelDate > localDelDate {
		result := locker.dBConn.Where("TRUE").Delete(&types.Secret{})
		if result.Error != nil {
			return false, result.Error
		}

		result = locker.dBConn.Where("TRUE").Delete(&types.Environment{})
		if result.Error != nil {
			return false, result.Error
		}

		return true, nil
	}

	return false, nil
}
