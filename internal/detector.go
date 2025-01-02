package internal

import (
	"fmt"
	"image"
	"log/slog"
	"math"
	"time"

	"gocv.io/x/gocv"
)

// Detector defines the interface for detecting person in a video stream
type Detector interface {
	Detect(*gocv.VideoCapture) ([]Segment, error)
	Close()
}

type YoloDetector struct {
	net            gocv.Net
	outputNames    []string
	mindur, maxgap time.Duration
	samplerate     float64
}

func NewYoloDetector(modelPath string, mindur, maxgap time.Duration, samplerate float64) (*YoloDetector, error) {
	net := gocv.ReadNetFromONNX(modelPath)
	if net.Empty() {
		return nil, fmt.Errorf("failed to load YOLO model")
	}

	// TODO：支持配置
	backend := gocv.NetBackendDefault
	target := gocv.NetTargetCPU
	net.SetPreferableBackend(gocv.NetBackendType(backend))
	net.SetPreferableTarget(gocv.NetTargetType(target))

	outputNames := getOutputNames(&net)
	if len(outputNames) == 0 {
		return nil, fmt.Errorf("failed to read output layer names")
	}

	return &YoloDetector{
		net:         net,
		outputNames: outputNames,
		mindur:      mindur,
		maxgap:      maxgap,
		samplerate:  samplerate,
	}, nil
}

func (y *YoloDetector) Close() {
	y.net.Close()
}

func getOutputNames(net *gocv.Net) []string {
	var outputLayers []string
	for _, i := range net.GetUnconnectedOutLayers() {
		layer := net.GetLayer(i)
		layerName := layer.GetName()
		if layerName != "_input" {
			outputLayers = append(outputLayers, layerName)
		}
	}
	return outputLayers
}

var (
	ratio    = 0.003921568627
	mean     = gocv.NewScalar(0, 0, 0, 0)
	swapRGB  = false
	padValue = gocv.NewScalar(144.0, 0, 0, 0)

	scoreThreshold float32 = 0.5
	nmsThreshold   float32 = 0.4
)

func (y *YoloDetector) detect(net *gocv.Net, src *gocv.Mat, outputNames []string) bool {
	params := gocv.NewImageToBlobParams(ratio, image.Pt(640, 640), mean, swapRGB, gocv.MatTypeCV32F, gocv.DataLayoutNCHW, gocv.PaddingModeLetterbox, padValue)
	blob := gocv.BlobFromImageWithParams(*src, params)
	defer blob.Close()

	// feed the blob into the detector
	net.SetInput(blob, "")

	// run a forward pass thru the network
	probs := net.ForwardLayers(outputNames)

	defer func() {
		for _, prob := range probs {
			prob.Close()
		}
	}()

	boxes, confidences, classIds := y.performDetection(probs)
	if len(boxes) == 0 {
		return false
	}

	iboxes := params.BlobRectsToImageRects(boxes, image.Pt(src.Cols(), src.Rows()))
	indices := gocv.NMSBoxes(iboxes, confidences, scoreThreshold, nmsThreshold)

	for _, i := range indices {
		if classIds[i] == 0 {
			return true
		}
	}

	return false
}

func (y *YoloDetector) performDetection(outs []gocv.Mat) ([]image.Rectangle, []float32, []int) {
	var classIds []int
	var confidences []float32
	var boxes []image.Rectangle

	gocv.TransposeND(outs[0], []int{0, 2, 1}, &outs[0])
	for _, out := range outs {
		out = out.Reshape(1, out.Size()[1])
		for i := 0; i < out.Rows(); i++ {
			cols := out.Cols()
			scoresCol := out.RowRange(i, i+1)

			scores := scoresCol.ColRange(4, cols)
			_, confidence, _, classIDPoint := gocv.MinMaxLoc(scores)

			if confidence > 0.5 {
				centerX := out.GetFloatAt(i, cols)
				centerY := out.GetFloatAt(i, cols+1)
				width := out.GetFloatAt(i, cols+2)
				height := out.GetFloatAt(i, cols+3)

				left := centerX - width/2
				top := centerY - height/2
				right := centerX + width/2
				bottom := centerY + height/2
				classIds = append(classIds, classIDPoint.X)
				confidences = append(confidences, float32(confidence))

				boxes = append(boxes, image.Rect(int(left), int(top), int(right), int(bottom)))
			}
			scores.Close()
			scoresCol.Close()
		}
		out.Close()
	}

	return boxes, confidences, classIds
}

