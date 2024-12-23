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
}

var defaultDetectFlags detectFlags

var detectCmd = &cobra.Command{
	Use:   "detect",
	Short: "Detect face in video",
	Long:  "Detect face in video",
	Run: func(cmd *cobra.Command, args []string) {
		detect()
	},
}

func init() {
	detectCmd.Flags().Float64VarP(&defaultDetectFlags.threshold, "threshold", "t", 50000, "画面变化检测阈值")
	detectCmd.Flags().Float64Var(&defaultDetectFlags.minDuration, "min_duration", 1, "最小变化持续时间")
	detectCmd.Flags().Float64Var(&defaultDetectFlags.maxGap, "max_gap", 5, "最大间隔时间")
	rootCmd.AddCommand(detectCmd)
}

func detect() {
	if inputPath == "" {
		slog.Error("input path is empty")
		return
	}
	min := time.Duration(defaultDetectFlags.minDuration) * time.Second
	max := time.Duration(defaultDetectFlags.maxGap) * time.Second

	d, err := internal.NewYoloDetection("onnx", "", min, max)
	if err != nil {
		slog.Error("Error initializing yolo detection: %v\n", err)
		return
	}

	video, err := gocv.VideoCaptureFile(inputPath)
	if err != nil {
		slog.Error("Error opening video file: %v\n", err)
		return
	}
	defer video.Close()

	segments, err := d.Detect(video)
	if err != nil {
		slog.Error("Error detecting: %v\n", err)
		return
	}
	if err = internal.ExtractSegments(inputPath, segments); err != nil {
		slog.Error("Error extracting segments: %v\n", err)
		return
	}
	slog.Info("Segments extracted successfully")
}

// func detect() {
// 	// TODO
// 	if defaultDetectFlags.inputPath == "" {
// 		slog.Error("input path is empty")
// 		return
// 	}

// 	processor, err := NewVideoProcessor(
// 		defaultDetectFlags.inputPath,
// 		defaultDetectFlags.threshold,
// 		defaultDetectFlags.minDuration,
// 		defaultDetectFlags.netCfg,
// 		defaultDetectFlags.netWeights)
// 	if err != nil {
// 		slog.Error("Error initializing video processor: %v\n", err)
// 		return
// 	}

// 	segments, err := processor.DetectChangesAndPeople()
// 	if err != nil {
// 		slog.Error("Error detecting changes and people: %v\n", err)
// 		return
// 	}

// 	if err := processor.ExtractSegments(segments); err != nil {
// 		slog.Error("Error extracting segments: %v\n", err)
// 		return
// 	}

// 	slog.Info("Segments extracted successfully")
// }

// // VideoDetectProcessor 视频处理结构体
// type VideoDetectProcessor struct {
// 	videoPath     string
// 	threshold     float64 // 画面变化检测阈值
// 	minDuration   float64 // 最小变化持续时间
// 	fps           float64
// 	width, height int
// 	net           gocv.Net
// }

// // NewVideoProcessor 初始化处理器
// func NewVideoProcessor(videoPath string, threshold, minDuration float64, netCfg, netWeights string) (*VideoDetectProcessor, error) {
// 	net := gocv.ReadNet(netWeights, netCfg)
// 	if net.Empty() {
// 		return nil, fmt.Errorf("failed to load YOLO model")
// 	}

// 	video, err := gocv.VideoCaptureFile(videoPath)
// 	if err != nil {
// 		return nil, fmt.Errorf("error opening video file: %v", err)
// 	}
// 	defer video.Close()

// 	return &VideoDetectProcessor{
// 		videoPath:   videoPath,
// 		threshold:   threshold,
// 		minDuration: minDuration,
// 		fps:         video.Get(gocv.VideoCaptureFPS),
// 		width:       int(video.Get(gocv.VideoCaptureFrameWidth)),
// 		height:      int(video.Get(gocv.VideoCaptureFrameHeight)),
// 		net:         net,
// 	}, nil
// }

// // DetectChangesAndPeople 检测画面变化和人物
// func (vp *VideoDetectProcessor) DetectChangesAndPeople() ([][2]float64, error) {
// 	video, err := gocv.VideoCaptureFile(defaultDetectFlags.inputPath)
// 	if err != nil {
// 		slog.Error("opening video file", "error", err)
// 		return nil, err
// 	}
// 	defer video.Close()

// 	prevFrame := gocv.NewMat()
// 	defer prevFrame.Close()

// 	currentFrame := gocv.NewMat()
// 	defer currentFrame.Close()

// 	diffFrame := gocv.NewMat()
// 	defer diffFrame.Close()

// 	var segments [][2]float64
// 	segmentStart := -1.0
// 	frameIndex := 0
// 	frameTime := func() float64 { return float64(frameIndex) / vp.fps }

// 	for {
// 		if ok := video.Read(&currentFrame); !ok || currentFrame.Empty() {
// 			break
// 		}

// 		if frameIndex == 0 {
// 			currentFrame.CopyTo(&prevFrame)
// 			frameIndex++
// 			continue
// 		}

// 		// 计算帧差
// 		gocv.AbsDiff(currentFrame, prevFrame, &diffFrame)
// 		gocv.CvtColor(diffFrame, &diffFrame, gocv.ColorBGRToGray)
// 		_, diffValue, _, _ := gocv.Sum(diffFrame).Val1

// 		// YOLO 人物检测
// 		peopleDetected := vp.DetectPeople(currentFrame)

// 		// 判断条件
// 		if diffValue > vp.threshold || peopleDetected {
// 			if segmentStart == -1 {
// 				segmentStart = frameTime()
// 			}
// 		} else if segmentStart != -1 {
// 			if frameTime()-segmentStart >= vp.minDuration {
// 				segments = append(segments, [2]float64{segmentStart, frameTime()})
// 			}
// 			segmentStart = -1
// 		}

// 		currentFrame.CopyTo(&prevFrame)
// 		frameIndex++
// 	}

// 	return segments, nil
// }

// // DetectPeople 使用 YOLO 模型检测人物
// func (vp *VideoDetectProcessor) DetectPeople(frame gocv.Mat) bool {
// 	blob := gocv.BlobFromImage(frame, 1/255.0, gocv.NewSize(416, 416), gocv.NewScalar(0, 0, 0, 0), true, false)
// 	vp.net.SetInput(blob, "")
// 	detections := vp.net.ForwardLayers(vp.net.GetUnconnectedOutLayersNames())

// 	for _, det := range detections {
// 		for i := 0; i < det.Total(); i += 85 {
// 			classID := int(det.GetFloatAt(i + 5))
// 			if classID == 0 { // 0 通常表示 "person" 类别
// 				return true
// 			}
// 		}
// 	}
// 	return false
// }
