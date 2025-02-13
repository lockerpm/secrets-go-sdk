package locker

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"sdk-test/types"
	"time"
)

func (locker *Locker) httpActionOut(method, endpoint string, body []byte) ([]byte, int, error) {
	var req *http.Request
	var err error
	switch method {
	case "POST":
		req, err = http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(body))
	case "PUT":
		req, err = http.NewRequest(http.MethodPut, endpoint, bytes.NewBuffer(body))
	default:
		req, err = http.NewRequest(http.MethodGet, endpoint, bytes.NewBuffer(body))
	}

	if err != nil {
		types.CURRENT_ERR = types.ERR_HTTP
		return nil, -1, fmt.Errorf("error creating new HTTP request: %w", err)
	}

	locker.setHeaders(req, true)

	client := http.Client{
		Timeout: 30 * time.Second,
	}

	res, err := client.Do(req)
	statusCode := -1
	if res != nil {
		statusCode = res.StatusCode
	}

	if err != nil {
		types.CURRENT_ERR = types.ERR_HTTP
		return nil, statusCode, fmt.Errorf("error executing HTTP request: %w", err)
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		types.CURRENT_ERR = types.ERR_FUNC
		return nil, statusCode, fmt.Errorf("error reading response: %w", err)
	}

	if statusCode != 201 && statusCode != 200 {
		srvMsg, err := unmarshalAny[types.ServerErrorMsg](resBody)
		if err != nil {
			return nil, statusCode, err
		}

		if statusCode < 600 && statusCode >= 500 {
			types.CURRENT_ERR = types.ERR_SERVER
		} else {
			types.CURRENT_ERR = types.ERR_HTTP
		}

		return nil, statusCode, fmt.Errorf("HTTP failure: %d, %s", statusCode, srvMsg.Message)
	}

	return resBody, statusCode, nil
}

func editItem[Struct any](locker *Locker, kind, ID string, body []byte) (*Struct, error) {
	var dataEndpoint string
	switch kind {
	case types.FETCH_KIND_SEC:
		dataEndpoint = fmt.Sprintf("%s/v1/secrets/%s", locker.APIBase, ID)
	case types.FETCH_KIND_ENV:
		dataEndpoint = fmt.Sprintf("%s/v1/environments/%s", locker.APIBase, ID)
	}

	resBody, code, err := locker.httpActionOut("PUT", dataEndpoint, body)
	if err != nil {
		return nil, err
	}

	if code == 404 {
		switch kind {
		case types.FETCH_KIND_SEC:
			result := locker.dBConn.Delete(&types.Secret{}, ID)
			if result.Error != nil {
				types.CURRENT_ERR = types.ERR_DB
				return nil, fmt.Errorf("error deleting secret: %w", result.Error)
			}
		case types.FETCH_KIND_ENV:
			result := locker.dBConn.Delete(&types.Environment{}, ID)
			if result.Error != nil {
				types.CURRENT_ERR = types.ERR_DB
				return nil, fmt.Errorf("error deleting environment: %w", result.Error)
			}
		}
	}

	encRes, err := unmarshalAny[Struct](resBody)
	if err != nil {
		return nil, err
	}

	return encRes, nil
}

func createItem[Struct any](locker *Locker, kind string, body []byte) (*Struct, error) {
	var dataEndpoint string
	switch kind {
	case types.FETCH_KIND_SEC:
		dataEndpoint = fmt.Sprintf("%s/v1/secrets", locker.APIBase)
	case types.FETCH_KIND_ENV:
		dataEndpoint = fmt.Sprintf("%s/v1/environments", locker.APIBase)
	}

	resBody, _, err := locker.httpActionOut("POST", dataEndpoint, body)
	if err != nil {
		return nil, err
	}

	encRes, err := unmarshalAny[Struct](resBody)
	if err != nil {
		return nil, err
	}

	return encRes, nil
}