// Detect detects segments in a video stream
func (y *YoloDetector) Detect(vc *gocv.VideoCapture) ([]Segment, error) {
	frame := gocv.NewMat()
	defer frame.Close()

	var (
		segments      []Segment
		start         = -1.0
		frameIndex    = 0
		fps           = vc.Get(gocv.VideoCaptureFPS)                        // 获取视频属性
		frameTime     = func() float64 { return float64(frameIndex) / fps } // 计算帧时间
		frameInterval = int(math.Round(fps / y.samplerate))                 // 每多少帧执行一次检测，根据采样率计算，四舍五入取整
	)
	slog.Debug("检测视频参数", "视频帧率", fps, "采样率", y.samplerate, "帧检测间隔", frameInterval, "最小持续时间", y.mindur, "最大间隔时间", y.maxgap)
	for {
		if ok := vc.Read(&frame); !ok {
			break
		}
		if frame.Empty() {
			break
		}

		nowtime := frameTime()
		frameIndex++
		if frameInterval > 0 && frameIndex%frameInterval != 0 {
			continue
		}

		ok := y.detect(&y.net, &frame, y.outputNames)
		if ok {
			lastIdx := len(segments) - 1
			if len(segments) > 0 && nowtime-segments[lastIdx].End < y.maxgap.Seconds() {
				segments[lastIdx].End = nowtime
			} else if start < 0 {
				start = nowtime
			}
		} else {
			if start >= 0 {
				if nowtime-start >= y.mindur.Seconds() {
					segments = append(segments, Segment{Start: start, End: nowtime})
				}
				start = -1.0
			}
		}
	}

	return segments, nil
}

type YunetDetector struct {
	detector       gocv.FaceDetectorYN
	mindur, maxgap time.Duration
	samplerate     float64
}

func NewYunetDetector(modelPath string, mindur, maxgap time.Duration, samplerate float64) (*YunetDetector, error) {
	return &YunetDetector{
		detector:   gocv.NewFaceDetectorYN(modelPath, "", image.Pt(640, 640)),
		mindur:     mindur,
		maxgap:     maxgap,
		samplerate: samplerate,
	}, nil
}

func (y *YunetDetector) Close() {
	y.detector.Close()
}

// Detect detects segments in a video stream
func (y *YunetDetector) Detect(vc *gocv.VideoCapture) ([]Segment, error) {
	// prepare image matrix
	frame := gocv.NewMat()
	defer frame.Close()

	// prepare faces matrix
	faces := gocv.NewMat()
	defer faces.Close()

	var (
		segments      []Segment
		start         = -1.0
		frameIndex    = 0
		fps           = vc.Get(gocv.VideoCaptureFPS) // 获取视频属性
		width         = vc.Get(gocv.VideoCaptureFrameWidth)
		height        = vc.Get(gocv.VideoCaptureFrameHeight)
		frameTime     = func() float64 { return float64(frameIndex) / fps } // 计算帧时间
		frameInterval = int(math.Round(fps / y.samplerate))                 // 每多少帧执行一次检测，根据采样率计算，四舍五入取整
	)

	y.detector.SetInputSize(image.Pt(int(width), int(height)))
	for {
		if ok := vc.Read(&frame); !ok {
			break
		}
		if frame.Empty() {
			break
		}

		nowtime := frameTime()
		frameIndex++
		if frameInterval > 0 && frameIndex%frameInterval != 0 {
			continue
		}

		//gocv.Resize(frame, &frame, image.Pt(int(width), int(height)), 0, 0, gocv.InterpolationLinear)
		y.detector.Detect(frame, &faces)
		if faces.Rows() >= 1 {
			lastIdx := len(segments) - 1
			if len(segments) > 0 && nowtime-segments[lastIdx].End < y.maxgap.Seconds() {
				segments[lastIdx].End = nowtime
			} else if start < 0 {
				start = nowtime
			}
		} else {
			if start >= 0 {
				if nowtime-start >= y.mindur.Seconds() {
					segments = append(segments, Segment{Start: start, End: nowtime})
				}
				start = -1.0
			}
		}

	}

	return segments, nil
}
