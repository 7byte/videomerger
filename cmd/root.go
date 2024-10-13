package cmd

import (
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	loglevel string
)

var rootCmd = &cobra.Command{
	Use:   "merge_xiaomi_monitor_video",
	Short: "Merge xiaomi monitor video files",
	Long:  "Merge xiaomi monitor video files",
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&loglevel, "log.level", "l", "info", "log level (debug, info, warn, error)")
	cobra.OnInitialize(initLog)
}

func initLog() {
	var l slog.Level
	switch strings.ToUpper(loglevel) {
	case "DEBUG":
		l = slog.LevelDebug
	case "INFO":
		l = slog.LevelInfo
	case "WARN":
		l = slog.LevelWarn
	case "ERROR":
		l = slog.LevelError
	default:
		l = slog.LevelInfo
	}
	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: l})
	slog.SetDefault(slog.New(h))
}
