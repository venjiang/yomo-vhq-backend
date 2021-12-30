package core

import (
	"encoding/json"
	"fmt"
	"net/http"
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
				panic(err)
			}
			network, addr := ParseProtoAddr(ds.Addr)
			conn, err := cli.Dial(network, addr)
			if err != nil {
				// panic(err)
				// TODO: 重连
				log.Error().Err(err).Msgf("[%s] %s connect err", "Init", ds.Name)
			}
			z.conns.Store(ds, conn)
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
	return
}

// OnClosed fires when a connection has been closed.
// The parameter err is the last known connection error.
func (z *Zipper) OnClosed(c gnet.Conn, err error) (action gnet.Action) {
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
	return
}

// Tick fires immediately after the server starts and will fire again
// following the duration specified by the delay return value.
func (z *Zipper) Tick() (delay time.Duration, action gnet.Action) {
	z.conns.Range(func(key, value interface{}) bool {
		if value == nil {
			if ds, ok := key.(*Client); ok {
				cli, err := gnet.NewClient(
					ds,
				)
				if err != nil {
					panic(err)
				}

				err = cli.Start()
				if err != nil {
					panic(err)
				}
				network, addr := ParseProtoAddr(ds.Addr)
				conn, err := cli.Dial(network, addr)
				if err != nil {
					log.Error().Err(err).Msgf("[%s] %s connect err", "Tick", ds.Name)
				}
				z.conns.Store(ds, conn)
			}
		}
		return true
	})
	return 3 * time.Second, gnet.None
}

func (z *Zipper) ConfigMesh(url string) error {
	if url == "" {
		return nil
	}

	log.Info().Msg("Downloading mesh config...")
	// download mesh conf
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)
	var configs []MeshZipper
	err = decoder.Decode(&configs)
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
