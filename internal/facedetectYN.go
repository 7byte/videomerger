package internal

import (
	"fmt"
	"image"
	"math"
	"time"

	"gocv.io/x/gocv"
)

type FaceDetectYN struct {
	detector       gocv.FaceDetectorYN
	mindur, maxgap time.Duration
	samplerate     float64
}

func NewFaceDetectYN(modelPath string, mindur, maxgap time.Duration, samplerate float64) (*FaceDetectYN, error) {
	return &FaceDetectYN{
		detector:   gocv.NewFaceDetectorYN(modelPath, "", image.Pt(200, 200)),
		mindur:     mindur,
		maxgap:     maxgap,
		samplerate: samplerate,
	}, nil
}

func (y *FaceDetectYN) Close() {
	y.detector.Close()
}

// Detect detects segments in a video stream
func (y *FaceDetectYN) Detect(vc *gocv.VideoCapture) ([]Segment, error) {

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
			fmt.Printf("Device closed")
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

		gocv.Resize(frame, &frame, image.Pt(int(width), int(height)), 0, 0, gocv.InterpolationLinear)

		y.detector.Detect(frame, &faces)

		if faces.Rows() >= 1 {
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

	}

	return segments, nil
}
