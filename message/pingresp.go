package message

type PingResp struct {
	*Frame
}

func ParsePingResp(f *Frame, p []byte) (*PingResp, error) {
	return &PingResp{
		Frame: f,
	}, nil
}

func NewPingResp() *PingResp {
	return &PingResp{
		Frame: newFrame(PINGRESP),
	}
}

func (p *PingResp) Encode() ([]byte, error) {
	return p.Frame.Encode([]byte{}), nil
}
