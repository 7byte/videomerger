package internal

import (
	_ "embed"
	"fmt"
	"image"
	"os"
	"time"

	"gocv.io/x/gocv"
)

//go:embed yolo11n.onnx
var yolo11n []byte

// Array of YOLOv8 class labels
var classes = []string{
	"person", "bicycle", "car", "motorcycle", "airplane", "bus", "train", "truck", "boat",
	"traffic light", "fire hydrant", "stop sign", "parking meter", "bench", "bird", "cat", "dog", "horse",
	"sheep", "cow", "elephant", "bear", "zebra", "giraffe", "backpack", "umbrella", "handbag", "tie",
	"suitcase", "frisbee", "skis", "snowboard", "sports ball", "kite", "baseball bat", "baseball glove",
	"skateboard", "surfboard", "tennis racket", "bottle", "wine glass", "cup", "fork", "knife", "spoon",
	"bowl", "banana", "apple", "sandwich", "orange", "broccoli", "carrot", "hot dog", "pizza", "donut",
	"cake", "chair", "couch", "potted plant", "bed", "dining table", "toilet", "tv", "laptop", "mouse",
	"remote", "keyboard", "cell phone", "microwave", "oven", "toaster", "sink", "refrigerator", "book",
	"clock", "vase", "scissors", "teddy bear", "hair drier", "toothbrush",
}

type YoloDetection struct {
	net            gocv.Net
	outputNames    []string
	mindur, maxgap time.Duration
}

func NewYoloDetection(kind, model string, min, max time.Duration) (*YoloDetection, error) {
	var bs []byte
	if model == "" {
		bs = yolo11n
	} else {
		var err error
		bs, err = os.ReadFile(model)
		if err != nil {
			return nil, fmt.Errorf("failed to read model file: %v", err)
		}
	}
	// TODO：支持其他模型
	net, err := gocv.ReadNetFromONNXBytes(bs)
	if err != nil {
		return nil, fmt.Errorf("failed to read YOLO model: %v", err)
	}
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

	return &YoloDetection{
		net:         net,
		outputNames: outputNames,
		mindur:      min,
		maxgap:      max,
	}, nil
}

func (y *YoloDetection) Close() {
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

func detect(net *gocv.Net, src *gocv.Mat, outputNames []string) bool {
	params := gocv.NewImageToBlobParams(ratio, image.Pt(640, 640), mean, swapRGB, gocv.MatTypeCV32F, gocv.DataLayoutNCHW, gocv.PaddingModeLetterbox, padValue)
	blob := gocv.BlobFromImageWithParams(*src, params)
	defer blob.Close()

	// feed the blob into the detector
	net.SetInput(blob, "")

	// run a forward pass thru the network
	probs := net.ForwardLayers(outputNames)

	// detected := false
	// for _, prob := range probs {
	// 	if !detected {
	// 		out := prob.Reshape(1, prob.Size()[1])
	// 		for i := 0; i < out.Rows(); i++ {
	// 			cols := out.Cols()
	// 			scoresCol := out.RowRange(i, i+1)
	// 			scores := scoresCol.ColRange(4, cols)
	// 			_, confidence, _, classIDPoint := gocv.MinMaxLoc(scores)
	// 			if confidence > 0.5 && classIDPoint.X == 0 { // 0 表示 "person" 类别
	// 				detected = true
	// 			}
	// 		}
	// 	}
	// 	prob.Close()
	// }
	// return detected

	defer func() {
		for _, prob := range probs {
			prob.Close()
		}
	}()

	boxes, confidences, classIds := performDetection(probs)
	if len(boxes) == 0 {
		return false
	}

	iboxes := params.BlobRectsToImageRects(boxes, image.Pt(src.Cols(), src.Rows()))
	indices := gocv.NMSBoxes(iboxes, confidences, scoreThreshold, nmsThreshold)

	for _, i := range indices {
		class := classes[classIds[i]]
		if class == "person" {
			return true
		}
	}

	return false
}

func performDetection(outs []gocv.Mat) ([]image.Rectangle, []float32, []int) {
	var classIds []int
	var confidences []float32
	var boxes []image.Rectangle

	// needed for yolov8
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
		}
	}

	return boxes, confidences, classIds
}

// Detect detects segments in a video stream
func (y *YoloDetection) Detect(vc *gocv.VideoCapture) ([]Segment, error) {
	frame := gocv.NewMat()
	defer frame.Close()

	var (
		segments   []Segment
		start      = -1.0
		frameIndex = 0
		fps        = vc.Get(gocv.VideoCaptureFPS) // 获取视频属性
		frameTime  = func() float64 { return float64(frameIndex) / fps }
	)
	for {
		if ok := vc.Read(&frame); !ok {
			break
		}
		if frame.Empty() {
			break
		}
		nowtime := frameTime()
		ok := detect(&y.net, &frame, y.outputNames)
		if ok {
			if len(segments) > 0 && nowtime-segments[len(segments)-1].End < y.maxgap.Seconds() {
				segments[len(segments)-1].End = nowtime
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
		frameIndex++
	}

	return segments, nil
}
