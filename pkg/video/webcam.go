package video

import (
	"errors"

	"gocv.io/x/gocv"
)

func startWebcam() (*gocv.VideoCapture, error) {

	webcam, err := gocv.VideoCaptureDevice(0)
	if err != nil {
		return nil, err
	}

	if !webcam.IsOpened() {
		return nil, errors.New("Webcam could not be opened")
	}

	return webcam, nil
}
