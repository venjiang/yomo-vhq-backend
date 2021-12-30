package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"yomo.run/vhq/app"
)

func main() {
	addr := "tcp://localhost:9000"
	ws := ""
	conn, err := app.Run(addr, ws)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Info().Msgf("connect to %s", addr)

	msg := []byte("hello new vhq")
	err = conn.AsyncWrite(msg)
	if err != nil {
		panic(err)
	}

	select {}
}
