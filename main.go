package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
)

type state string

const (
	IDLE         state = "IDLE"
	STARTING     state = "STARTING"
	STARTED      state = "STARTED"
	PAUSED       state = "PAUSED"
	STOPPING     state = "STOPPING"
	RECONNECTING state = "RECONNECTING"
)

type caster struct {
	cli     net.Conn
	pusher  *actor.PID
	closeCh chan struct{}
	rtmp    *program

	tvip string
	res  string
	self *actor.PID
	s    state
}

func (c *caster) onLog(msg string) {
	label, _ := parseLog(msg)
	if len(label) == 0 {
		return
	}
	switch label {
	case SessionOpened:
	case SessionPublishing:
	case SessionClosed:
	case ReceiverOpened:
	case ReceiverReading:
		system.Root.Send(c.self, STARTED)
	case ReceiverClosed:
	case ReceiverNoSender:
		system.Root.Send(c.self, RECONNECTING)
	}
}

func (c *caster) call(cmd int) error {
	type command struct {
		Command int    `json:"command"`
		Data    string `json:"data"`
	}
	cmdStr, err := json.Marshal(&command{
		Command: cmd,
		Data:    "10.224.193.13",
	})
	if err != nil {
		return err
	}

	conn, err := net.Dial("tcp4", c.tvip+":10099")
	if err != nil {
		return err
	}
	defer conn.Close()
	conn.SetWriteDeadline(time.Now().Add(time.Second * 3))
	toSend := string(cmdStr) + "\n"
	n, err := conn.Write([]byte(toSend))
	if err != nil {
		return err
	}

	if n != len(toSend) {
		return fmt.Errorf("not write all")
	}

	return nil
}

func (c *caster) setState(s state) {
	c.s = s
	fmt.Println("-----------", c.s)
}

func (c *caster) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		program, ok := newProgram([]string{}, c.onLog)
		if !ok {
			panicOnErr(fmt.Errorf("rtmp server not inited"))
		}
		c.self = context.Self()
		c.rtmp = program

		c.res = "1920x1080"
		go func() {
			defer func() {
				context.PoisonFuture(context.Self()).Wait()
			}()
			br := bufio.NewReader(c.cli)
			for {
				data, _, err := br.ReadLine()
				if err == io.EOF {
					break
				}
				panicOnErr(err)
				context.Send(context.Self(), string(data))
			}
		}()
	case *actor.Stopping:
		c.cli.Close()
		c.rtmp.close()
	case *actor.Stopped:
		close(c.closeCh)
		c.closeCh = nil
	case *actor.Restarting:
		context.Poison(context.Self())
	case string:
		args := strings.Fields(msg)
		switch args[0] {
		case "start":
			if c.pusher != nil {
				log.Println("not idle")
				return
			}
			c.tvip = "10.224.215.34"
			if len(args) > 1 {
				c.tvip = args[1]
			}

			c.setState(STARTING)

			c.pusher = spawnPusher(context, c.res)
			time.Sleep(time.Millisecond * 1000)
			c.call(1)
		case "stop":
			if c.pusher == nil {
				log.Println("not started")
				return
			}
			c.setState(STOPPING)
			context.PoisonFuture(c.pusher).Wait()
			c.pusher = nil
			c.call(4)
			c.setState(IDLE)
		case "pause":
			c.setState(PAUSED)
			c.call(2)
		case "resume":
			c.setState(STARTED)
			c.call(3)
		case "setres":
			c.res = "1920x1080"
			if len(args) > 1 {
				c.res = args[1]
			}

			if c.pusher != nil {
				context.PoisonFuture(c.pusher).Wait()
				c.pusher = spawnPusher(context, c.res)
				time.Sleep(time.Millisecond * 500)
				c.call(5)
			}
		}
	case pusherMsg:
		// log.Println("======", msg)
		c.onLog(string(msg))
	case *actor.Terminated:
		if msg.Who.Equal(c.pusher) {
			c.pusher = nil
			c.setState(IDLE)
		}
	case state:
		c.setState(msg)
	}
}

func panicOnErr(e error) {
	if e != nil {
		panic(e)
	}
}

var system = actor.NewActorSystem()

func main() {

	cli, err := net.Dial("tcp4", "127.0.0.1:9999")
	panicOnErr(err)

	writeLine := func(msg string) {
		cli.Write([]byte(msg + "\n"))
	}

	writeLine("casterkey")
	writeLine("ver:111")

	closeCh := make(chan struct{})

	props := actor.PropsFromProducer(func() actor.Actor { return &caster{cli: cli, closeCh: closeCh} }) //.WithReceiverMiddleware(middleware.Logger)
	_, err = system.Root.SpawnNamed(props, "caster")
	panicOnErr(err)

	<-closeCh
}
