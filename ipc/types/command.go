package types

type CommandType uint8

type Packet struct {
	CommandID [16]byte
	Payload   any
}
