package client

type optionName string

const (
	nameBasicAuth optionName = "basic"
	nameLoginAuth optionName = "login"
)

type ClientOption struct {
	name  optionName
	value interface{}
}

func WithBasicAuth(user, password string) ClientOption {
	return ClientOption{
		name: nameBasicAuth,
		value: map[string]string{
			"user": user,
			"pass": password,
		},
	}
}

func WithLoginAuth(user, password string) ClientOption {
	return ClientOption{
		name: nameLoginAuth,
		value: map[string]string{
			"user": user,
			"pass": password,
		},
	}
}
