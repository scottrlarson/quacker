package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
)

var (
	ctx = context.Background()
	rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	currentVersion = "development" // Default if not set during build
)

func ensureSetup() {
	// Check if the configuration exists in Redis
	_, err := rdb.Get(ctx, "config").Result()
	if err != nil {
		fmt.Println("Please configure Quacker first using ./quacker --setup")
		os.Exit(1)
	}
}

func incrementVersion() string {
	parts := strings.Split(currentVersion, ".")
	if len(parts) != 3 {
		return currentVersion // Keep the current version if the format is invalid
	}

	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return currentVersion
	}

	patch++
	newVersion := fmt.Sprintf("%s.%s.%d", parts[0], parts[1], patch)
	currentVersion = newVersion
	return newVersion
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--setup":
			setup()
		case "--run":
			runServer()
		case "--job":
			job()
		case "--allow":
			if len(os.Args) < 3 {
				fmt.Println("Usage: quacker --allow <GitHubUsername>")
				os.Exit(1)
			}
			allowGitHubUser(os.Args[2])
		case "--remove":
			if len(os.Args) < 3 {
				fmt.Println("Usage: quacker --remove <GitHubUsername>")
				os.Exit(1)
			}
			removeGitHubUser(os.Args[2])
		case "--version":
			fmt.Printf("Quacker version: %s\n", currentVersion)
		case "--clean":
			cleanInteractive()
		default:
			fmt.Println("Invalid argument")
		}
	} else {
		runServer()
	}
}
