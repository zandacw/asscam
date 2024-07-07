package video

import (
	"image"

	"gocv.io/x/gocv"
)

// Define ASCII characters from dark to light
var asciiChars = " .:-=+*#%@"

// Function to convert frame to ASCII art
func frameToAscii(frame gocv.Mat, width int) [][]rune {
	// Convert the frame to grayscale
	defer frame.Close()
	gray := gocv.NewMat()
	defer gray.Close()
	gocv.CvtColor(frame, &gray, gocv.ColorBGRToGray)

	// Resize the frame to a smaller size for better ASCII art visualization
	aspectRatio := float64(gray.Rows()) / float64(gray.Cols())
	height := int(float64(width) * aspectRatio * 0.75)

	resized := gocv.NewMat()
	defer resized.Close()
	gocv.Resize(gray, &resized, image.Point{X: width, Y: height}, 0, 0, gocv.InterpolationArea)

	rows, cols := resized.Rows(), resized.Cols()

	runeFrame := make([][]rune, rows)
	for rowIdx := range rows {
		runeFrame[rowIdx] = make([]rune, cols)
	}

	// Map each pixel to an ASCII character based on intensity
	for rowIdx := 0; rowIdx < rows; rowIdx++ {
		for colIdx := 0; colIdx < cols; colIdx++ {
			pixel := resized.GetUCharAt(rowIdx, colIdx)
			index := int(float64(pixel) / 255.0 * float64(len(asciiChars)-1))
			runeFrame[rowIdx][colIdx] = rune(asciiChars[index])
		}
	}

	return runeFrame
}

func (f Frame) String() string {
	var output string
	for rowIdx := range f {
		row := f[rowIdx]
		output += string(row)
		output += "\n"
	}
	return output
}
