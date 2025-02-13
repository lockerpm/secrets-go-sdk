package locker

import (
	"fmt"
	"os"
	"sdk-test/types"
)

func jsonLog(errMessage string) {
	fmt.Printf("{\n  \"object\": \"error\",\n  \"error\": \"%s\",\n  \"message\": \"%s\"\n}\n", types.CURRENT_ERR, errMessage)
	os.Exit(1)
}

func jsonLogSucess(message string) {
	fmt.Printf("{\n  \"object\": \"log\",\n  \"message\": \"%s\"\n}\n", message)
}
