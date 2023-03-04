package types

type Packet struct {
	CommandID [16]byte
	Payload   any
}
