package test

import (
	"log"
	"os"
	"sdk-test/locker"
	"sdk-test/types"
	"strings"
	"testing"

	"github.com/joho/godotenv"
)

// expected initial state:
// 1 secret (key: secret key 1, value: secret value 1, environment: ALL)
// 1 environment (name: env name 1, url: env url 1)

var lockerClient locker.Locker

const ERR_INDUCED = "invalid value"

const INIT_SEC_KEY = "init secret key"
const INIT_SEC_VAL = "init secret value"
const INIT_ENV_NAME = "init env name"
const INIT_ENV_URL = "init env url"

const CREATE_ENV_NAME = "created env name"
const CREATE_ENV_URL = "created env url"
const UPDATE_ENV_NAME = "updated env name"
const UPDATE_ENV_URL = "updated env url"

const CREATE_SEC_KEY = "created secret key"
const CREATE_SEC_VAL = "created secret value"
const UPDATE_SEC_KEY = "updated secret key"
const UPDATE_SEC_VAL = "updated secret value"

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	accessKeyID := os.Getenv("ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("SECRET_ACCESS_KEY")

	lockerClient.NewLockerClient()
	lockerClient.SetAccessKeyID(accessKeyID)
	lockerClient.SetSecretAccessKey(secretAccessKey)
	lockerClient.SetFetch(true)
}

// LIST/GET TEST
func TestInititalSecList(t *testing.T) {
	// time.Sleep(3 * time.Second)
	initialSecList, err := lockerClient.ListSecret(nil)
	if err != nil {
		t.Fatalf("list secret broke, error: %v", err)
	}
	if len(initialSecList) != 1 {
		t.Fatalf("list secret broke, expecting 1 secret")
	}
	for _, secret := range initialSecList {
		if secret.Key != INIT_SEC_KEY {
			t.Fatalf("list secret decryption broke, expecting key \"%s\", getting \"%s\" instead", INIT_SEC_KEY, secret.Key)
		}
		if secret.Value != INIT_SEC_VAL {
			t.Fatalf("list secret decryption broke, expecting value \"%s\", getting \"%s\" instead", INIT_SEC_VAL, secret.Value)
		}
	}
}

func TestInititalEnvList(t *testing.T) {
	// time.Sleep(3 * time.Second)
	initialEnvList, err := lockerClient.ListEnvironment()
	if err != nil {
		t.Fatalf("list env broke, error: %v", err)
	}
	if len(initialEnvList) != 1 {
		t.Fatalf("list env broke, expecting 1 environment")
	}
	for _, env := range initialEnvList {
		if env.Name != INIT_ENV_NAME {
			t.Fatalf("list env decryption broke, expecting name \"%s\", getting %s instead", INIT_ENV_NAME, env.Name)
		}
		if env.ExternalURL != INIT_ENV_URL {
			t.Fatalf("list env decryption broke, expecting url \"%s\", getting %s instead", INIT_ENV_URL, env.ExternalURL)
		}
	}
}

func TestInititalSecGet(t *testing.T) {
	// time.Sleep(3 * time.Second)
	secret, err := lockerClient.GetSecret(INIT_SEC_KEY, nil)
	if err != nil {
		t.Fatalf("get secret broke, error: %v", err)
	}
	if (secret == types.Secret{}) {
		t.Fatalf("get secret broke, expecting non nil result")
	}
	if secret.Key != INIT_SEC_KEY {
		t.Fatalf("get secret decryption broke, expecting key \"%s\", getting \"%s\" instead", INIT_SEC_KEY, secret.Key)
	}
	if secret.Value != INIT_SEC_VAL {
		t.Fatalf("get secret decryption broke, expecting value \"%s\", getting \"%s\" instead", INIT_SEC_VAL, secret.Value)
	}
	if secret.EnvironmentName != nil {
		t.Fatalf("get secret decryption broke, expecting environment \"nil\", getting \"%+v\" instead", secret.EnvironmentName)
	}
}

func TestInititalInvalidSecGet(t *testing.T) {
	// time.Sleep(3 * time.Second)
	_, err := lockerClient.GetSecret(ERR_INDUCED, nil)
	if err != nil {
		if err.Error() != "no secret found with provided name and env" {
			t.Fatalf("get secret broke, expecting no secret found error, getting error: %v", err)
		}
	} else {
		t.Fatalf("expecting no secret found error")
	}
}

