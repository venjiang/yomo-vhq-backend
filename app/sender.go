package app

import (
	"fmt"

	"github.com/panjf2000/gnet"
	"github.com/rs/zerolog/log"
	"yomo.run/vhq/core"
)

type WSSender struct {
	*core.Client
	wsURL string
}

func NewSender(addr string, wsURL string) *WSSender {
	return &WSSender{
		Client: core.NewClient("sender", addr),
		wsURL:  wsURL,
	}
}

func (c *WSSender) React(packet []byte, svr gnet.Conn) (out []byte, action gnet.Action) {
	fmt.Println("received: ", string(packet))
	return
}

func Run(addr, wsURL string) (gnet.Conn, error) {
	sender := NewSender(addr, wsURL)
	client, err := gnet.NewClient(
		sender,
		gnet.WithTCPNoDelay(gnet.TCPNoDelay),
	)
	if err != nil {
		log.Error().Err(err).Msg("NewClient")
		return nil, err
	}

	err = client.Start()
	if err != nil {
		log.Error().Err(err).Msg("Client.Start")
		return nil, err
	}

	network, addr := core.ParseProtoAddr(sender.Addr)
	conn, err := client.Dial(network, addr)
	if err != nil {
		log.Error().Err(err).Msg("Client.Dial")
		return nil, err
	}
	return conn, nil
}
