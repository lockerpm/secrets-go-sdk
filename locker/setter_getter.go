package locker

import "log"

func (locker *Locker) GetAccessKeyID() string {
	return locker.AccessKeyID
}

func (locker *Locker) SetAccessKeyID(accessKeyID string) {
	locker.AccessKeyID = accessKeyID
	err := locker.initDB()
	if err != nil {
		log.Fatal(err)
	}

	err = locker.prepareProfile()
	if err != nil {
		log.Fatal(err)
	}
}

func (locker *Locker) GetSecretAccessKey() string {
	return locker.SecretAccessKey
}

func (locker *Locker) SetSecretAccessKey(secretAccessKey string) {
	locker.SecretAccessKey = secretAccessKey
}

func (locker *Locker) GetAPIBase() string {
	return locker.APIBase
}

func (locker *Locker) SetAPIBase(apiBase string) {
	locker.APIBase = apiBase
}

func (locker *Locker) GetAPIVersion() string {
	return locker.APIVersion
}

func (locker *Locker) SetAPIVersion(apiVersion string) {
	locker.APIVersion = apiVersion
}

func (locker *Locker) GetOutput() string {
	return locker.Output
}

func (locker *Locker) SetOutput(output string) {
	locker.Output = output
}

func (locker *Locker) GetHeaders() map[string]string {
	return locker.Headers
}

func (locker *Locker) SetHeaders(headers map[string]string) {
	locker.Headers = headers
}

func (locker *Locker) GetLogLevel() int {
	return locker.LogLevel
}

func (locker *Locker) SetLogLevel(logLevel int) {
	locker.LogLevel = logLevel
}

func (locker *Locker) GetMaxRetry() int {
	return locker.MaxRetry
}

func (locker *Locker) SetMaxRetry(maxRetry int) {
	locker.MaxRetry = maxRetry
}

func (locker *Locker) GetCooldown() int {
	return locker.Cooldown
}

func (locker *Locker) SetCooldown(cooldown int) {
	locker.Cooldown = cooldown
}

func (locker *Locker) GetFetch() bool {
	return locker.Fetch
}

func (locker *Locker) SetFetch(fetch bool) {
	locker.Fetch = fetch
}

func (locker *Locker) GetUnsafe() bool {
	return locker.Unsafe
}

func (locker *Locker) SetUnsafe(unsafe bool) {
	locker.Unsafe = unsafe
}

func (locker *Locker) GetWorkingDir() string {
	return locker.WorkingDir
}

func (locker *Locker) SetWorkingDir(workingDir string) {
	locker.WorkingDir = workingDir
}

func (locker *Locker) GetGettingFromLocal() bool {
	return locker.GettingFromLocal
}

func (locker *Locker) SetGettingFromLocal(gettingFromLocal bool) {
	locker.GettingFromLocal = gettingFromLocal
}