func TestInititalEnvGet(t *testing.T) {
	// time.Sleep(3 * time.Second)
	env, err := lockerClient.GetEnvironment(INIT_ENV_NAME)
	if err != nil {
		t.Fatalf("get env broke, error: %v", err)
	}
	if (env == types.Environment{}) {
		t.Fatalf("get env broke, expecting non nil result")
	}
	if env.Name != INIT_ENV_NAME {
		t.Fatalf("get env decryption broke, expecting name \"%s\", getting %s instead", INIT_ENV_NAME, env.Name)
	}
	if env.ExternalURL != INIT_ENV_URL {
		t.Fatalf("get env decryption broke, expecting url \"%s\", getting %s instead", INIT_ENV_URL, env.ExternalURL)
	}
}

func TestInititalInvalidEnvGet(t *testing.T) {
	// time.Sleep(3 * time.Second)
	_, err := lockerClient.GetEnvironment(ERR_INDUCED)
	if err != nil {
		if err.Error() != "no environment found with provided name" {
			t.Fatalf("get env broke, expecting no environment error, error: %v", err)
		}
	} else {
		t.Fatalf("expecting no environment found error")
	}
}

// ENV TEST
// Current state
// Secret: no change
// Env: (INIT_ENV_NAME, INIT_ENV_URL)
func TestCreateEnv(t *testing.T) {
	// time.Sleep(3 * time.Second)
	name := CREATE_ENV_NAME
	url := CREATE_ENV_URL
	input := locker.InputEnvData{
		Name: &name,
		Url:  &url,
	}
	resp, err := lockerClient.CreateEnvironment(&input)
	if err != nil {
		t.Fatalf("create env broke, error: %v", err)
	}
	if resp.Name != CREATE_ENV_NAME {
		t.Fatalf("create env decryption broke, expecting name \"%s\", getting %s instead", INIT_ENV_NAME, resp.Name)
	}
	if resp.ExternalURL != CREATE_ENV_URL {
		t.Fatalf("create env decryption broke, expecting url \"%s\", getting %s instead", INIT_ENV_URL, resp.ExternalURL)
	}
}

// Current state
// Secret: no change
// Env: (INIT_ENV_NAME, INIT_ENV_URL), (CREATE_ENV_NAME, CREATE_ENV_URL)
func TestCreateEnvDup(t *testing.T) {
	// time.Sleep(3 * time.Second)
	name := CREATE_ENV_NAME
	url := CREATE_ENV_URL
	input := locker.InputEnvData{
		Name: &name,
		Url:  &url,
	}
	_, err := lockerClient.CreateEnvironment(&input)
	if err != nil {
		if !strings.Contains(err.Error(), "This environment hash already exists") {
			t.Fatalf("create env broke, error: %v", err)
		}
	} else {
		t.Fatalf("expecting duplicated error")
	}
}

// Current state
// Secret: no change
// Env: (INIT_ENV_NAME, INIT_ENV_URL), (CREATE_ENV_NAME, CREATE_ENV_URL)
func TestUpdateEnv(t *testing.T) {
	// time.Sleep(3 * time.Second)
	name := UPDATE_ENV_NAME
	url := UPDATE_ENV_URL
	input := locker.InputEnvData{
		Name: &name,
		Url:  &url,
	}
	resp, err := lockerClient.UpdateEnvironment(CREATE_ENV_NAME, &input)
	if err != nil {
		t.Fatalf("update env broke, error: %v", err)
	}
	if resp.Name != UPDATE_ENV_NAME {
		t.Fatalf("update env decryption broke, expecting name \"%s\", getting %s instead", UPDATE_ENV_NAME, resp.Name)
	}
	if resp.ExternalURL != UPDATE_ENV_URL {
		t.Fatalf("update env decryption broke, expecting url \"%s\", getting %s instead", UPDATE_ENV_URL, resp.ExternalURL)
	}
}

