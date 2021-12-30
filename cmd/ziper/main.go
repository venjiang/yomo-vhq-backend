package main

import (
	"os"
	"time"

	"github.com/panjf2000/gnet"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"yomo.run/vhq/core"
)

func main() {

	addr := "tcp://localhost:9000"
	name := "vhq-zipper"

	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Info().Msgf("%s starting on %s ...", name, addr)

	z := core.NewZipper(name, addr)
	err := gnet.Serve(z, addr, gnet.WithMulticore(true), gnet.WithTCPKeepAlive(time.Minute*5))
	if err != nil {
		panic(err)
	}
}
