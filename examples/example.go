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
