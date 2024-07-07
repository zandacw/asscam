package audio

import (
	"encoding/binary"
	"io"
	"math"
	"time"

	"github.com/ebitengine/oto/v3"
	"github.com/gordonklaus/portaudio"
)

const (
	sampleRate        = 16000
	framesPerBuffer   = 120 // Approx 1/30 second of audio
	numInputChannels  = 1
	numOutputChannels = 0
	volumeScaleFactor = 1
)

type Audio struct {
	stream *portaudio.Stream
	Output chan []byte
	buffer []float32
}

type Player struct {
	player *oto.Player
	Input  chan []byte
}

func (p *Player) Close() {
	p.player.Close()
}

func (p *Player) Start() {
	p.player.Play()
}

func (a *Audio) Close() {
	a.stream.Stop()
	a.stream.Close()
	portaudio.Terminate()
}

func (a *Audio) Start() error {
	for {
		err := a.stream.Read()
		if err != nil {
			return err
		}
		pcm := float32ToPCM(a.buffer)
		a.Output <- pcm
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
		Output: make(chan []byte),
		stream: stream,
	}, nil

}
func NewPlayer() (*Player, error) {
	op := &oto.NewContextOptions{
		SampleRate:   sampleRate,
		ChannelCount: 1,
		Format:       oto.FormatSignedInt16LE,
		BufferSize:   time.Duration(33 * time.Millisecond),
	}
	otoCtx, readyChan, err := oto.NewContext(op)
	if err != nil {
		return nil, err
	}
	<-readyChan

	c := make(chan []byte)
	reader := NewAudioBufferReader(c)
	player := otoCtx.NewPlayer(reader)

	return &Player{
		player: player,
		Input:  c,
	}, nil

}

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

func float32ToPCM(buffer []float32) []byte {
	pcm := make([]byte, len(buffer)*2) // 2 bytes per sample for int16
	for i, sample := range buffer {
		sample = sample * volumeScaleFactor
		sample = clampBetweenOne(sample)
		intSample := int16(sample * math.MaxInt16)
		binary.LittleEndian.PutUint16(pcm[i*2:], uint16(intSample))
	}
	return pcm
}

func clampBetweenOne(sample float32) float32 {
	if sample > 1.0 {
		sample = 1.0
	} else if sample < -1.0 {
		sample = -1.0
	}
	return sample
}
