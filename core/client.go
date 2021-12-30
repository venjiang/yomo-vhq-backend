package core

import (
	"time"

	"github.com/panjf2000/gnet"
)

type Client struct {
	*gnet.EventServer
	Name string
	Addr string
}

func NewClient(name string, addr string) *Client {
	c := Client{
		Name: name,
		Addr: addr,
	}

	return &c
}

// OnInitComplete fires when the server is ready for accepting connections.
// The parameter server has information and various utilities.
func (c *Client) OnInitComplete(svr gnet.Server) (action gnet.Action) {
	return
}

// OnShutdown fires when the server is being shut down, it is called right after
// all event-loops and connections are closed.
func (c *Client) OnShutdown(svr gnet.Server) {
}

// OnOpened fires when a new connection has been opened.
// The parameter out is the return value which is going to be sent back to the peer.
func (c *Client) OnOpened(svr gnet.Conn) (out []byte, action gnet.Action) {
	return
}

// OnClosed fires when a connection has been closed.
// The parameter err is the last known connection error.
func (c *Client) OnClosed(svr gnet.Conn, err error) (action gnet.Action) {
	return
}

// PreWrite fires just before a packet is written to the peer socket, this event function is usually where
// you put some code of logging/counting/reporting or any fore operations before writing data to the peer.
func (c *Client) PreWrite(svr gnet.Conn) {
}

// AfterWrite fires right after a packet is written to the peer socket, this event function is usually where
// you put the []byte's back to your memory pool.
func (c *Client) AfterWrite(svr gnet.Conn, b []byte) {
}

// React fires when a connection sends the server data.
// Call c.Read() or c.ReadN(n) of Conn c to read incoming data from the peer.
// The parameter out is the return value which is going to be sent back to the peer.
func (c *Client) React(packet []byte, svr gnet.Conn) (out []byte, action gnet.Action) {
	return
}

// Tick fires immediately after the server starts and will fire again
// following the duration specified by the delay return value.
func (c *Client) Tick() (delay time.Duration, action gnet.Action) {
	return
}
