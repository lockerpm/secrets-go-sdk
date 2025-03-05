package locker

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lockerpm/secrets-sdk-go/types"
)

func (locker *Locker) ExportOutput(result interface{}, dataFormat string) error {
	var data []byte
	var err error
	switch dataFormat {
	case strings.ToLower("json"):
		data, err = json.MarshalIndent(result, "", "  ")
		if err != nil {
			types.CURRENT_ERR = types.ERR_FUNC
			return fmt.Errorf("error writing data to file: %w", err)
		}
		locker.OutputPath = filepath.Join(locker.WorkingDir, "output.json")
	default:
		switch resultAsserted := result.(type) {
		case []types.Secret:
			for _, item := range resultAsserted {
				statement := fmt.Sprintf("%s = %s\n", item.Key, item.Value)
				data = append(data, []byte(statement)...)
			}
		case types.Secret:
			statement := fmt.Sprintf("%s = %s\n", resultAsserted.Key, resultAsserted.Value)
			data = append(data, []byte(statement)...)
		case []types.Environment:
			for _, item := range resultAsserted {
				statement := fmt.Sprintf("%s = %s\n", item.Name, item.ExternalURL)
				data = append(data, []byte(statement)...)
			}
		case types.Environment:
			statement := fmt.Sprintf("%s = %s\n", resultAsserted.Name, resultAsserted.ExternalURL)
			data = append(data, []byte(statement)...)
		case types.EncryptedSecResponse:
			switch locker.currentOperation {
			case types.OPERATION_CREATE:
				statement := fmt.Sprintf("creation of secret item with key %s and value %s completed\n", resultAsserted.Key, resultAsserted.Value)
				data = append(data, []byte(statement)...)
			case types.OPERATION_UPDATE:
				statement := fmt.Sprintf("modification of secret item %s completed, current key, value is %s and %s\n", resultAsserted.ID, resultAsserted.Key, resultAsserted.Value)
				data = append(data, []byte(statement)...)
			}
		case types.EncryptedEnvResponse:
			switch locker.currentOperation {
			case types.OPERATION_CREATE:
				statement := fmt.Sprintf("creation of enviroment item with name %s and url %s completed\n", resultAsserted.Name, resultAsserted.ExternalURL)
				data = append(data, []byte(statement)...)
			case types.OPERATION_UPDATE:
				statement := fmt.Sprintf("modification of enviroment item %s completed, current name, url is %s and %s\n", resultAsserted.ID, resultAsserted.Name, resultAsserted.ExternalURL)
				data = append(data, []byte(statement)...)
			}
		}
	}

	err = os.WriteFile(locker.OutputPath, []byte(data), 0640)
	if err != nil {
		types.CURRENT_ERR = types.ERR_PATH
		return fmt.Errorf("error writing data to file: %w", err)
	}
	return nil
}

func unmarshalAny[Struct any](bytes []byte) (*Struct, error) {
	out := new(Struct)
	if err := json.Unmarshal(bytes, out); err != nil {
		types.CURRENT_ERR = types.ERR_FUNC
		return nil, fmt.Errorf("error unmarshalling data: %w", err)
	}
	return out, nil
}
