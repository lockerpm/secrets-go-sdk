package locker

import (
	"fmt"
	"reflect"
	"regexp"
	"sdk-test/types"
	"strings"
)

var encryptedStringPattern = regexp.MustCompile(`2\..*\|.*\|.*`)

func dataDecryption(rawData interface{}, symKey []byte, macKey []byte) error {
	value := reflect.ValueOf(rawData)
	if value.Kind() != reflect.Pointer && value.Kind() != reflect.Interface {
		types.CURRENT_ERR = types.ERR_FUNC
		return fmt.Errorf("input is not pointer or interface")
	}
	value = value.Elem()

	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		if !field.CanInterface() {
			continue
		}

		if field.Kind() == reflect.Struct {
			err := dataDecryptionNested(field, symKey, macKey)
			if err != nil {
				return err
			}
			continue
		}

		if field.Kind() == reflect.Pointer {
			field = field.Elem()
		}

		if field.Kind() == reflect.String {
			str := field.Interface().(string)
			if len(str) == 0 || !encryptedStringPattern.MatchString(str) {
				continue
			}
			var err error
			str, err = aes256DecryptToString(str, symKey, macKey)
			if err != nil {
				return err
			}
			field.SetString(str)
		}
	}
	return nil
}

// value is expected to be of Struct kind
func dataDecryptionNested(value reflect.Value, symKey []byte, macKey []byte) error {
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		if !field.CanInterface() {
			continue
		}

		if field.Kind() == reflect.Struct {
			err := dataDecryptionNested(field, symKey, macKey)
			if err != nil {
				return err
			}

			continue
		}

		if field.Kind() == reflect.String {
			str := field.Interface().(string)
			if len(str) == 0 || !encryptedStringPattern.MatchString(str) {
				continue
			}
			var err error
			str, err = aes256DecryptToString(str, symKey, macKey)
			if err != nil {
				return err
			}
			field.SetString(str)
		}
	}
	return nil
}

func dataEncryption(rawData interface{}, symKey []byte, macKey []byte) error {
	value := reflect.ValueOf(rawData)
	if value.Kind() != reflect.Pointer && value.Kind() != reflect.Interface {
		types.CURRENT_ERR = types.ERR_FUNC
		return fmt.Errorf("input is not pointer or interface")
	}
	value = value.Elem()

	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		if !field.CanInterface() {
			continue
		}

		if field.Kind() == reflect.Struct {
			err := dataEncryptionNested(field, symKey, macKey)
			if err != nil {
				return err
			}

			continue
		}

		if strings.EqualFold(value.Type().Field(i).Name, "UpdateEnv") || strings.EqualFold(value.Type().Field(i).Name, "EnvID") {
			continue
		}

		if field.Kind() == reflect.String {
			str := field.Interface().(string)
			var err error
			str, err = aes256EncryptToString([]byte(str), symKey, macKey)
			if err != nil {
				return err
			}
			field.SetString(str)
		}

		if field.Kind() == reflect.Pointer {
			str := field.Interface().(*string)
			if str == nil && types.CurrentOperation == types.OPERATION_CREATE {
				tmp := ""
				str = &tmp
			}
			var err error
			if str != nil {
				*str, err = aes256EncryptToString([]byte(*str), symKey, macKey)
				if err != nil {
					return err
				}
			}
			field.Set(reflect.ValueOf(str))
		}
	}
	return nil
}

// value is expected to be of Struct kind
func dataEncryptionNested(value reflect.Value, symKey []byte, macKey []byte) error {
	for i := 0; i < value.NumField(); i++ {

		field := value.Field(i)
		if !field.CanInterface() {
			continue
		}

		if field.Kind() == reflect.Struct {
			err := dataEncryptionNested(field, symKey, macKey)
			if err != nil {
				return err
			}

			continue
		}

		if field.Kind() == reflect.String {
			str := field.Interface().(string)
			var err error
			str, err = aes256EncryptToString([]byte(str), symKey, macKey)
			if err != nil {
				return err
			}
			field.SetString(str)
		}
	}
	return nil
}
