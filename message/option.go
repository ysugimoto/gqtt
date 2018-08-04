package message

type optionName int

const (
	optionNameRetain optionName = iota
	optionNameQoS
)

type option struct {
	name  optionName
	value interface{}
}

func WithRetain() option {
	return option{
		name:  optionNameRetain,
		value: true,
	}
}

func WithQoS(qos uint8) option {
	return option{
		name:  optionNameQoS,
		value: qos,
	}
}
