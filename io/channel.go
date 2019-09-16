package io

import (
	"fmt"
	"syscall"
	"unsafe"
)

type Channel struct {
	Ev     *syscall.EpollEvent
	in     *IOBuffer
	out    *IOBuffer
	proto  Protocol
	bizSrv Service
}

func NewChannel(ev *syscall.EpollEvent) *Channel {
	return &Channel{
		Ev:     ev,
		in:     NewIOBuffer(),
		out:    NewIOBuffer(),
		proto:  NewHTTPProtocol(),
		bizSrv: NewBizzService(),
	}
}
func (c *Channel) OnRead() {
	for {
		wt, _ := c.in.EnsureWrite(512)
		n, err := syscall.Read(int(c.Ev.Fd), c.in.Buffer()[wt:])
		if n <= 0 {
			break
		}
		c.in.Produce(uint64(n))
		if c.in.GetReadSize() >= 512 {
			if err = c.DecodeMessage(); err != nil {
				syscall.Shutdown(int(c.Ev.Fd), syscall.SHUT_RD)
			}
		}
	}

	if err := c.DecodeMessage(); err != nil {
		syscall.Shutdown(int(c.Ev.Fd), syscall.SHUT_RD)
	}
}

func (c *Channel) DecodeMessage() error {
	for c.in.GetReadSize() > 0 {
		msg, err := c.proto.Decode(c.in)
		if err != nil {
			return err
		}
		msg.Udata = (*byte)(unsafe.Pointer(c))
		c.bizSrv.Serve(msg)
	}
	return nil
}

func (c *Channel) OnWrite() {
	for c.out.GetReadSize() > 0 {
		n, err := syscall.Write(int(c.Ev.Fd), c.out.GetReadBuffer())
		if n <= 0 {
			if err != nil {
				fmt.Printf("write %v\n", err)
			}
			c.Reset()
			break
		}
		c.out.Consume(uint64(n))
	}
}

func (c *Channel) Reset() {
	c.in.Reset()
	c.out.Reset()
}

func (c *Channel) EncodeMessage(m Message) {
	c.proto.Encode(m, c.out)

	c.OnWrite()
}
