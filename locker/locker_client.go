package locker

import (
	"log"
	"os"
	"path/filepath"

	"github.com/lockerpm/secrets-sdk-go/types"

	"gorm.io/gorm"
)

type Locker struct {
	// secret
	Key         string
	Environment string

	// environment
	Name string

	// input
	NewKey                 string
	NewValue               string
	NewDescription         string
	NewEnvironment         string
	IdentifyingEnvironment string

	// common
	AccessKeyID      string
	SecretAccessKey  string
	APIBase          string
	APIVersion       string
	Output           string
	WorkingDir       string
	OutputPath       string
	DBPath           string
	Headers          map[string]string
	LogLevel         int
	MaxRetry         int
	Cooldown         int
	Fetch            bool
	Export           bool
	Unsafe           bool
	GettingFromLocal bool

	dBConn           *gorm.DB
	hash             string
	emptyFetch       bool
	symKey           []byte
	macKey           []byte
	currentOperation string
}

func (locker *Locker) NewLockerClient() {
	locker.APIBase = types.DEFAULT_API_BASE
	locker.APIVersion = "v1"
	locker.LogLevel = 1
	locker.Headers = make(map[string]string)
	locker.Headers["Cache-Control"] = "no-cache"
	locker.Headers["Content-Type"] = "application/json"
	locker.Fetch = true
	locker.Cooldown = 120

	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = os.TempDir()
	}

	workingDir := filepath.Join(homeDir, ".locker")

	if _, err := os.Stat(workingDir); os.IsNotExist(err) {
		err := os.MkdirAll(workingDir, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}

	locker.WorkingDir = workingDir
	locker.OutputPath = filepath.Join(locker.WorkingDir, "output.txt")
}
