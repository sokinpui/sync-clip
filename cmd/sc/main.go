package main

import (
	"os"
	syncclip "sync-clip"
)

func main() {
	if err := syncclip.Execute(); err != nil {
		os.Exit(1)
	}
}
