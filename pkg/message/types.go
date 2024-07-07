package message

type MessageType uint8

const (
	Info    MessageType = iota
	Frame   MessageType = 1
	Audio   MessageType = 2
	Error   MessageType = 99
	Unknown MessageType = 255
)

func Parse(data []byte) ([]byte, MessageType) {
	switch data[0] {
	case 0:
		return data[1:], Info
	case 1:
		return data[1:], Frame
	case 2:
		return data[1:], Audio
	case 99:
		return data[1:], Error
	default:
		return data, Unknown
	}

}

func MakeError(e string) []byte {
	return append([]byte{byte(Error)}, []byte(e)...)
}

func MakeInfo(msg string) []byte {
	return append([]byte{byte(Info)}, []byte(msg)...)
}

func MakeFrame(data []byte) []byte {
	return append([]byte{byte(Frame)}, data...)
}

func MakeAudio(data []byte) []byte {
	return append([]byte{byte(Audio)}, data...)
}
