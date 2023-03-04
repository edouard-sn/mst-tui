package notification

import (
	"encoding/gob"
	"errors"
	"net"
	"sync"

	"golang.org/x/exp/slog"
)

type Notifier struct {
	conns    map[net.Conn]*gob.Encoder
	chAll    chan NotifyAllRequest
	chSingle chan NotifyRequest
	chStop   chan int
	mu       sync.Mutex
}

func NewNotifier() *Notifier {
	return &Notifier{
		conns:    make(map[net.Conn]*gob.Encoder),
		chAll:    make(chan NotifyAllRequest),
		chSingle: make(chan NotifyRequest),
		mu:       sync.Mutex{},
	}
}

func (nt *Notifier) AddConn(n net.Conn, e *gob.Encoder) {
	nt.mu.Lock()
	defer nt.mu.Unlock()
	nt.conns[n] = e
}

func (nt *Notifier) RemoveConn(n net.Conn) {
	nt.mu.Lock()
	defer nt.mu.Unlock()
	delete(nt.conns, n)
}

func (nt *Notifier) notify(nr *NotifyRequest) error {
	nt.mu.Lock()
	genc, ok := nt.conns[nr.Conn]
	if !ok {
		return errors.New("no such connection")
	}
	nt.mu.Unlock()

	if err := genc.Encode(nr.Message); err != nil {
		return err
	}
	return nil
}

func (nt *Notifier) notifyAll(nr *NotifyAllRequest) {
	nt.mu.Lock()
	defer nt.mu.Unlock()
	for _, genc := range nt.conns {
		if err := genc.Encode(nr.Message); err != nil {
			slog.Error("notifyAll", err)
		}
	}
}

func (nt *Notifier) GetNotifyAllChannel() chan NotifyAllRequest {
	return nt.chAll
}

func (nt *Notifier) GetNotifyChannel() chan NotifyRequest {
	return nt.chSingle
}

func (nt *Notifier) GetStopChannel() chan int {
	return nt.chStop
}

func (nt *Notifier) Done() {
	nt.chStop <- 0
}

// Listen on the 3 channels :
//   - chAll which is used to notify everybody
//   - chSingle which is used to notify a single person
//   - chStop which is used to stop the notifier from listening
func (nt *Notifier) Listen() {
	for {
		select {
		case n := <-nt.chAll:
			go nt.notifyAll(&n)
		case n := <-nt.chSingle:
			go nt.notify(&n)
		case <-nt.chStop:
			return
		}
	}
}
