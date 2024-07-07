package main

import (
	"context"
	"io"
	"math"

	"github.com/gordonklaus/portaudio"
)

const seconds = 1

type AudioBufferReader struct {
	buffer []byte
	stream <-chan []byte
}

func NewAudioBufferReader(stream <-chan []byte) *AudioBufferReader {
	return &AudioBufferReader{
		stream: stream,
	}
}

func (r *AudioBufferReader) Read(b []byte) (int, error) {

	if len(r.buffer) == 0 {
		data, ok := <-r.stream
		if !ok {
			return 0, io.EOF
		}
		// var loud bool
		// for i := range data {
		// 	if data[i] > 0.3 {
		// 		loud = true
		// 	}
		// }
		// if loud {
		// 	fmt.Println("[IN] loud")
		// }
		r.buffer = data
	}

	n := copy(b, r.buffer)

	r.buffer = r.buffer[n:]

	return n, nil

}

//	func float32ToByte(buffer []float32) []byte {
//		byteBuffer := make([]byte, len(buffer)*2)
//		for i, sample := range buffer {
//			intSample := int16(sample * 32767)
//			byteBuffer[2*i] = byte(intSample)
//			byteBuffer[2*i+1] = byte(intSample >> 8)
//		}
//		return byteBuffer
//	}
func float32ToByte(buffer []float32) []byte {
	byteBuffer := make([]byte, len(buffer)*4)
	for i, sample := range buffer {
		bits := math.Float32bits(sample)
		byteBuffer[4*i] = byte(bits)
		byteBuffer[4*i+1] = byte(bits >> 8)
		byteBuffer[4*i+2] = byte(bits >> 16)
		byteBuffer[4*i+3] = byte(bits >> 24)
	}
	return byteBuffer
}

func normalizeFloat32Buffer(buffer []float32) {
	maxAbs := float32(0)

	// Find the maximum absolute value in the buffer
	for _, sample := range buffer {
		if abs := float32(math.Abs(float64(sample))); abs > maxAbs {
			maxAbs = abs
		}
	}

	// Normalize each sample to the range [-1.0, 1.0]
	if maxAbs > 0 {
		scaleFactor := float32(1.0 / maxAbs)
		for i := range buffer {
			buffer[i] *= scaleFactor
		}
	}
}

func audioLoop(ctx context.Context, streamC chan []float32) {
	portaudio.Initialize()
	defer portaudio.Terminate()

	stream, err := portaudio.OpenDefaultStream(1, 0, sampleRate, bufferSize, func(in []float32) {
		for i := range in {
			in[i] *= 50
		}

		buffer := make([]float32, bufferSize)
		copy(buffer, in)

		select {
		case streamC <- buffer:
		case <-ctx.Done():
			return
		}
	})

	if err != nil {
		panic(err)
	}

	err = stream.Start()
	if err != nil {
		panic(err)
	}
	defer stream.Stop()

	<-ctx.Done()

}
