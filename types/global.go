package types

import (
	"encoding/json"
	"fmt"
)

type ConfigFile struct {
	AccessKeyId     string `mapstructure:"access_key_id"`
	AccessKeySecret string `mapstructure:"secret_access_key"`
}

func (configData *ConfigFile) NewConfig(fileData []byte) error {
	err := json.Unmarshal(fileData, &configData)
	if err != nil {
		CURRENT_ERR = ERR_INPUT
		return fmt.Errorf("error unmarshalling config data to struct: %w", err)
	}

	return nil
}

var VERSION string
var DEFAULT_CLIENT = "Locker Secret CLI - version " + VERSION

var CURRENT_ERR string
var CurrentOperation string
