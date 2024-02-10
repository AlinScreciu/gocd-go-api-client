
# GoCD Golang API Client
[![Go Reference](https://pkg.go.dev/badge/github.com/AlinScreciu/gocd-go-api-client.svg)](https://pkg.go.dev/github.com/AlinScreciu/gocd-go-api-client)
[![codecov](https://codecov.io/github/AlinScreciu/gocd-go-api-client/graph/badge.svg?token=JDFXFK55XV)](https://codecov.io/github/AlinScreciu/gocd-go-api-client)

A golang client for the GoCD API.


## Installation

```bash
go get github.com/AlinScreciu/gocd-go-api-client
```
    

    
## Usage/Examples

```golang
package main

import (
	"fmt"
	"os"

	"github.com/AlinScreciu/gocd-go-api-client/pkg/client"
)

func main() {

	client, err := client.NewClient("https://gocd.8x8.com/go")
	if err != nil {
		fmt.Printf("failed to create client: '%s'\n", err.Error())
		os.Exit(1)
	}

	client.SetAccessToken("<YOUR_ACCESS_TOKEN>")
	// client.SetBasicAuth("<USERNAME>", "<PASSWORD>")

	version, err := client.GetVersion()
	if err != nil {
		os.Exit(1)
	}

	currentUser, err := client.GetCurrentUser()
	if err != nil {
		os.Exit(1)
	}

	fmt.Printf("I am %s, GoCD version: %s\n", currentUser.DisplayName, version.Version)
}

```




## License

[MIT](https://choosealicense.com/licenses/mit/)


## Contributing

Contributions are always welcome!

See `CONTRIBUTING.md` for ways to get started.


