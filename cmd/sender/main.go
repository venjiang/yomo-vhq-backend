package main

import (
	"flag"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"yomo.run/vhq/app"
)

var addr = flag.String("addr", "tcp://localhost:9000", "zipper address")

func init() {
	flag.Parse()
}

func main() {
	ws := ""
	conn, err := app.Run(*addr, ws)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Info().Msgf("connect to %s", *addr)

	msg := []byte("hello new vhq")
	err = conn.AsyncWrite(msg)
	if err != nil {
		panic(err)
	}

	select {}
}
