package main

import (
	"flag"
	"os"
	"time"

	"github.com/panjf2000/gnet"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"yomo.run/vhq/core"
)

var (
	addr    = flag.String("addr", "tcp://0.0.0.0:9000", "zipper address")
	meshURL = flag.String("m", "dev.json", "mesh url address")
)

func init() {
	flag.Parse()
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func main() {
	name := "vhq-zipper"

	z := core.NewZipper(name, *addr)
	if *meshURL != "" {
		z.ConfigMesh(*meshURL)
	}
	log.Info().Str("mesh", *meshURL).Msgf("%s starting on %s", name, *addr)
	err := gnet.Serve(z, *addr, gnet.WithMulticore(true), gnet.WithTCPKeepAlive(time.Minute*5), gnet.WithTicker(true))
	if err != nil {
		panic(err)
	}
}
