package main

import (
	"context"
	"fmt"
	_ "image/jpeg"
	"os"
	"os/signal"
	"syscall"

	"github.com/langlandsbrogram/asscam/pkg/video"
)

const videoWidth = 300

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		cancel()
	}()

	// audioStreamC := make(chan []float32, 20)
	//
	// op := &oto.NewContextOptions{
	// 	SampleRate:   sampleRate,
	// 	ChannelCount: 1,
	// 	Format:       oto.FormatFloat32LE,
	// 	BufferSize:   time.Duration(float64(bufferSize) / float64(sampleRate) * float64(time.Second)),
	// }
	//
	// otoCtx, readyChan, err := oto.NewContext(op)
	// if err != nil {
	// 	panic("oto.NewContext failed: " + err.Error())
	// }
	// <-readyChan
	//
	// reader := NewAudioBufferReader(audioStreamC)
	// player := otoCtx.NewPlayer(reader)
	// defer player.Close()
	//
	// go audioLoop(ctx, audioStreamC)
	//
	// player.Play()
	//
	// <-ctx.Done()
	frames, err := video.Start(ctx, videoWidth)
	if err != nil {
		panic(err)
	}

	// startTime := time.Now()
	var totalFrames int
	for {
		select {
		case frame := <-frames:
			l := frame.RunLengthEncode()
			fmt.Println(len(l))
		case <-ctx.Done():
			return
		}
		totalFrames++
		// elapsed := time.Since(startTime)
		// elapsedSecs := elapsed.Seconds()
		// fmt.Println("frame rate: %d", float64(totalFrames)/elapsedSecs)
	}

}
