package internal

import "gocv.io/x/gocv"

type Detector interface {
	Detect(*gocv.VideoCapture) ([]Segment, error)
	Close()
}
