package main

import (
	"flag"
	"fmt"
	"net"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/actor/middleware"
)

func panicOnErr(e error) {
	if e != nil {
		panic(e)
	}
}

var ffmpegVer, _ = getFfmpegVer()
var system = actor.NewActorSystem()
var port = flag.Int("p", -1, "help message for flag n")

func main() {
	flag.Parse()

	cli, err := net.Dial("tcp4", fmt.Sprintf("127.0.0.1:%d", *port))
	panicOnErr(err)

	writeLine := func(msg string) {
		cli.Write([]byte(msg + "\n"))
	}

	writeLine("casterkey")
	writeLine("ver:" + ffmpegVer)

	closeCh := make(chan struct{})

	props := actor.PropsFromProducer(func() actor.Actor { return &caster{cli: cli, closeCh: closeCh} })
	props.WithSupervisor(actor.NewOneForOneStrategy(1, time.Microsecond, func(_ interface{}) actor.Directive { return actor.StopDirective }))
	props.WithReceiverMiddleware(middleware.Logger)
	_, err = system.Root.SpawnNamed(props, "caster")
	panicOnErr(err)

	<-closeCh
}
