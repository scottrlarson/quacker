// clean.go

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func cleanInteractive() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Do you want to clean all sites? (yes/no)")
	sitesInput, _ := reader.ReadString('\n')
	sitesInput = strings.TrimSpace(strings.ToLower(sitesInput))
	if sitesInput == "yes" {
		cleanSites()
	}

	fmt.Println("Do you want to clean configuration settings? (yes/no)")
	configInput, _ := reader.ReadString('\n')
	configInput = strings.TrimSpace(strings.ToLower(configInput))
	if configInput == "yes" {
		cleanConfig()
	}

	fmt.Println("Do you want to clean all subscribers? (yes/no)")
	subscribersInput, _ := reader.ReadString('\n')
	subscribersInput = strings.TrimSpace(strings.ToLower(subscribersInput))
	if subscribersInput == "yes" {
		cleanSubscribers()
	}

	fmt.Println("Cleaning process completed.")
}

func cleanSites() {
	fmt.Println("Cleaning all sites...")
	rdb.FlushDB(ctx)
	fmt.Println("All sites cleaned.")
}

func cleanConfig() {
	fmt.Println("Cleaning configuration settings...")
	rdb.Del(ctx, "config")
	fmt.Println("Configuration settings cleaned.")
}

func cleanSubscribers() {
	fmt.Println("Cleaning all subscribers...")
	keys, _ := rdb.Keys(ctx, "subs:*").Result()
	for _, key := range keys {
		rdb.Del(ctx, key)
	}
	fmt.Println("All subscribers cleaned.")
}
