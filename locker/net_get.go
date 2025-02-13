package locker

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"sdk-test/types"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm/clause"
)

func (locker *Locker) httpActionIn(endpoint string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		types.CURRENT_ERR = types.ERR_HTTP
		return nil, fmt.Errorf("error creating new HTTP request: %w", err)
	}

	locker.setHeaders(req, false)

	var httpClient http.Client

	if strings.Contains(endpoint, "/v1/sync/revision_date") {
		httpClient.Timeout = 3 * time.Second
	} else {
		httpClient.Timeout = 10 * time.Second
	}

	if locker.Unsafe {
		customTransport := http.DefaultTransport.(*http.Transport).Clone()
		customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		httpClient.Transport = customTransport
	}

	res, err := httpClient.Do(req)
	statusCode := -1
	if res != nil {
		statusCode = res.StatusCode
	}
	if err != nil {
		types.CURRENT_ERR = types.ERR_HTTP
		return nil, fmt.Errorf("error executing HTTP request: %w", err)
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		types.CURRENT_ERR = types.ERR_FUNC
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	if statusCode != 200 {
		srvMsg, err := unmarshalAny[types.ServerErrorMsg](resBody)
		if err != nil {
			return nil, err
		}

		if statusCode < 600 && statusCode >= 500 {
			types.CURRENT_ERR = types.ERR_SERVER
		} else {
			types.CURRENT_ERR = types.ERR_HTTP
		}

		return nil, fmt.Errorf("HTTP failure: %d, %s", statusCode, srvMsg.Message)
	}

	return resBody, nil
}

func (locker *Locker) bulkUpdate(kind string, resBody []byte) (string, error) {
	var next string
	switch kind {
	case types.FETCH_KIND_SEC, types.FETCH_KIND_RUN:
		fetchedSec, err := unmarshalAny[types.SecretResponse](resBody)
		if err != nil {
			return "", err
		}

		if fetchedSec.Count == 0 {
			return fetchedSec.Next, nil
		}

		// if table empty, batch insert
		result := locker.dBConn.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).CreateInBatches(&fetchedSec.Results, 2000)
		if result.Error != nil {
			return "", result.Error
		}

		next = fetchedSec.Next

		// upsert revision date
		if next == "" {
			revDateEntry := types.RevisionDate{
				ID:           0,
				RevisionDate: fetchedSec.RevisionDate,
				LastCallSec:  float64(time.Now().Unix()),
			}

			result := locker.dBConn.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "ID"}},
				DoUpdates: clause.AssignmentColumns([]string{"revision_date", "last_call_sec"}),
			}).Create(&revDateEntry)
			if result.Error != nil {
				return "", result.Error
			}
		}

	case types.FETCH_KIND_ENV:
		fetchedEnv, err := unmarshalAny[types.EnvironmentResponse](resBody)
		if err != nil {
			return "", err
		}

		if fetchedEnv.Count == 0 {
			return fetchedEnv.Next, nil
		}

		result := locker.dBConn.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "hash"}},
			UpdateAll: true,
		}).CreateInBatches(&fetchedEnv.Results, 2000)
		if result.Error != nil {
			return "", result.Error
		}

		next = fetchedEnv.Next

		// upsert revision date
		if next == "" {
			revDateEntry := types.RevisionDate{
				ID:           0,
				RevisionDate: fetchedEnv.RevisionDate,
				LastCallEnv:  float64(time.Now().Unix()),
			}

			result := locker.dBConn.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "ID"}},
				DoUpdates: clause.AssignmentColumns([]string{"revision_date", "last_call_env"}),
			}).Create(&revDateEntry)
			if result.Error != nil {
				return "", result.Error
			}
		}

	case types.FETCH_KIND_PROFILE:
		// insert/update to profile table
		fetchedProfile, err := unmarshalAny[types.ProfileResponse](resBody)
		if err != nil {
			return "", err
		}

		profile := formatProfile(*fetchedProfile)

		result := locker.dBConn.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&profile)
		if result.Error != nil {
			return "", result.Error
		}
	}

	return next, nil
}

