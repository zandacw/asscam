package audio

import (
	"testing"
	"time"
)

func TestAudioPlayerIntegration(t *testing.T) {
	audio, err := NewAudio()
	if err != nil {
		t.Fatalf("Error initializing audio: %v", err)
	}
	defer audio.Close()

	player, err := NewPlayer()
	if err != nil {
		t.Fatalf("Error initializing player: %v", err)
	}
	defer player.Close()

	go func() {
		err := audio.Start()
		if err != nil {
			t.Fatalf("Error starting audio stream: %v", err)
		}
	}()

	player.Start()

	duration := 5 * time.Second
	start := time.Now()
	for time.Since(start) < duration {
		select {
		case audioSeg := <-audio.Output:
			player.Input <- audioSeg
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}
