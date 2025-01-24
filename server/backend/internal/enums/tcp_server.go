package enums

type ServerStatus byte

const (
	ACK ServerStatus = iota + 1
	FAIL
)

func (s ServerStatus) String() string {
	return [...]string{"ACK\n", "FAIL\n"}[s-1]
}

type Command byte

const (
	LOGIN Command = iota + 1
	HANDSHAKE
)

func (c Command) String() string {
	return [...]string{"LOGIN", "HANDSHAKE"}[c-1]
}
