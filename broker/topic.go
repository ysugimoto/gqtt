package broker

type Topic struct {
	clientId string
	names    []string
}

// client id getter with out side effect
func (t *Topic) ClientId() string {
	return t.clientId
}

// topic names getter with out side effect
func (t *Topic) Names() []string {
	// explicit copy slice
	return append([]string{}, t.names...)
}
