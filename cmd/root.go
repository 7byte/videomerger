package cmd

import (
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	loglevel   string
	inputPath  string
	outputPath string
)

var rootCmd = &cobra.Command{
	Use:   "videomerger",
	Short: "Merge monitor video files",
	Long:  "Merge monitor video files",
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&loglevel, "log_level", "l", "info", "日志等级（debug, info, warn, error）")
	rootCmd.PersistentFlags().StringVarP(&inputPath, "input_path", "i", "", "待处理视频文件的目录")
	rootCmd.PersistentFlags().StringVarP(&outputPath, "output_path", "o", ".", "处理后视频的输出路径，默认为当前路径")
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
