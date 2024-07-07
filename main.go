package main

import (
	"context"
	"encoding/binary"
	"fmt"
	_ "image/jpeg"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ebitengine/oto/v3"
	"github.com/gordonklaus/portaudio"
)

const videoWidth = 300

const (
	sampleRate        = 16000
	framesPerBuffer   = 533 // Approx 1/30 second of audio
	numInputChannels  = 1
	numOutputChannels = 0
	bufferSize        = 512
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		cancel()
	}()

	audioStreamC := make(chan []byte, 20)
	//
	op := &oto.NewContextOptions{
		SampleRate:   sampleRate,
		ChannelCount: 1,
		Format:       oto.FormatSignedInt16LE,
		BufferSize:   time.Duration(33 * time.Millisecond),
	}
	otoCtx, readyChan, err := oto.NewContext(op)
	if err != nil {
		panic("oto.NewContext failed: " + err.Error())
	}
	<-readyChan
	//
	reader := NewAudioBufferReader(audioStreamC)
	player := otoCtx.NewPlayer(reader)
	defer player.Close()

	// go audioLoop(ctx, audioStreamC)

	player.Play()
	//
	// <-ctx.Done()

	audio, err := NewAudio()
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
	defer audio.Close()

	go audio.Read()

	for {
		select {
		case audioData := <-audio.output:
			fmt.Println(audioData)
		case <-ctx.Done():
			return
		}
	}

}

type Audio struct {
	stream *portaudio.Stream
	output chan []byte
	buffer []float32
}

func (a *Audio) Close() {
	a.stream.Stop()
	a.stream.Close()
	portaudio.Terminate()
}

func (a *Audio) Read() error {
	for {
		err := a.stream.Read()
		if err != nil {
			return err
		}
		pcm := float32ToPCM(a.buffer)
		a.output <- pcm
	}
}

func NewAudio() (*Audio, error) {
	err := portaudio.Initialize()
	if err != nil {
		return nil, err
	}

	buffer := make([]float32, framesPerBuffer)
	stream, err := portaudio.OpenDefaultStream(
		numInputChannels,
		numOutputChannels,
		sampleRate,
		framesPerBuffer,
		buffer,
	)
	if err != nil {
		return nil, err
	}

	err = stream.Start()
	if err != nil {
		return nil, err
	}

	return &Audio{
		buffer: buffer,
		output: make(chan []byte),
		stream: stream,
	}, nil

}

func float32ToPCM(buffer []float32) []byte {
	pcm := make([]byte, len(buffer)*2) // 2 bytes per sample for int16
	for i, sample := range buffer {
		sample = sample * 50
		// Clamp sample to the range [-1.0, 1.0]
		if sample > 1.0 {
			sample = 1.0
		} else if sample < -1.0 {
			sample = -1.0
		}
		// Convert to int16
		intSample := int16(sample * math.MaxInt16)
		binary.LittleEndian.PutUint16(pcm[i*2:], uint16(intSample))
	}
	return pcm
}
