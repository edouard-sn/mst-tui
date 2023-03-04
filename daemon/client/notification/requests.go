package notification

import "net"

type NotifyAllRequest struct {
	Message interface{}
}

type NotifyRequest struct {
	Conn    net.Conn
	Message interface{}
}
