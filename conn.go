package main

import (
	"net"
	"time"
)

type conn struct {
	net.Conn

	IdleTimeout   time.Duration
	MaxReadBuffer int64
	Tag           string
}

func (c *conn) Close() (err error) {
	err = c.Conn.Close()
	return
}

func (c *conn) updateDeadline() {
	idleDeadline := time.Now().Add(c.IdleTimeout)
	c.Conn.SetDeadline(idleDeadline)
}

func (c *conn) sendBadPacket() {
	r, err := GenResponse([]byte("bad"), true)
	if err != nil {
		return
	}
	c.Conn.Write(r)
}

func (c *conn) sendGoodPacket() error {
	r, err := GenResponse([]byte("ok"), true)
	if err != nil {
		return err
	}
	_, err = c.Conn.Write(r)
	if err == nil {
		c.updateDeadline()
	}
	return err
}

func (c *conn) sendMessagePacket(m []byte) error {
	r, err := GenResponse(m, false)
	if err != nil {
		return err
	}
	_, err = c.Conn.Write(r)
	if err == nil {
		c.updateDeadline()
	}
	return err
}
