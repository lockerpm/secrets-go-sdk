package locker

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lockerpm/secrets-sdk-go/types"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

func (locker *Locker) initDB() error {
	locker.DBPath = filepath.Join(locker.WorkingDir, fmt.Sprintf("%s-data.db", locker.AccessKeyID))
	if !locker.checkIfDBExist() {
		_, err := os.Create(locker.DBPath)
		if err != nil {
			types.CURRENT_ERR = types.ERR_PATH
			return fmt.Errorf("error creating DB file: %w", err)
		}
	}

	err := locker.connectToDB()
	if err != nil {
		return err
	}

	err = locker.dBConn.AutoMigrate(&types.DBVersion{})
	if err != nil {
		types.CURRENT_ERR = types.ERR_DB
		return fmt.Errorf("error migrating db version: %w", err)
	}
	var CurrentDbVer types.DBVersion
	result := locker.dBConn.First(&CurrentDbVer)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		types.CURRENT_ERR = types.ERR_DB
		return fmt.Errorf("error querying db version: %w", result.Error)
	}

	// remove legacy tables
	locker.dBConn.Migrator().DropTable("secret")
	locker.dBConn.Migrator().DropTable("environment")
	locker.dBConn.Migrator().DropTable("profile")
	locker.dBConn.Migrator().DropTable("date")

	if types.DB_REVISION_NUMBER > CurrentDbVer.DbRevisionNumber {
		locker.dBConn.Migrator().DropTable("secrets")
		locker.dBConn.Migrator().DropTable("environments")
		locker.dBConn.Migrator().DropTable("profiles")
		locker.dBConn.Migrator().DropTable("revision_dates")
		locker.dBConn.Migrator().DropTable("deletion_dates")

		CurrentDbVer.DbRevisionNumber = types.DB_REVISION_NUMBER
		result := locker.dBConn.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&CurrentDbVer)
		if result.Error != nil {
			types.CURRENT_ERR = types.ERR_DB
			return fmt.Errorf("error querying db version: %w", result.Error)
		}

		err = locker.dBConn.AutoMigrate(&types.Profile{}, &types.Secret{}, &types.Environment{}, &types.DeletionDate{}, &types.RevisionDate{})
		if err != nil {
			types.CURRENT_ERR = types.ERR_DB
			return fmt.Errorf("error migrating typesbase: %w", err)
		}

	}

	return nil
}

func (locker *Locker) connectToDB() error {
	var err error
	if locker.dBConn == nil {
		locker.dBConn, err = gorm.Open(sqlite.Open(locker.DBPath), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		locker.dBConn.Statement.RaiseErrorOnNotFound = true
		if err != nil {
			types.CURRENT_ERR = types.ERR_PATH
			return fmt.Errorf("typesbase file at %s not available: %w", locker.DBPath, err)
		}
	}

	return nil
}

func (locker *Locker) checkIfDBExist() bool {
	if _, err := os.Stat(locker.DBPath); os.IsNotExist(err) {
		return false
	}
	return true
}

func (locker *Locker) queryRevisionDate() (types.RevisionDate, error) {
	var localRevDate types.RevisionDate
	result := locker.dBConn.First(&localRevDate)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			localRevDate.ID = 0
			localRevDate.LastCallEnv = 0
			localRevDate.LastCallSec = 0
			localRevDate.RevisionDate = 0
			return localRevDate, nil
		}
		types.CURRENT_ERR = types.ERR_DB
		return localRevDate, fmt.Errorf("error querying revision date: %w", result.Error)
	}
	return localRevDate, nil
}

func (locker *Locker) queryDeletionDate() (types.DeletionDate, error) {
	var localDelDate types.DeletionDate
	result := locker.dBConn.First(&localDelDate)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		types.CURRENT_ERR = types.ERR_DB
		return localDelDate, fmt.Errorf("error querying deletion date: %w", result.Error)
	} else if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		localDelDate.ID = 0
		localDelDate.DeletionDate = 0
	}

	return localDelDate, nil
}

func (locker *Locker) upsertRevisionDate(revDate types.RevisionDate) error {
	revDate.ID = 0
	result := locker.dBConn.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&revDate)

	if result.Error != nil {
		types.CURRENT_ERR = types.ERR_DB
		return fmt.Errorf("error upserting revision date: %w", result.Error)
	}

	return nil
}

func formatProfile(input types.ProfileResponse) types.Profile {
	return types.Profile{
		ID:             input.Profile.ID,
		ClientID:       input.Profile.ClientID,
		Key:            input.Profile.Key,
		Activated:      input.Profile.Activated,
		Editable:       input.Profile.Editable,
		CreationDate:   input.Profile.CreationDate,
		RevisionDate:   input.Profile.RevisionDate,
		ExpirationDate: input.Profile.ExpirationDate,
		ProjectID:      input.Profile.ProjectID,
		ProjectsStr:    strings.Join(input.Profile.Projects, ","),
		RestrictIPStr:  strings.Join(input.Profile.RestrictIP, ","),
	}
}

func (locker *Locker) beforeDelete(tx *gorm.DB, env *types.Environment) (err error) {
	var secretCascasde []types.Secret
	result := locker.dBConn.Where("environment_id = ?", env.ID).Find(&secretCascasde)
	if result.Error != nil && result.RowsAffected != 0 {
		return err
	}

	result = locker.dBConn.Delete(&secretCascasde)
	if result.Error != nil && result.RowsAffected != 0 {
		return err
	}

	return nil
}
