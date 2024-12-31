package cmd

import (
	"log/slog"
	"time"

	"github.com/7byte/videomerger/internal"
	"github.com/spf13/cobra"
	"gocv.io/x/gocv"
)

type detectFlags struct {
	threshold   float64
	minDuration float64
	maxGap      float64
	sampleRate  float64
}

var defaultDetectFlags detectFlags

var detectCmd = &cobra.Command{
	Use:   "detect",
	Short: "Detect face in video",
	Long:  "Detect face in video",
	Run: func(cmd *cobra.Command, args []string) {
		detect()

		// 检查内存泄漏
		// fmt.Printf("final MatProfile count: %v\n", gocv.MatProfile.Count())
		// var b bytes.Buffer
		// gocv.MatProfile.WriteTo(&b, 1)
		// fmt.Print(b.String())
	},
}

func init() {
	//detectCmd.Flags().Float64VarP(&defaultDetectFlags.threshold, "threshold", "t", 50000, "画面变化检测阈值")
	detectCmd.Flags().Float64Var(&defaultDetectFlags.minDuration, "min_duration", 1, "最小持续时间，单位秒，默认为 1 秒。只有当画面人物出现的持续时间超过该时长，才认为是一个有效片段")
	detectCmd.Flags().Float64Var(&defaultDetectFlags.maxGap, "max_gap", 10, "最大间隔时间，单位秒，默认为 10 秒。两个有效片段之间的间隔时间超过该时长，会被认为是两个不同的片段，否则会被合并为一个片段")
	detectCmd.Flags().Float64Var(&defaultDetectFlags.sampleRate, "sample_rate", 1, "采样率，单位次/秒，默认为 1。每秒采样检测的帧数")
	rootCmd.AddCommand(detectCmd)
}

func detect() {
	if inputPath == "" {
		slog.Error("input path is empty")
		return
	}
	min := time.Duration(defaultDetectFlags.minDuration) * time.Second
	max := time.Duration(defaultDetectFlags.maxGap) * time.Second

	d, err := internal.NewYoloDetection("onnx", "", min, max, defaultDetectFlags.sampleRate)
	if err != nil {
		slog.Error("Error initializing yolo detection", "error", err)
		return
	}
	defer d.Close()

	video, err := gocv.VideoCaptureFile(inputPath)
	if err != nil {
		slog.Error("Error opening video file", "error", err)
		return
	}
	defer video.Close()

	slog.Info("人物检测开始", "视频路径", inputPath)
	segments, err := d.Detect(video)
	if err != nil {
		slog.Error("Error detecting", "error", err)
		return
	}
	if err = internal.ExtractSegments(inputPath, outputPath, segments); err != nil {
		slog.Error("Error extracting segments", "error", err)
		return
	}
	slog.Info("人物检测完成", "输出路径", outputPath)
}
