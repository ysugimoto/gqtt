package message

type Disconnect struct {
	*Frame
}

func ParseDisconnect(f *Frame, p []byte) (*Disconnect, error) {
	return &Disconnect{
		Frame: f,
	}, nil
}

func NewDisconnect(f *Frame) *Disconnect {
	return &Disconnect{
		Frame: f,
	}
}

func (d *Disconnect) Encode() ([]byte, error) {
	return d.Frame.Encode([]byte{}), nil
}
