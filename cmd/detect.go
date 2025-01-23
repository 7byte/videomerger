package cmd

import (
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/7byte/videomerger/internal"
	"github.com/spf13/cobra"
	"gocv.io/x/gocv"
)

type detectFlags struct {
	minSize     int32
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
	detectCmd.Flags().Int32VarP(&dFlags.minSize, "min_size", "s", 1024*1024, "最小视频大小（字节），小于该大小的视频文件将被忽略，默认为1MB")
	detectCmd.Flags().StringVarP(&dFlags.modelType, "model_type", "t", "yolo", "检测模型：yolo、yunet")
	detectCmd.Flags().StringVarP(&dFlags.modelPath, "model_path", "m", "", "模型路径，模型文件需要与检测类别对应")
	detectCmd.Flags().Float64VarP(&dFlags.minDuration, "min_duration", "d", 5.0, "最小持续时间（秒）。只有当画面人物出现的持续时间超过该时长，才认为是一个有效片段")
	detectCmd.Flags().Float64VarP(&dFlags.maxGap, "max_gap", "g", 20.0, "最大间隔时间（秒）。两个有效片段之间的间隔时间超过该时长，会被认为是两个不同的片段，否则会被合并为一个片段")
	detectCmd.Flags().Float64VarP(&dFlags.sampleRate, "sample_rate", "r", 1.0, "采样率（次/秒）。每秒采样检测的帧数")
	rootCmd.AddCommand(detectCmd)
}

// 常见视频文件扩展名
var videoExtensions = []string{".mp4"}

func detect() {
	if inputPath == "" {
		slog.Error("需要指定视频文件所在路径")
		return
	}

	if dFlags.modelPath == "" {
		slog.Error("需要指定模型文件")
		return
	}

	// 检查输出目录是否存在，目录不存在时创建目录
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		err := os.MkdirAll(outputPath, os.ModePerm)
		if err != nil {
			slog.Error("Failed to create directory", "error", err)
			return
		}
		slog.Info("Directory created:", "path", outputPath)
	} else if err != nil {
		slog.Error("Error checking directory", "error", err)
		return
	}

	min := time.Duration(dFlags.minDuration) * time.Second
	max := time.Duration(dFlags.maxGap) * time.Second

	var detector internal.Detector
	switch dFlags.modelType {
	case "yolo":
		d, err := internal.NewYoloDetector(dFlags.modelPath, min, max, dFlags.sampleRate)
		if err != nil {
			slog.Error("Error initializing yolo detection", "error", err)
			return
		}
		detector = d
	case "yunet":
		d, err := internal.NewYunetDetector(dFlags.modelPath, min, max, dFlags.sampleRate)
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

	err := filepath.WalkDir(inputPath, func(path string, d os.DirEntry, e error) error {
		if d.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		if !slices.Contains(videoExtensions, ext) {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			slog.Error("获取文件信息失败", "文件", path, "错误原因", err)
			return err
		}
		if info.Size() < int64(dFlags.minSize) {
			slog.Info("视频文件太小，忽略", "文件", path)
			return nil
		}
		video, err := gocv.VideoCaptureFile(path)
		if err != nil {
			slog.Error("打开视频文件失败", "文件", path, "错误原因", err)
			return err
		}
		defer video.Close()

		slog.Info("人物检测开始", "文件", path)
		segments, err := detector.Detect(video)
		if err != nil {
			slog.Error("检测失败", "错误原因", err)
			return err
		}
		if err = internal.ExtractSegments(path, outputPath, strings.TrimSuffix(d.Name(), ext), segments); err != nil {
			slog.Error("提取视频片段失败", "错误原因", err)
			return err
		}
		return nil
	})
	if err != nil {
		return
	}

	slog.Info("人物检测完成", "路径", outputPath)
}
