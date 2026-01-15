package main

import (
	"os"
	syncclip "github.com/sokinpui/sync-clip"
)

func main() {
	if err := syncclip.Execute(); err != nil {
		os.Exit(1)
	}
}