// Current state
// Secret: no change
// Env: (INIT_ENV_NAME, INIT_ENV_URL), (UPDATE_ENV_NAME, UPDATE_ENV_URL)
func TestUpdateEnvDuplicate(t *testing.T) {
	// time.Sleep(3 * time.Second)
	name := UPDATE_ENV_NAME
	url := UPDATE_ENV_URL
	input := locker.InputEnvData{
		Name: &name,
		Url:  &url,
	}
	_, err := lockerClient.UpdateEnvironment(INIT_ENV_NAME, &input)
	if err != nil {
		if !strings.Contains(err.Error(), "The environment hash already existed") {
			t.Fatalf("update env broke, error: %v", err)
		}
	} else {
		t.Fatalf("expecting duplicated error")
	}
}

// SECRET TEST
// Current state
// Secret: no change
// Env: (INIT_ENV_NAME, INIT_ENV_URL), (UPDATE_ENV_NAME, UPDATE_ENV_URL)
func TestCreateSecretAll(t *testing.T) {
	// time.Sleep(3 * time.Second)
	key := CREATE_SEC_KEY
	value := CREATE_SEC_VAL
	input := locker.InputSecData{
		Key:   &key,
		Value: &value,
	}
	resp, err := lockerClient.CreateSecret(&input)
	if err != nil {
		t.Fatalf("create secret broke, error: %v", err)
	}
	if resp.Key != CREATE_SEC_KEY {
		t.Fatalf("create secret decryption broke, expecting key \"%s\", getting %s instead", CREATE_SEC_KEY, resp.Key)
	}
	if resp.Value != CREATE_SEC_VAL {
		t.Fatalf("create secret decryption broke, expecting value \"%s\", getting %s instead", CREATE_SEC_VAL, resp.Value)
	}
	if resp.EnvironmentName != nil {
		t.Fatalf("create secret decryption broke, expecting env nil, getting %s instead", *resp.EnvironmentName)
	}
}

// Current state
// Secret: (INIT_SEC_KEY, INIT_SEC_VAL, ALL), (CREATE_SEC_KEY, CREATE_SEC_VAL, ALL)
// Env: (INIT_ENV_NAME, INIT_ENV_URL), (UPDATE_ENV_NAME, UPDATE_ENV_URL)
func TestCreateSecretAllAgain(t *testing.T) {
	// time.Sleep(3 * time.Second)
	key := CREATE_SEC_KEY
	value := CREATE_SEC_VAL
	input := locker.InputSecData{
		Key:   &key,
		Value: &value,
	}
	_, err := lockerClient.CreateSecret(&input)
	if err != nil {
		if !strings.Contains(err.Error(), "This secret hash already exists") {
			t.Fatalf("create secret broke, error: %v", err)
		}
	} else {
		t.Fatalf("expecting duplicated error")
	}
}

// Current state
// Secret: (INIT_SEC_KEY, INIT_SEC_VAL, ALL), (CREATE_SEC_KEY, CREATE_SEC_VAL, ALL)
// Env: (INIT_ENV_NAME, INIT_ENV_URL), (UPDATE_ENV_NAME, UPDATE_ENV_URL)
func TestCreateSecretWithEnv(t *testing.T) {
	// time.Sleep(3 * time.Second)
	key := CREATE_SEC_KEY
	value := CREATE_SEC_VAL
	env := INIT_ENV_NAME
	input := locker.InputSecData{
		Key:   &key,
		Value: &value,
		Env:   &env,
	}
	resp, err := lockerClient.CreateSecret(&input)
	if err != nil {
		t.Fatalf("create secret broke, error: %v", err)
	}
	if resp.Key != CREATE_SEC_KEY {
		t.Fatalf("create secret decryption broke, expecting key \"%s\", getting %s instead", CREATE_SEC_KEY, resp.Key)
	}
	if resp.Value != CREATE_SEC_VAL {
		t.Fatalf("create secret decryption broke, expecting value \"%s\", getting %s instead", CREATE_SEC_VAL, resp.Value)
	}
	if *resp.EnvironmentName != INIT_ENV_NAME {
		t.Fatalf("create secret decryption broke, expecting env \"%s\", getting %s instead", INIT_ENV_NAME, *resp.EnvironmentName)
	}
}

