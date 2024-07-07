package video

import (
	"testing"
)

func TestRLE(t *testing.T) {
	t.Run("run length encoding", func(t *testing.T) {
		expects := append([]byte{9}, 1, 'z', 4, 'w', 1, 'd', 2, 'a', 1, 'b')
		frame := Frame{{'z', 'w', 'w', 'w', 'w', 'd', 'a', 'a', 'b'}}

		encoded := frame.RunLengthEncode()

		for i := range encoded {
			if encoded[i] != expects[i] {
				t.FailNow()
			}
		}

		expectStr := frame.String()
		output := RunLengthDecode(encoded)
		if expectStr != output.String() {
			t.FailNow()
		}

	})
}
