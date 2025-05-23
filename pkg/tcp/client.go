package tcp

import (
	"log"
	"net"
)

type Client struct {
	ID   string
	Conn net.Conn
	Addr string
	Quit chan struct{}
}

func (c *Client) Disconnect() {
	select {
	case <-c.Quit:
		// already closed
	default:
		close(c.Quit)
	}
	log.Printf("Disconnecting client %s\n", c.ID)
	c.Conn.Write([]byte("You have been disconnected.\n"))
	_ = c.Conn.Close()
}
