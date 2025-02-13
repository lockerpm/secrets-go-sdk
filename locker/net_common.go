package locker

import (
	"net/http"
	"sdk-test/types"
)

func (locker *Locker) setHeaders(req *http.Request, post bool) {
	req.Header.Set("Authorization", "Bearer "+locker.AccessKeyID)
	req.Header.Set("User-Agent", "Locker Secret Go SDK - version "+types.VERSION)

	if post {
		req.Header.Set("Content-Type", "application/json")
	}

	if locker.Headers == nil {
		return
	}

	for header, value := range locker.Headers {
		req.Header.Set(header, value)
	}
}
