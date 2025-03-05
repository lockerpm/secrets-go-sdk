package locker

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/lockerpm/secrets-sdk-go/types"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type InputEnvData struct {
	Name *string `json:"name,omitempty"`
	Hash string  `json:"hash,omitempty"`
	Url  *string `json:"external_url,omitempty"`
	Desc *string `json:"description,omitempty"`
}

func (locker *Locker) GetEnvironment(name string) (types.Environment, error) {
	err := locker.prepare(name, types.FETCH_KIND_ENV)
	if err != nil {
		return types.Environment{}, err
	}

	if locker.emptyFetch {
		result := locker.dBConn.Where("hash = ?", locker.hash).Delete(&types.Environment{})
		if result.Error != nil {
			types.CURRENT_ERR = types.ERR_DB
			return types.Environment{}, fmt.Errorf("error deleting environment: %v", result.Error)
		}
	}

	var envObj types.Environment
	result := locker.dBConn.Where("hash = ?", locker.hash).First(&envObj)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		types.CURRENT_ERR = types.ERR_DB
		return types.Environment{}, fmt.Errorf("error querying environment: %v", result.Error)
	}

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		err := locker.fetchDataFromServer(locker.hash, types.RevisionDate{}, types.FETCH_KIND_ENV)
		if err != nil {
			return types.Environment{}, err
		}

		result := locker.dBConn.Where("hash = ?", locker.hash).First(&envObj)
		if result.Error != nil {
			if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
				types.CURRENT_ERR = types.ERR_DB
				return types.Environment{}, fmt.Errorf("error querying secret: %v", result.Error)
			} else {
				types.CURRENT_ERR = types.ERR_NOT_FOUND
				return types.Environment{}, fmt.Errorf("no environment found with provided name")
			}
		}
	}

	// decrypt data
	err = dataDecryption(&envObj, locker.symKey, locker.macKey)
	if err != nil {
		return types.Environment{}, err
	}

	// err = locker.processOutputDecryption(envObj, types.FETCH_KIND_ENV, locker.hash)
	// if err != nil {
	// 	return types.Environment{}, err
	// }

	return envObj, nil
}

func (locker *Locker) ListEnvironment() ([]types.Environment, error) {
	err := locker.prepare("", types.FETCH_KIND_ENV)
	if err != nil {
		return []types.Environment{}, err
	}

	var envObjs []types.Environment
	result := locker.dBConn.Find(&envObjs)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		types.CURRENT_ERR = types.ERR_DB
		return []types.Environment{}, fmt.Errorf("error querying environment: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		err = locker.fetchDataFromServer("", types.RevisionDate{}, types.FETCH_KIND_ENV)
		if err != nil {
			return []types.Environment{}, err
		}

		result = locker.dBConn.Find(&envObjs)

		if result.Error != nil {
			if result.RowsAffected != 0 {
				types.CURRENT_ERR = types.ERR_DB
				return []types.Environment{}, fmt.Errorf("error querying environment: %v", result.Error)
			}
			types.CURRENT_ERR = types.ERR_NOT_FOUND
			return []types.Environment{}, fmt.Errorf("no environment found")
		}
	}

	for i := range envObjs {
		err = dataDecryption(&envObjs[i], locker.symKey, locker.macKey)
		if err != nil {
			return []types.Environment{}, err
		}
	}

	// err = locker.processOutputDecryption(envObjs, types.FETCH_KIND_ENV, "")
	// if err != nil {
	// 	return []types.Environment{}, err
	// }

	return envObjs, nil
}

func (locker *Locker) CreateEnvironment(input *InputEnvData) (types.EncryptedEnvResponse, error) {
	locker.currentOperation = types.OPERATION_CREATE
	if input.Name == nil {
		return types.EncryptedEnvResponse{}, fmt.Errorf("environment's name must not be empty")
	}

	err := locker.prepare(*input.Name, types.FETCH_KIND_ENV)
	if err != nil {
		return types.EncryptedEnvResponse{}, err
	}

	// namePreEnc := *input.Name
	err = dataEncryption(input, locker.symKey, locker.macKey)
	if err != nil {
		return types.EncryptedEnvResponse{}, err
	}
	input.Hash = locker.hash

	jsonBody, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		return types.EncryptedEnvResponse{}, err
	}

	createResult, err := createItem[types.EncryptedEnvResponse](locker, types.FETCH_KIND_ENV, jsonBody)
	if err != nil {
		return types.EncryptedEnvResponse{}, err
	}

	if (createResult != &types.EncryptedEnvResponse{}) {
		dataToInsert := types.Environment{
			Object:       createResult.Object,
			ID:           createResult.ID,
			Name:         createResult.Name,
			Hash:         createResult.Hash,
			ExternalURL:  createResult.ExternalURL,
			Description:  createResult.Description,
			CreationDate: createResult.CreationDate,
			RevisionDate: createResult.RevisionDate,
			UpdatedDate:  &createResult.UpdatedDate,
			ProjectID:    createResult.ProjectID,
		}

		result := locker.dBConn.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "hash"}},
			UpdateAll: true,
		}).Create(&dataToInsert)
		if result.Error != nil {
			return types.EncryptedEnvResponse{}, err
		}

		err := dataDecryption(createResult, locker.symKey, locker.macKey)
		if err != nil {
			return types.EncryptedEnvResponse{}, err
		}

	}

	return *createResult, nil
}

func (locker *Locker) UpdateEnvironment(name string, input *InputEnvData) (types.EncryptedEnvResponse, error) {
	locker.currentOperation = types.OPERATION_UPDATE
	if input == nil {
		return types.EncryptedEnvResponse{}, fmt.Errorf("there must be atleast one field in update data")
	}

	err := locker.prepare(name, types.FETCH_KIND_ENV)
	if err != nil {
		return types.EncryptedEnvResponse{}, err
	}

	envID := ""
	getResult, err := locker.GetEnvironment(name)
	if err != nil {
		return types.EncryptedEnvResponse{}, err
	}
	envID = getResult.ID

	var namePreEnc string
	if input.Name != nil {
		namePreEnc = *input.Name
	} else {
		namePreEnc = name
	}

	err = dataEncryption(input, locker.symKey, locker.macKey)
	if err != nil {
		return types.EncryptedEnvResponse{}, err
	}

	input.Hash, err = locker.getHash(namePreEnc)
	if err != nil {
		return types.EncryptedEnvResponse{}, err
	}

	jsonBody, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		return types.EncryptedEnvResponse{}, err
	}

	editResult, err := editItem[types.EncryptedEnvResponse](locker, types.FETCH_KIND_ENV, envID, jsonBody)
	if err != nil {
		return types.EncryptedEnvResponse{}, err
	}

	if (editResult != &types.EncryptedEnvResponse{}) {
		dataToUpdate := types.Environment{
			Object:       editResult.Object,
			ID:           editResult.ID,
			Name:         editResult.Name,
			Hash:         editResult.Hash,
			ExternalURL:  editResult.ExternalURL,
			Description:  editResult.Description,
			CreationDate: editResult.CreationDate,
			RevisionDate: editResult.RevisionDate,
			UpdatedDate:  &editResult.UpdatedDate,
			ProjectID:    editResult.ProjectID,
		}

		result := locker.dBConn.Save(&dataToUpdate)
		if result.Error != nil {
			return types.EncryptedEnvResponse{}, err
		}

		err := dataDecryption(editResult, locker.symKey, locker.macKey)
		if err != nil {
			return types.EncryptedEnvResponse{}, err
		}

	}

	return *editResult, nil
}