// Current state
// Secret: (INIT_SEC_KEY, INIT_SEC_VAL, ALL), (CREATE_SEC_KEY, CREATE_SEC_VAL, ALL), (CREATE_SEC_KEY, CREATE_SEC_VAL, INIT_ENV_NAME)
// Env: (INIT_ENV_NAME, INIT_ENV_URL), (UPDATE_ENV_NAME, UPDATE_ENV_URL)
func TestCreateSecretWithEnvDup(t *testing.T) {
	// time.Sleep(3 * time.Second)
	key := CREATE_SEC_KEY
	value := CREATE_SEC_VAL
	env := INIT_ENV_NAME
	input := locker.InputSecData{
		Key:   &key,
		Value: &value,
		Env:   &env,
	}
	_, err := lockerClient.CreateSecret(&input)
	if err != nil {
		if !strings.Contains(err.Error(), "This secret hash already exists") {
			t.Fatalf("create secret broke, error: %v", err)
		}
	} else {
		t.Fatalf("expecting duplicated error")
	}
}

// Current state
// Secret: (INIT_SEC_KEY, INIT_SEC_VAL, ALL), (CREATE_SEC_KEY, CREATE_SEC_VAL, ALL), (CREATE_SEC_KEY, CREATE_SEC_VAL, INIT_ENV_NAME)
// Env: (INIT_ENV_NAME, INIT_ENV_URL), (UPDATE_ENV_NAME, UPDATE_ENV_URL)
func TestCreateSecretWithInvalidEnv(t *testing.T) {
	// time.Sleep(3 * time.Second)
	key := CREATE_SEC_KEY
	value := CREATE_SEC_VAL
	env := "not existed"
	input := locker.InputSecData{
		Key:   &key,
		Value: &value,
		Env:   &env,
	}
	_, err := lockerClient.CreateSecret(&input)
	if err != nil {
		if !strings.Contains(err.Error(), "no environment found with provided name") {
			t.Fatalf("create secret broke, error: %v", err)
		}
	} else {
		t.Fatalf("expecting env error")
	}
}

// Current state
// Secret: (INIT_SEC_KEY, INIT_SEC_VAL, ALL), (CREATE_SEC_KEY, CREATE_SEC_VAL, ALL), (CREATE_SEC_KEY, CREATE_SEC_VAL, INIT_ENV_NAME)
// Env: (INIT_ENV_NAME, INIT_ENV_URL), (UPDATE_ENV_NAME, UPDATE_ENV_URL)
func TestCreateSecretWithEnvDupKeyDiffEnv(t *testing.T) {
	// time.Sleep(3 * time.Second)
	key := CREATE_SEC_KEY
	value := CREATE_SEC_VAL
	env := UPDATE_ENV_NAME
	input := locker.InputSecData{
		Key:   &key,
		Value: &value,
		Env:   &env,
	}
	resp, err := lockerClient.CreateSecret(&input)
	if err != nil {
		t.Fatalf("create secret broke, error: %v", err)
	}
	if resp.Key != CREATE_SEC_KEY {
		t.Fatalf("create secret decryption broke, expecting key \"%s\", getting %s instead", CREATE_SEC_KEY, resp.Key)
	}
	if resp.Value != CREATE_SEC_VAL {
		t.Fatalf("create secret decryption broke, expecting value \"%s\", getting %s instead", CREATE_SEC_VAL, resp.Value)
	}
	if *resp.EnvironmentName != UPDATE_ENV_NAME {
		t.Fatalf("create secret decryption broke, expecting env \"%s\", getting %s instead", UPDATE_ENV_NAME, *resp.EnvironmentName)
	}
}

