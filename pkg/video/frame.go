package video

import (
	"context"
	"fmt"

	"gocv.io/x/gocv"
)

type update struct {
	rowIdx int
	colIdx int
	char   rune
}

type Frame [][]rune
type updates []update

func Start(ctx context.Context, width int) (chan Frame, error) {

	webcam, err := startWebcam()
	if err != nil {
		return nil, err
	}

	frameC := make(chan Frame)
	go func() {
		defer webcam.Close()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				screenMaterial := gocv.NewMat()
				if ok := webcam.Read(&screenMaterial); !ok {
					return
				}
				if screenMaterial.Empty() {
					continue
				}
				frame := frameToAscii(screenMaterial, width)
				frameC <- frame
			}
		}
	}()
	return frameC, nil
}

func (newFrame Frame) Display(oldFrame Frame) {

	if updates := newFrame.diff(oldFrame); updates != nil {
		updates.do()
	} else {
		newFrame.show()

	}

}

func (f Frame) show() {
	ClearScreen()
	fmt.Print(f)
}
