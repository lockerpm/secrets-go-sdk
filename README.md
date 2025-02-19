# Locker Secret NodeJS SDK

<p align="center">
  <img src="https://cystack.net/images/logo-black.svg" alt="CyStack" width="50%"/>
</p>


---

The Locker Secret Go SDK provides convenient access to the Locker Secret API from applications written in the 
Go language. It includes a pre-defined set of classes for API resources that initialize themselves dynamically 
from API responses which makes it compatible with a wide range of versions of the Locker Secret API.


## The Developer - CyStack

The Locker Secret Go SDK is developed by CyStack, one of the leading cybersecurity companies in Vietnam. 
CyStack is a member of Vietnam Information Security Association (VNISA) and Vietnam Association of CyberSecurity 
Product Development. CyStack is a partner providing security solutions and services for many large domestic and 
international enterprises.

CyStack’s research has been featured at the world’s top security events such as BlackHat USA (USA), 
BlackHat Asia (Singapore), T2Fi (Finland), XCon - XFocus (China)... CyStack experts have been honored by global 
corporations such as Microsoft, Dell, Deloitte, D-link...


## Documentation

The documentation will be updated later.

## Requirements

- Go 1.12+

## Installation

Install with go get:

```bash
go get https://github.com/lockerpm/secrets-sdk-go
```

## Usages

### Set up access key

The SDK needs to be configured with your access key which is available in your Locker Secret Dashboard. 
Either setup a .env file or use the `SetAccessKeyID` and `SetSecretAccessKey` setters directly. 
You may also need to set `APIBase` value (default value is `https://api.locker.io/locker_secrets`).

```go
import (
	"sdk-test/locker"
)

func main() {
	var lockerClient locker.Locker
	lockerClient.NewLockerClient()
	lockerClient.SetAccessKeyID("YOUR_ACCESS_KEY_ID")
	lockerClient.SetSecretAccessKey("YOUR_SECRET_ACCESS_KEY")
}
```

All initialization options are listed below:

| Key                   | Description                                                                                  | Type                   | Required |
| --------------------- | ---------------------------------------------------------------------------------------------| -----------------------| :--:     |
| SetAccessKeyID        | Your access key id                                                                           | `string`               | ✅       | 
| SetSecretAccessKey    | Your access key secret                                                                       | `string`               | ✅       | 
| SetAPIBase            | Your server base API URL, default value is `https://api.locker.io/locker_secrets`            | `string`               | ❌       | 
| SetHeaders            | Custom headers for API calls                                                                 | `map[string]string`    | ❌       | 
| SetCooldown           | Timeout between API calls, default is `120` (seconds)                                        | `int`                  | ❌       | 
| SetFetch              | Force fetching data from the server, can override SetCooldown, default is `true`             | `boolean`              | ❌       |
| SetUnsafe             | Set TLS to unsafe if you use a server with self-signed certificate, default value is `false` | `boolean`              | ❌       |
| SetWorkingDir         | Secret's working directory, containing sqlite database, default is `$home/.locker`           | `string`               | ❌       |

Now, you can use SDK to get or set values:

```go
// Get list secrets quickly
secrets, err := lockerClient.ListSecret(nil)

// List secrets by environment
targetEnv := "production"
secret, err := lockerClient.ListSecret(*targetedEnv)

// Get a secret value by secret key
// Replace 'ENVIRONMENT' with nil to get secret from the environment ALL
secretValue1, err := lockerClient.GetSecret("SECRET_NAME_1", nil)
secretValue2, err := lockerClient.GetSecret("SECRET_NAME_2", "ENVIRONMENT")

// Create new secret
key := "key"
value := "value"
desc := "description"
env := "environment"
input := locker.InputSecData{
    Key:   &key,
    Value: &value,
    Desc:  &desc,
    Env:   &env,
}
result, err := lockerClient.CreateSecret(&input)

// Update secret
newKey := "new key"
newValue := "new value"
newEnv := "new env" 
targetKey := "target key"
targetEnv := "target env"
input := locker.InputSecData{
    Key:   &newKey,
    Value: &newValue,
    Env:   &newEnv, // use nil to set environment to ALL
}
resp, err := lockerClient.UpdateSecret(targetKey, &targetEnv, &input)

// List environments
envs, err := lockerClient.ListEnvironment()

// Get an environment object by name
env, err := lockerClient.ListSecret("production")

// Create new environment
name := "name"
url := "externalUrl"
desc := "description"
input := locker.InputEnvData{
    Name: &name,
    Url:  &url,
    Desc: &desc,
}
resp, err := lockerClient.CreateEnvironment(&input)

// Update an environment by name
name := "new name"
url := "new externalUrl"
desc := "new description"
input := locker.InputEnvData{
    Name: &name,
    Url:  &url,
    Desc: &desc
}
resp, err := lockerClient.UpdateEnvironment("target env", &input)
```

### Caching

By default, Locker fetches data from the cloud server once and stores it in local storage. It only checks for updates every 120 seconds to prevent unnecessary API calls. You can change this behavior by using `SetFetch` and `SetCooldown`

```go
// Object level, this config will apply to all methods
var lockerClient locker.Locker
// ...
lockerClient.SetFetch(true)   // setting it to true will force Locker to fetch from the cloud server instead of local storage
lockerClient.SetCooldown(5)   // seconds, only accept integer value

## Development

Install required packages.
```bash
go get https://github.com/lockerpm/secrets-sdk-go
```

### Run tests

Create a .env file with required access keys (refer to `.env.example`)

To run all tests, use:
```bash
go test
```

## Reporting security issues

We take the security and our users' trust very seriously. If you found a security issue in Locker SDK Python, please 
report the issue by contacting us at <contact@locker.io>. Do not file an issue on the tracker. 


## Contributing

Please check [CONTRIBUTING](CONTRIBUTING.md) before making a contribution.


## Help and media

- FAQ: https://support.locker.io

- Community Q&A: https://forum.locker.io

- News: https://locker.io/blog


## License