// Current state
// Secret: (INIT_SEC_KEY, INIT_SEC_VAL, ALL), (CREATE_SEC_KEY, CREATE_SEC_VAL, ALL), (CREATE_SEC_KEY, CREATE_SEC_VAL, INIT_ENV_NAME), (CREATE_SEC_KEY, CREATE_SEC_VAL, UPDATE_ENV_NAME)
// Env: (INIT_ENV_NAME, INIT_ENV_URL), (UPDATE_ENV_NAME, UPDATE_ENV_URL)
func TestUpdateSecret(t *testing.T) {
	// time.Sleep(3 * time.Second)
	key := UPDATE_SEC_KEY
	value := UPDATE_SEC_VAL
	targetEnv := UPDATE_ENV_NAME
	input := locker.InputSecData{
		Key:   &key,
		Value: &value,
	}
	resp, err := lockerClient.UpdateSecret(CREATE_SEC_KEY, &targetEnv, &input)
	if err != nil {
		t.Fatalf("update secret broke, error: %v", err)
	}
	if resp.Key != UPDATE_SEC_KEY {
		t.Fatalf("update secret decryption broke, expecting key \"%s\", getting %s instead", UPDATE_SEC_KEY, resp.Key)
	}
	if resp.Value != UPDATE_SEC_VAL {
		t.Fatalf("update secret decryption broke, expecting value \"%s\", getting %s instead", UPDATE_SEC_VAL, resp.Value)
	}
	if *resp.EnvironmentName != UPDATE_ENV_NAME {
		t.Fatalf("update secret decryption broke, expecting env \"%s\", getting %s instead", UPDATE_ENV_NAME, *resp.EnvironmentName)
	}
}

// Current state
// Secret: (INIT_SEC_KEY, INIT_SEC_VAL, ALL), (CREATE_SEC_KEY, CREATE_SEC_VAL, ALL), (CREATE_SEC_KEY, CREATE_SEC_VAL, INIT_ENV_NAME), (UPDATE_SEC_KEY, UPDATE_SEC_VAL, UPDATE_ENV_NAME)
// Env: (INIT_ENV_NAME, INIT_ENV_URL), (UPDATE_ENV_NAME, UPDATE_ENV_URL)
func TestUpdateSecretDup(t *testing.T) {
	// time.Sleep(3 * time.Second)
	key := UPDATE_SEC_KEY
	value := UPDATE_SEC_VAL
	env := UPDATE_ENV_NAME
	input := locker.InputSecData{
		Key:   &key,
		Value: &value,
		Env:   &env,
	}
	_, err := lockerClient.UpdateSecret(CREATE_SEC_KEY, nil, &input)
	if err != nil {
		if !strings.Contains(err.Error(), "This hash value already existed") {
			t.Fatalf("update secret broke, error: %v", err)
		}
	} else {
		t.Fatalf("expecting duplicated error")
	}
}

// Current state
// Secret: (INIT_SEC_KEY, INIT_SEC_VAL, ALL), (CREATE_SEC_KEY, CREATE_SEC_VAL, ALL), (CREATE_SEC_KEY, CREATE_SEC_VAL, INIT_ENV_NAME), (UPDATE_SEC_KEY, UPDATE_SEC_VAL, UPDATE_ENV_NAME)
// Env: (INIT_ENV_NAME, INIT_ENV_URL), (UPDATE_ENV_NAME, UPDATE_ENV_URL)
func TestUpdateSecretEnvToALL(t *testing.T) {
	// time.Sleep(3 * time.Second)
	key := UPDATE_SEC_KEY
	env := ""
	targetEnv := INIT_ENV_NAME
	input := locker.InputSecData{
		Key: &key,
		Env: &env,
	}
	resp, err := lockerClient.UpdateSecret(CREATE_SEC_KEY, &targetEnv, &input)
	if err != nil {
		t.Fatalf("update secret broke, error: %v", err)
	}
	if resp.Key != UPDATE_SEC_KEY {
		t.Fatalf("update secret decryption broke, expecting key \"%s\", getting %s instead", UPDATE_SEC_KEY, resp.Key)
	}
	if resp.Value != CREATE_SEC_VAL {
		t.Fatalf("update secret decryption broke, expecting value \"%s\", getting %s instead", CREATE_SEC_VAL, resp.Value)
	}
	if resp.EnvironmentName != nil {
		t.Fatalf("update secret decryption broke, expecting env nil, getting \"%s\" instead", *resp.EnvironmentName)
	}
}

// // Current state
// // Secret: (INIT_SEC_KEY, INIT_SEC_VAL, ALL), (CREATE_SEC_KEY, CREATE_SEC_VAL, ALL), (UPDATE_SEC_KEY, CREATE_SEC_VAL, ALL), (UPDATE_SEC_KEY, UPDATE_SEC_VAL, UPDATE_ENV_NAME)
// // Env: (INIT_ENV_NAME, INIT_ENV_URL), (UPDATE_ENV_NAME, UPDATE_ENV_URL)
