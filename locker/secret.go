package locker

import (
	"encoding/json"
	"errors"
	"fmt"
	"sdk-test/types"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type InputSecData struct {
	Key   *string `json:"key,omitempty"`
	Hash  string  `json:"hash,omitempty"`
	Value *string `json:"value,omitempty"`
	Desc  *string `json:"description,omitempty"`
	EnvID *string `json:"environment_id,omitempty"`
	Env   *string `json:"environment_name,omitempty"`
}

func (locker *Locker) GetSecret(key string, env *string) (types.Secret, error) {
	err := locker.prepare(key, types.FETCH_KIND_SEC)
	if err != nil {
		return types.Secret{}, err
	}

	var envHash string
	var result *gorm.DB
	var secObj types.Secret

	if env != nil {
		envHash, err = locker.getHash(*env)
		if err != nil {
			return types.Secret{}, err
		}
	}

	if locker.emptyFetch {
		if env != nil {
			result = locker.dBConn.Where("secret_hash = ? AND environment_hash = ?", locker.hash, envHash).Delete(&types.Secret{})
		} else {
			result = locker.dBConn.Where("secret_hash = ? AND environment_hash is NULL", locker.hash).Delete(&types.Secret{})
		}

		if result.Error != nil {
			return types.Secret{}, fmt.Errorf("error deleting secret: %v", result.Error)
		}
	}

	if env != nil {
		result = locker.dBConn.Where("secret_hash = ? AND environment_hash = ?", locker.hash, envHash).First(&secObj)
	} else {
		result = locker.dBConn.Where("secret_hash = ? AND environment_hash is NULL", locker.hash).First(&secObj)
	}
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return types.Secret{}, fmt.Errorf("error querying secret: %v", result.Error)
	}

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		err = locker.fetchDataFromServer(locker.hash, types.RevisionDate{}, types.FETCH_KIND_SEC)
		if err != nil {
			return types.Secret{}, err
		}

		if env != nil {
			result = locker.dBConn.Where("secret_hash = ? AND environment_hash = ?", locker.hash, envHash).First(&secObj)
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				result = locker.dBConn.Where("secret_hash = ? AND environment_hash is NULL", locker.hash).First(&secObj)
			}
		} else {
			result = locker.dBConn.Where("secret_hash = ? AND environment_hash is NULL", locker.hash).First(&secObj)
		}

		if result.Error != nil {
			if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return types.Secret{}, fmt.Errorf("error querying secret: %v", result.Error)
			}
			return types.Secret{}, fmt.Errorf("no secret found with provided name and env")
		}
	}

	// decrypt data
	err = dataDecryption(&secObj, locker.symKey, locker.macKey)
	if err != nil {
		return types.Secret{}, err
	}

	// err = locker.processOutputDecryption(secObj, types.FETCH_KIND_SEC, locker.hash)
	// if err != nil {
	// 	return types.Secret{}, err
	// }

	return secObj, nil
}

func (locker *Locker) ListSecret(env *string) ([]types.Secret, error) {
	err := locker.prepare("", types.FETCH_KIND_SEC)
	if err != nil {
		return []types.Secret{}, err
	}

	var envHash string
	var result *gorm.DB

	if env != nil {
		envHash, err = locker.getHash(*env)
		if err != nil {
			return []types.Secret{}, err
		}
	}

	if locker.emptyFetch {
		if env != nil {
			result = locker.dBConn.Where("environment_hash = ?", envHash).Delete(&types.Secret{})
		} else {
			result = locker.dBConn.Delete(&types.Secret{})
		}
		if result.Error != nil {
			return []types.Secret{}, fmt.Errorf("error deleting secret: %v", result.Error)
		}
	}

	var secObjs []types.Secret
	if env != nil {
		result = locker.dBConn.Where("environment_hash = ?", envHash).Find(&secObjs)
	} else {
		result = locker.dBConn.Find(&secObjs)
	}

	if result.Error != nil && result.RowsAffected != 0 {
		return []types.Secret{}, fmt.Errorf("error querying secret: %v", result.Error)
	}

	// last ditch effort
	if result.RowsAffected == 0 {
		err = locker.fetchDataFromServer("", types.RevisionDate{}, types.FETCH_KIND_SEC)
		if err != nil {
			return []types.Secret{}, err
		}

		if env != nil {
			result = locker.dBConn.Where("environment_hash = ?", envHash).Find(&secObjs)
		} else {
			result = locker.dBConn.Find(&secObjs)
		}

		if result.Error != nil {
			if result.RowsAffected != 0 {
				return []types.Secret{}, fmt.Errorf("error querying secret: %v", result.Error)
			}
			return []types.Secret{}, fmt.Errorf("no secret found with provided name and env")
		}
	}

	for i := range secObjs {
		err = dataDecryption(&secObjs[i], locker.symKey, locker.macKey)
		if err != nil {
			return []types.Secret{}, err
		}
	}

	if locker.Export {
	}

	return secObjs, nil
}

