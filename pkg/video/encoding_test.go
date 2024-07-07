package video

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestEncoding(t *testing.T) {

	frame := Frame{
		{'#', '#', '#', '.', '%', '%'},
		{'#', '#', '#', '.', '%', '%'},
	}

	expectedRle := []byte{6, 3, '#', 1, '.', 2, '%', 3, '#', 1, '.', 2, '%'}

	rle := frame.RunLengthEncode()

	if string(rle) != string(expectedRle) {
		t.Log("ERROR: rle bad")
		t.FailNow()
	}

	t.Log("> passed rle check")

	chunks := ChunkFrameData(rle, 2, 1)

	if len(chunks) != 7 {
		t.FailNow()
	}

	fcc := NewFrameCatcher()

	rand.Shuffle(len(chunks), func(i, j int) {
		chunks[i], chunks[j] = chunks[j], chunks[i]
	})

	var outerFrame Frame
	for _, c := range chunks {
		data := c.Encode()
		f := fcc.Catch(data)
		if f != nil {
			outerFrame = f
		}
	}

	if len(fcc) != 0 {
		t.FailNow()
	}

	if frame.String() != outerFrame.String() {
		t.FailNow()
	}

	t.Log("[original]")
	t.Log(frame.String())
	t.Log("[after]")
	t.Log(outerFrame.String())

}

func TestEncodingAndDecoding(t *testing.T) {
	// Test frame with different sizes and content
	tests := []struct {
		name  string
		frame Frame
	}{
		{
			name: "Simple Frame",
			frame: Frame{
				{'#', '#', '#', '.', '%', '%'},
				{'#', '#', '#', '.', '%', '%'},
			},
		},
		{
			name: "Large Frame",
			frame: Frame{
				{'#', '#', '#', '.', '%', '%', '#', '#', '#', '.', '%', '%'},
				{'*', '^', '#', '.', '%', '%', '#', '#', '#', '.', '%', '%'},
				{'#', '#', '#', '.', '%', '%', '#', '#', '#', '.', '%', '%'},
				{'#', '#', '#', '.', '%', '%', '#', '#', '#', '.', '%', '%'},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode the frame to RLE
			rle := tt.frame.RunLengthEncode()

			// Decode the RLE into chunks
			chunks := ChunkFrameData(rle, 2, 1) // Assuming chunk size of 6 bytes

			// Shuffle chunks to simulate random transmission order
			rand.Shuffle(len(chunks), func(i, j int) {
				chunks[i], chunks[j] = chunks[j], chunks[i]
			})

			// Create a frame catcher
			fcc := NewFrameCatcher()

			// Reconstruct the frame from chunks
			var reconstructed Frame
			for _, c := range chunks {
				data := c.Encode()
				f := fcc.Catch(data)
				if f != nil {
					reconstructed = f
				}
			}

			// Compare reconstructed frame with original frame
			if !reflect.DeepEqual(tt.frame, reconstructed) {
				t.Errorf("Expected:\n%s\nGot:\n%s", tt.frame.String(), reconstructed.String())
			}
		})
	}
}

func TestRunLengthEncodeAndDecode(t *testing.T) {
	testCases := []struct {
		name  string
		frame Frame
	}{
		{
			name: "Simple Frame",
			frame: Frame{
				{'#', '#', '#', '.', '%', '%'},
				{'#', '#', '#', '.', '%', '%'},
			},
		},
		{
			name:  "Empty Frame",
			frame: Frame{},
		},
		{
			name: "Frame with Single Character",
			frame: Frame{
				{'#', '#', '#', '#'},
				{'#', '#', '#', '#'},
				{'#', '#', '#', '#'},
			},
		},
		{
			name: "Frame with Different Characters",
			frame: Frame{
				{'#', '.', '#', '.'},
				{'%', '%', '%', '%'},
				{'*', '*', '*', '*'},
			},
		},
		// Add more test cases as needed
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encoded := tc.frame.RunLengthEncode()
			decoded := RunLengthDecode(encoded)

			if !reflect.DeepEqual(tc.frame, decoded) {
				t.Errorf("expected decoded frame:\n%s\nbut got:\n%s", tc.frame.String(), decoded.String())
			}
		})
	}
}

func TestFrameChunkEncodingAndDecoding(t *testing.T) {
	testCases := []struct {
		name     string
		frame    FrameChunk
		expected FrameChunk
	}{
		{
			name: "Simple Frame Chunk",
			frame: FrameChunk{
				FrameId:        1,
				SequenceNumber: 0,
				TotalChunks:    2,
				Data:           []byte{0x01, 0x02, 0x03},
			},
			expected: FrameChunk{
				FrameId:        1,
				SequenceNumber: 0,
				TotalChunks:    2,
				Data:           []byte{0x01, 0x02, 0x03},
			},
		},
		{
			name: "Large Frame Chunk",
			frame: FrameChunk{
				FrameId:        2,
				SequenceNumber: 1,
				TotalChunks:    3,
				Data:           []byte{0x10, 0x20, 0x30, 0x40, 0x50},
			},
			expected: FrameChunk{
				FrameId:        2,
				SequenceNumber: 1,
				TotalChunks:    3,
				Data:           []byte{0x10, 0x20, 0x30, 0x40, 0x50},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encoded := tc.frame.Encode()

			var decoded FrameChunk
			err := decoded.Decode(encoded)
			if err != nil {
				t.Fatalf("error decoding frame chunk: %v", err)
			}

			if !reflect.DeepEqual(decoded, tc.expected) {
				t.Errorf("decoded frame chunk does not match expected\nGot:     %+v\nExpected:%+v", decoded, tc.expected)
			}
		})
	}

	// Test invalid input for decoding
	t.Run("Invalid Input for Decoding", func(t *testing.T) {
		var c FrameChunk
		err := c.Decode([]byte{0x01, 0x02, 0x03})
		if err == nil {
			t.Errorf("expected error for decoding invalid frame chunk, but got nil")
		}
		if err.Error() != "frame chunk too small" {
			t.Errorf("expected 'frame chunk too small' error, but got: %v", err)
		}
	})
}
