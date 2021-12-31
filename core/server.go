package core

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/panjf2000/gnet"
	"github.com/rs/zerolog/log"
)

type Zipper struct {
	*gnet.EventServer
	Name        string
	Addr        string
	downstreams []*Client
	conns       sync.Map
}

func NewZipper(name string, addr string, ds ...*Client) *Zipper {
	z := Zipper{
		Name:        name,
		Addr:        addr,
		downstreams: make([]*Client, 0),
		conns:       sync.Map{},
	}
	if len(ds) > 0 {
		z.downstreams = ds
	}

	return &z
}

// OnInitComplete fires when the server is ready for accepting connections.
// The parameter server has information and various utilities.
func (z *Zipper) OnInitComplete(svr gnet.Server) (action gnet.Action) {
	// 连接 DownStreams
	// 此zipper作为客户端连接到downstreams
	go func() {
		for _, ds := range z.downstreams {
			cli, err := gnet.NewClient(
				ds,
			)
			if err != nil {
				panic(err)
			}

			err = cli.Start()
			if err != nil {
				log.Error().Err(err).Msgf("[%s] %s Client.Start err[%T]", "Init", ds.Name, err)
				panic(err)
			}
			network, addr := ParseProtoAddr(ds.Addr)
			if _, err := cli.Dial(network, addr); err != nil {
				// panic(err)
				// TODO: 重连
				log.Error().Err(err).Msgf("[%s] %s connect err[%T]", "Init", ds.Name, err)
			}
			// z.downstreams.Store(ds, conn)
			// defer conn.Close()
		}
	}()

	return
}

// OnShutdown fires when the server is being shut down, it is called right after
// all event-loops and connections are closed.
func (z *Zipper) OnShutdown(svr gnet.Server) {
}

// OnOpened fires when a new connection has been opened.
// The parameter out is the return value which is going to be sent back to the peer.
func (z *Zipper) OnOpened(c gnet.Conn) (out []byte, action gnet.Action) {
	log.Debug().Str("addr", c.RemoteAddr().String()).Msg("OnOpened")
	z.conns.Store(c.RemoteAddr().String(), c)
	return
}

// OnClosed fires when a connection has been closed.
// The parameter err is the last known connection error.
func (z *Zipper) OnClosed(c gnet.Conn, err error) (action gnet.Action) {
	log.Debug().Str("addr", c.RemoteAddr().String()).Msg("OnClosed")
	z.conns.Delete(c.RemoteAddr().String())
	return
}

// PreWrite fires just before a packet is written to the peer socket, this event function is usually where
// you put some code of logging/counting/reporting or any fore operations before writing data to the peer.
func (z *Zipper) PreWrite(c gnet.Conn) {
}

// AfterWrite fires right after a packet is written to the peer socket, this event function is usually where
// you put the []byte's back to your memory pool.
func (z *Zipper) AfterWrite(c gnet.Conn, b []byte) {
}

// React fires when a connection sends the server data.
// Call c.Read() or c.ReadN(n) of Conn c to read incoming data from the peer.
// The parameter out is the return value which is going to be sent back to the peer.
func (z *Zipper) React(packet []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	out = packet
	// z.sendToDownstreams(packet)
	return
}

// Tick fires immediately after the server starts and will fire again
// following the duration specified by the delay return value.
func (z *Zipper) Tick() (delay time.Duration, action gnet.Action) {
	// downstreams
	// z.downstreams.Range(func(key, value interface{}) bool {
	// 	if value == nil {
	// 		if ds, ok := key.(*Client); ok {
	// 			cli, err := gnet.NewClient(
	// 				ds,
	// 			)
	// 			if err != nil {
	// 				panic(err)
	// 			}

	// 			err = cli.Start()
	// 			if err != nil {
	// 				panic(err)
	// 			}
	// 			network, addr := ParseProtoAddr(ds.Addr)
	// 			conn, err := cli.Dial(network, addr)
	// 			if err != nil {
	// 				log.Error().Err(err).Msgf("[%s] %s connect err", "Tick", ds.Name)
	// 			}
	// 			z.conns.Store(ds, conn)
	// 		}
	// 	}
	// 	return true
	// })
	// test 发送数据
	z.conns.Range(func(key, value interface{}) bool {
		addr := key.(string)
		c := value.(gnet.Conn)
		data := fmt.Sprintf("heart[%d] beating to %s", time.Now().UnixMilli(), addr)
		log.Debug().Str("addr", addr).Str("data", data).Msg("Tick send")
		c.AsyncWrite([]byte(data))
		return true
	})
	return 1 * time.Second, gnet.None
}

func (z *Zipper) sendToDownstreams(packet []byte) {
	z.conns.Range(func(key, value interface{}) bool {
		if value != nil {
			if conn, ok := key.(gnet.Conn); ok {
				err := conn.AsyncWrite(packet)
				if err != nil {
					log.Error().Err(err).
						Str("local_addr", conn.LocalAddr().String()).
						Str("remote_addr", conn.RemoteAddr().String()).
						Msg("sendToDownstreams")
				}
			}
		}
		return true
	})
}

func (z *Zipper) ConfigMesh(url string) error {
	if url == "" {
		return nil
	}

	var reader io.Reader
	if strings.Contains(url, "http") {
		log.Info().Msg("Downloading mesh config...")
		// download mesh conf
		res, err := http.Get(url)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		reader = res.Body
	} else {
		f, err := os.Open(url)
		if err != nil {
			return err
		}
		defer f.Close()
		reader = f
	}

	decoder := json.NewDecoder(reader)
	var configs []MeshZipper
	err := decoder.Decode(&configs)
	if err != nil {
		return err
	}

	if len(configs) == 0 {
		return nil
	}

	for _, downstream := range configs {
		if downstream.Name == z.Name {
			continue
		}
		addr := fmt.Sprintf("%s:%d", downstream.Host, downstream.Port)
		log.Debug().Str("name", downstream.Name).Str("addr", addr).Msg("add downstreams")
		z.downstreams = append(z.downstreams, NewClient(downstream.Name, addr))
	}

	return nil
}

func ParseProtoAddr(addr string) (network, address string) {
	network = "tcp"
	address = strings.ToLower(addr)
	if strings.Contains(address, "://") {
		pair := strings.Split(address, "://")
		network = pair[0]
		address = pair[1]
	}
	return
}