func (locker *Locker) CreateSecret(input *InputSecData) (types.EncryptedSecResponse, error) {
	locker.currentOperation = types.OPERATION_CREATE
	if input == nil || input.Key == nil || input.Value == nil {
		return types.EncryptedSecResponse{}, fmt.Errorf("secret's name and value must not be empty")
	}

	err := locker.prepare(*input.Key, types.FETCH_KIND_SEC)
	if err != nil {
		return types.EncryptedSecResponse{}, err
	}
	tmpHash := locker.hash

	if input.Env != nil {
		getEnvironmentResult, err := locker.GetEnvironment(*input.Env)
		if err != nil {
			return types.EncryptedSecResponse{}, err
		}
		input.EnvID = &getEnvironmentResult.ID
	}

	keyPreEnc := *input.Key
	var envPreEnc *string
	if input.Env != nil {
		envPreEnc = input.Env
	}

	err = dataEncryption(input, locker.symKey, locker.macKey)
	if err != nil {
		return types.EncryptedSecResponse{}, err
	}
	input.Hash = tmpHash

	jsonBody, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		return types.EncryptedSecResponse{}, err
	}

	createResult, err := createItem[types.EncryptedSecResponse](locker, types.FETCH_KIND_SEC, jsonBody)
	if err != nil {
		return types.EncryptedSecResponse{}, err
	}

	if (createResult != &types.EncryptedSecResponse{}) {
		dataToInsert := types.Secret{
			Object:          "secret",
			ID:              createResult.ID,
			CreationDate:    createResult.CreationDate,
			RevisionDate:    createResult.RevisionDate,
			UpdatedDate:     createResult.UpdatedDate,
			DeletedDate:     createResult.DeletedDate,
			LastUseDate:     createResult.LastUseDate,
			ProjectID:       createResult.ProjectID,
			EnvironmentID:   createResult.EnvironmentID,
			EnvironmentName: createResult.EnvironmentName,
			Key:             createResult.Key,
			SecretHash:      createResult.SecretHash,
			EnvironmentHash: createResult.EnvironmentHash,
			Value:           createResult.Value,
			Description:     createResult.Description,
		}

		// handle special case (secret_hash, NULL) being able to bypass unique rule
		if dataToInsert.EnvironmentHash == nil {
			getResult, err := locker.GetSecret(keyPreEnc, envPreEnc)
			if err != nil && !strings.Contains(err.Error(), "no secret found with provided name and env") {
				return types.EncryptedSecResponse{}, err
			}

			if getResult.Value != dataToInsert.Value {
				delResult := locker.dBConn.Where("secret_hash = ? AND environment_hash is NULL", getResult.SecretHash).Delete(&types.Secret{})
				if delResult.Error != nil {
					return types.EncryptedSecResponse{}, err
				}
			}
		}

		result := locker.dBConn.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&dataToInsert)
		if result.Error != nil {
			return types.EncryptedSecResponse{}, err
		}

		err := dataDecryption(createResult, locker.symKey, locker.macKey)
		if err != nil {
			return types.EncryptedSecResponse{}, err
		}

	}

	return *createResult, nil
}

func (locker *Locker) UpdateSecret(key string, env *string, input *InputSecData) (types.EncryptedSecResponse, error) {
	locker.currentOperation = types.OPERATION_UPDATE
	if input == nil {
		return types.EncryptedSecResponse{}, fmt.Errorf("there must be atleast one field in update data")
	}

	err := locker.prepare(key, types.FETCH_KIND_SEC)
	if err != nil {
		return types.EncryptedSecResponse{}, err
	}

	getSecretResult, err := locker.GetSecret(key, env)
	if err != nil {
		return types.EncryptedSecResponse{}, err
	}

	if input.Env != nil {
		if *input.Env != "" {
			getEnvironmentResult, err := locker.GetEnvironment(*input.Env)
			if err != nil {
				return types.EncryptedSecResponse{}, err
			}
			input.EnvID = &getEnvironmentResult.ID
		} else {
			tmp := ""
			input.EnvID = &tmp
		}
	}

	var keyPreEnc string
	if input.Key != nil {
		keyPreEnc = *input.Key
	} else {
		keyPreEnc = key
	}

	err = dataEncryption(input, locker.symKey, locker.macKey)
	if err != nil {
		return types.EncryptedSecResponse{}, err
	}

	input.Hash, err = locker.getHash(keyPreEnc)
	if err != nil {
		return types.EncryptedSecResponse{}, err
	}

	jsonBody, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		return types.EncryptedSecResponse{}, err
	}

	editResult, err := editItem[types.EncryptedSecResponse](locker, types.FETCH_KIND_SEC, getSecretResult.ID, jsonBody)
	if err != nil {
		return types.EncryptedSecResponse{}, err
	}

	if (editResult != &types.EncryptedSecResponse{}) {
		dataToUpdate := types.Secret{
			Object:          editResult.Object,
			ID:              editResult.ID,
			CreationDate:    editResult.CreationDate,
			RevisionDate:    editResult.RevisionDate,
			UpdatedDate:     editResult.UpdatedDate,
			DeletedDate:     editResult.DeletedDate,
			LastUseDate:     editResult.LastUseDate,
			ProjectID:       editResult.ProjectID,
			EnvironmentID:   editResult.EnvironmentID,
			EnvironmentName: editResult.EnvironmentName,
			Key:             editResult.Key,
			SecretHash:      editResult.SecretHash,
			EnvironmentHash: editResult.EnvironmentHash,
			Value:           editResult.Value,
			Description:     editResult.Description,
		}

		result := locker.dBConn.Save(&dataToUpdate)
		if result.Error != nil {
			return types.EncryptedSecResponse{}, err
		}

		err := dataDecryption(editResult, locker.symKey, locker.macKey)
		if err != nil {
			return types.EncryptedSecResponse{}, err
		}
	}

	return *editResult, nil
}