func (locker *Locker) fetchDataFromServer(primaryFilter string, revisionDate types.RevisionDate, kind string) error {
	var dataEndpoint string
	var fallbackEndpoint string
	page := 1
	revDate := revisionDate.RevisionDate
	switch kind {
	case types.FETCH_KIND_SEC:
		if revisionDate.LastCallSec == 0 {
			revDate = 0
		}
		dataEndpoint = fmt.Sprintf("%s/v1/%s?count_secrets=1&page=%d&paging=1&revision_date=%f&size=2000&hash=%s", locker.APIBase, kind, page, revDate, primaryFilter)
		fallbackEndpoint = fmt.Sprintf("%s/v1/%s?count_secrets=1&page=%d&paging=1&revision_date=0&size=2000&hash=%s", locker.APIBase, kind, page, primaryFilter)

	case types.FETCH_KIND_ENV:
		if revisionDate.LastCallEnv == 0 {
			revDate = 0
		}
		dataEndpoint = fmt.Sprintf("%s/v1/%s?count_environment=1&page=%d&paging=1&revision_date=%f&size=2000&hash=%s", locker.APIBase, kind, page, revDate, primaryFilter)
		fallbackEndpoint = fmt.Sprintf("%s/v1/%s?count_environment=1&page=%d&paging=1&revision_date=0&size=2000&hash=%s", locker.APIBase, kind, page, primaryFilter)

	case types.FETCH_KIND_PROFILE:
		dataEndpoint = fmt.Sprintf("%s/v1/%s", locker.APIBase, kind)

	case types.FETCH_KIND_RUN:
		if revisionDate.LastCallSec == 0 {
			revDate = 0
		}
		dataEndpoint = fmt.Sprintf("%s/v1/%s?count_secrets=1&page=%d&paging=1&revision_date=%f&size=2000&environment_id=%s", locker.APIBase, "secrets", page, revDate, primaryFilter)
		fallbackEndpoint = fmt.Sprintf("%s/v1/%s?count_secrets=1&page=%d&paging=1&revision_date=0&size=2000&environment_id=%s", locker.APIBase, "secrets", page, primaryFilter)
	}

	resBody, err := locker.httpActionIn(dataEndpoint)
	if err != nil {
		return err
	}

	genericData, err := unmarshalAny[types.GenericList](resBody)
	if err != nil {
		return err
	}

	if genericData.Count == 0 && kind != types.FETCH_KIND_PROFILE {
		// last ditch effort, fetch again with revDate = 0
		resBody, err = locker.httpActionIn(fallbackEndpoint)
		if err != nil {
			return err
		}
		genericData, err = unmarshalAny[types.GenericList](resBody)
		if err != nil {
			return err
		}

		if genericData.Count == 0 {
			locker.emptyFetch = true
		}
	}

	// insert if not exist, else update
	next, err := locker.bulkUpdate(kind, resBody)
	if err != nil {
		return err
	}

	// if there's still data left
	for next != "" {
		// fetch again
		dataEndpoint = fmt.Sprintf("%s%s", locker.APIBase, next)
		resBody, err := locker.httpActionIn(dataEndpoint)
		if err != nil {
			return err
		}

		next, err = locker.bulkUpdate(kind, resBody)
		if err != nil {
			return err
		}

	}

	return nil
}

func (locker *Locker) fetchRevisionDate(kind string) (types.RevisionDate, bool, error) {
	var lastCall float64
	var localRevDate types.RevisionDate
	var err error

	// query for last call time
	localRevDate, err = locker.queryRevisionDate()
	if err != nil {
		return localRevDate, false, err
	}

	switch kind {
	case types.FETCH_KIND_SEC:
		lastCall = localRevDate.LastCallSec
	case types.FETCH_KIND_ENV:
		lastCall = localRevDate.LastCallEnv
	}

	// if the current time is less than 2 minutes since last call, default to local data
	if float64(time.Now().Unix()) < lastCall+float64(locker.Cooldown) {
		return localRevDate, true, nil
	}

	// otherwise, call revision date api
	dataEndpoint := locker.APIBase + "/v1/sync/revision_date"
	resBody, err := locker.httpActionIn(dataEndpoint)
	if err != nil {
		// encountering problem fetching date from server, falling back to local rev date
		return localRevDate, true, nil
	}

	fetchedRevDate, err := strconv.ParseFloat(string(resBody), 64)
	if err != nil {
		types.CURRENT_ERR = types.ERR_FUNC
		return localRevDate, false, fmt.Errorf("error parsing revision date: %w", err)
	}

	localRevDate.RevisionDate = fetchedRevDate
	switch kind {
	case types.FETCH_KIND_SEC:
		localRevDate.LastCallSec = float64(time.Now().Unix())
	case types.FETCH_KIND_ENV:
		localRevDate.LastCallEnv = float64(time.Now().Unix())
	}

	return localRevDate, false, nil
}

func (locker *Locker) fetchCount(kind string) (int64, error) {
	var dataEndpoint string
	switch kind {
	case types.FETCH_KIND_SEC, types.FETCH_KIND_RUN:
		dataEndpoint = locker.APIBase + "/v1/sync/secrets/count"

	case types.FETCH_KIND_ENV:
		dataEndpoint = locker.APIBase + "/v1/sync/environments/count"
	}

	resBody, err := locker.httpActionIn(dataEndpoint)
	if err != nil {
		return -1, err
	}

	fetchedCount, err := strconv.ParseInt(string(resBody), 10, 64)
	if err != nil {
		return -1, fmt.Errorf("error parsing data count: %w", err)
	}

	return fetchedCount, nil
}
