package message

type PingReq struct {
	*Frame
}

func ParsePingReq(f *Frame, p []byte) (*PingReq, error) {
	return &PingReq{
		Frame: f,
	}, nil
}

func NewPingReq() *PingReq {
	return &PingReq{
		Frame: newFrame(PINGREQ),
	}
}

func (p *PingReq) Encode() ([]byte, error) {
	return p.Frame.Encode([]byte{}), nil
}
