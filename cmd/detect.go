package cmd

import (
	"log/slog"
	"time"

	"github.com/7byte/videomerger/internal"
	"github.com/spf13/cobra"
	"gocv.io/x/gocv"
)

type detectFlags struct {
	modelType   string
	modelPath   string
	minDuration float64
	maxGap      float64
	sampleRate  float64
}

var dFlags detectFlags

var detectCmd = &cobra.Command{
	Use:   "detect",
	Short: "识别并提取视频中出现人物的片段",
	Long: `识别并提取视频中出现人物的片段
支持的检测模型：
    yolo：使用yolo模型检测人物，模型文件下载地址：https://colab.research.google.com/github/ultralytics/ultralytics/blob/main/examples/tutorial.ipynb
    yunet：使用face_detection_yunet模型检测人物，模型文件下载地址：https://github.com/opencv/opencv_zoo/tree/main/models/face_detection_yunet`,
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
	detectCmd.Flags().StringVarP(&dFlags.modelType, "model_type", "m", "yolo", "检测模型：yolo、yunet")
	detectCmd.Flags().StringVarP(&dFlags.modelPath, "model_path", "p", "", "模型路径，模型文件需要与检测类别对应")
	detectCmd.Flags().Float64VarP(&dFlags.minDuration, "min_duration", "d", 1.0, "最小持续时间（秒）。只有当画面人物出现的持续时间超过该时长，才认为是一个有效片段")
	detectCmd.Flags().Float64VarP(&dFlags.maxGap, "max_gap", "g", 10.0, "最大间隔时间（秒）。两个有效片段之间的间隔时间超过该时长，会被认为是两个不同的片段，否则会被合并为一个片段")
	detectCmd.Flags().Float64VarP(&dFlags.sampleRate, "sample_rate", "r", 1.0, "采样率（次/秒）。每秒采样检测的帧数")
	rootCmd.AddCommand(detectCmd)
}

func detect() {
	if inputPath == "" {
		slog.Error("input path is empty")
		return
	}
	min := time.Duration(dFlags.minDuration) * time.Second
	max := time.Duration(dFlags.maxGap) * time.Second

	var detector internal.Detector
	switch dFlags.modelType {
	case "yolo":
		d, err := internal.NewYoloDetection(dFlags.modelPath, min, max, dFlags.sampleRate)
		if err != nil {
			slog.Error("Error initializing yolo detection", "error", err)
			return
		}
		detector = d
	case "yunet":
		d, err := internal.NewFaceDetectYN(dFlags.modelPath, min, max, dFlags.sampleRate)
		if err != nil {
			slog.Error("Error initializing yunet detection", "error", err)
			return
		}
		detector = d
	default:
		slog.Error("Unknown detect class", "detect_class", dFlags.modelType)
		return
	}
	defer detector.Close()

	video, err := gocv.VideoCaptureFile(inputPath)
	if err != nil {
		slog.Error("Error opening video file", "error", err)
		return
	}
	defer video.Close()

	slog.Info("人物检测开始", "视频路径", inputPath)
	segments, err := detector.Detect(video)
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
