package syncclip

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"os"
)

var serverAddr string
var configPath string

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sc",
		Short: "Sync-clip CLI client",
		Long:  "A command-line tool to synchronize clipboard contents across devices.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := LoadConfig(configPath, "sc.conf")
			if err != nil {
				return err
			}
			serverAddr = cfg.URL
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if isInputPiped() {
				return pushToRemote()
			}
			return pullFromRemote()
		},
	}

	cmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "Config file path")

	return cmd
}

func isInputPiped() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) == 0
}

func pushToRemote() error {
	reader := bufio.NewReader(os.Stdin)
	peek, _ := reader.Peek(1024)
	contentType := http.DetectContentType(peek)

	resp, err := http.Post(serverAddr, contentType, reader)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned error: %s", resp.Status)
	}

	return nil
}

func pullFromRemote() error {
	resp, err := http.Get(serverAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned error: %s", resp.Status)
	}

	_, err = io.Copy(os.Stdout, resp.Body)
	return err
}

func Execute() error {
	rootCmd := NewRootCommand()
	return rootCmd.Execute()
}
