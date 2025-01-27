package enums

type ServerStatus byte

const (
	ACK ServerStatus = iota + 1
	FAIL
)

func (s ServerStatus) String() string {
	return [...]string{"ACK", "FAIL"}[s-1]
}

func (s ServerStatus) EnumIndex() int {
	return int(s)
}

type Command byte

const (
	LOGIN Command = iota + 1
	HANDSHAKE
)

func (c Command) String() string {
	return [...]string{"LOGIN", "HANDSHAKE"}[c-1]
}
