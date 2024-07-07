package video

import (
	"encoding/binary"
	"errors"
	"fmt"
	"sort"
	"strings"
)

func (f Frame) RunLengthEncode() []byte {

	var tally uint8
	var prevChar rune
	var output []byte

	if len(f) == 0 || len(f[0]) == 0 {
		// Empty frame case
		return output
	}

	appendToOutput := func() {
		output = append(output, tally)
		output = append(output, byte(prevChar))
	}

	numCols := uint8(len(f[0]))
	output = append(output, numCols)

	for rowIdx := range f {
		for colIdx := range f[rowIdx] {
			char := f[rowIdx][colIdx]

			if char == prevChar || prevChar == 0 {
				prevChar = char
				tally++
				if tally == 255 {
					appendToOutput()
					tally = 1
				}
			} else {
				appendToOutput()
				prevChar = char
				tally = 1
			}
		}
	}

	if tally > 0 {
		appendToOutput()
	}

	return output
}

func RunLengthDecode(data []byte) Frame {

	if len(data) == 0 {
		return Frame{}
	}

	cols := int(data[0])

	// fmt.Printf("read columns: %d\n", cols)
	data = data[1:]

	s := ""
	for i := 0; i < len(data); i += 2 {
		n := data[i]
		char := data[i+1]
		s += strings.Repeat(string(char), int(n))
	}

	var output Frame
	for i, c := range s {
		col := i % cols
		if col == 0 {
			output = append(output, make([]rune, cols))
		}
		row := i / cols
		output[row][col] = c
	}
	return output
}

type FrameChunk struct {
	FrameId        uint32
	SequenceNumber uint8
	TotalChunks    uint8
	Data           []byte
}

func (c *FrameChunk) Encode() []byte {
	size := 4 + 1 + 1 + len(c.Data)
	buf := make([]byte, size)
	binary.LittleEndian.PutUint32(buf[:4], c.FrameId)
	buf[4] = c.SequenceNumber
	buf[5] = c.TotalChunks
	copy(buf[6:], c.Data)
	return buf
}

func (c *FrameChunk) Decode(bs []byte) error {

	if len(bs) < 6 {
		return errors.New("frame chunk too small")
	}

	frameId := binary.LittleEndian.Uint32(bs[:4])
	seqNum := bs[4]
	totalChunks := bs[5]
	data := bs[6:]

	c.FrameId = frameId
	c.SequenceNumber = seqNum
	c.TotalChunks = totalChunks
	c.Data = data

	return nil
}

//   * * * | * * * | * *
//   0 1 2   3 4 5   6 7
// size=3
// dataLen=8
// totalChunks=2+1=3

func ChunkFrameData(data []byte, size int, id uint32) []FrameChunk {
	dataLen := len(data)
	totalChunks := (dataLen + size - 1) / size
	chunks := make([]FrameChunk, totalChunks)

	for chunkIdx := 0; chunkIdx < totalChunks; chunkIdx++ {
		start := chunkIdx * size
		end := start + size
		if end > dataLen {
			end = dataLen
		}
		c := data[start:end]
		frameChunk := FrameChunk{
			FrameId:        id,
			SequenceNumber: uint8(chunkIdx),
			TotalChunks:    uint8(totalChunks),
			Data:           c,
		}
		chunks[chunkIdx] = frameChunk
	}

	return chunks
}

type FrameChunkCatcher map[uint32][]FrameChunk

func NewFrameCatcher() FrameChunkCatcher {
	return make(FrameChunkCatcher)
}

func (fcc FrameChunkCatcher) Catch(data []byte) Frame {

	var chunk FrameChunk
	err := (&chunk).Decode(data)
	if err != nil {
		fmt.Printf("ERROR: decoding chunk: %s\n", err)
		return nil
	}

	// fmt.Printf("got seq %d for id=%d\n", chunk.SequenceNumber, chunk.FrameId)

	if chunk.TotalChunks == 1 {
		return RunLengthDecode(chunk.Data)
	}

	if chunks, ok := fcc[chunk.FrameId]; ok {
		chunks = append(chunks, chunk)
		if len(chunks) == int(chunk.TotalChunks) {
			sort.Slice(chunks, func(i, j int) bool {
				return chunks[i].SequenceNumber < chunks[j].SequenceNumber
			})
			joinedDate := []byte{}
			for _, c := range chunks {
				joinedDate = append(joinedDate, c.Data...)
			}
			delete(fcc, chunk.FrameId)
			return RunLengthDecode(joinedDate)
		}

		fcc[chunk.FrameId] = chunks
	} else {
		chunks := []FrameChunk{}
		chunks = append(chunks, chunk)
		fcc[chunk.FrameId] = chunks
	}

	return nil
}
