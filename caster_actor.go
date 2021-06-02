package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
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

	self *actor.PID
	s    state

	localIp     string
	tvip        string
	res         string
	connTimeout int
}

func (c *caster) onLog(msg string) {
	fmt.Println(msg)
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

var (
	errInitRtmp = fmt.Errorf("err init rtmp server")
)

func (c *caster) doOp(context actor.Context, op int) error {
	switch op {
	case 1:
		program, ok := newProgram([]string{}, c.onLog)
		if !ok {
			return errInitRtmp
		}
		c.rtmp = program
		c.pusher = spawnPusher(context, c.res)
		return nil
	case 0:
		if c.pusher != nil {
			context.PoisonFuture(c.pusher).Wait()
			c.pusher = nil
		}

		if c.rtmp != nil {
			c.rtmp.close()
			c.rtmp = nil
		}
		return nil
	default:
		return fmt.Errorf("invalid op:%d", op)
	}
}

func (c *caster) call(cmd int) error {
	type command struct {
		Command int    `json:"command"`
		Data    string `json:"data"`
	}
	cmdStr, err := json.Marshal(&command{
		Command: cmd,
		Data:    c.localIp,
	})
	if err != nil {
		return err
	}

	conn, err := net.DialTimeout("tcp4", c.tvip+":10099", time.Second*time.Duration(c.connTimeout))
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
	c.writeToCli(fmt.Sprintf("state:%s", c.s))
}

func (c *caster) writeToCli(msg string) {
	if c.cli != nil {
		c.cli.Write([]byte(msg + "\n"))
	}
}

func (c *caster) needGainPrivacy() bool {
	return needGainPrivacy()

}

func (c *caster) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		c.self = context.Self()

		c.res = "1920x1080"
		go func() {
			defer func() {
				context.PoisonFuture(context.Self()).Wait()
			}()
			br := bufio.NewReader(c.cli)
			for {
				data, _, err := br.ReadLine()
				if err != nil {
					break
				}
				context.Send(context.Self(), string(data))
			}
		}()
	case *actor.Stopping:
		c.doOp(context, 0)

		if c.cli != nil {
			c.cli.Close()
			c.cli = nil
		}
	case *actor.Stopped:
		close(c.closeCh)
		c.closeCh = nil
	case *actor.Restarting:
		context.Poison(context.Self())
	case string:
		c.onLog(fmt.Sprintf("recv cmd: %s", msg))
		args := strings.Fields(msg)
		if len(args) == 0 {
			return
		}

		c.writeToCli("cmdEcho:" + args[0])
		switch args[0] {
		case "ver":
			c.writeToCli("ver:" + ffmpegVer)
		case "start":
			// start	192001080	10.224.72.95	xxxx	10.224.215.34	xxxx	20		10
			// start 	res			localip 		dummy	tvip			dummy 	dummy	connectTimeout
			// 0	 	1	 	    2		 		3	   	4				5	  	6		7
			if c.pusher != nil || c.rtmp != nil {
				c.doOp(context, 0)
			}

			// old format XXXX0YYYY to XXXXxYYYY
			if len(args) > 1 {
				c.res = args[1]
				aa := []rune(c.res)
				aa[4] = rune('x')
				c.res = string(aa)
			}

			if len(args) > 2 {
				c.localIp = args[2]
			}

			c.tvip = "10.224.215.34"
			if len(args) > 4 {
				c.tvip = args[4]
			}

			c.connTimeout = 5
			if len(args) > 7 {
				connTimeout, err := strconv.ParseInt(args[7], 10, 32)
				panicOnErr(err)
				c.connTimeout = int(connTimeout)
			}

			c.setState(STARTING)

			c.writeToCli("log:begin cast_start")
			err := c.doOp(context, 1)
			if err == errInitRtmp {
				c.writeToCli("event:rtmp_init_error")
				c.setState(IDLE)
				return
			}

			c.writeToCli("log:end cast_start")

			if c.needGainPrivacy() {
				c.doOp(context, 0)
				c.writeToCli("event:gain_privacy")
				c.cli.Close()
				c.cli = nil
				return
			}

			// wait local stream init
			time.Sleep(time.Millisecond * 1000)

			if c.call(1) != nil {
				c.writeToCli("event:tv_communicate_error")
				c.setState(IDLE)

				c.doOp(context, 0)
			}
		case "stop":
			c.setState(STOPPING)
			c.call(4)

			c.writeToCli("log:begin cast_stop")
			c.doOp(context, 0)
			c.writeToCli("log:end cast_stop")

			c.setState(IDLE)
			if len(args) > 1 {
				c.writeToCli("event:" + args[1])
			}
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
		case "exit":
			c.cli.Close()
			c.cli = nil
		}
	case pusherMsg:
		c.onLog("pusher: " + string(msg))
	case *actor.Terminated:
		if msg.Who.Equal(c.pusher) {
			c.pusher = nil
			if c.s == STARTING {
				panicOnErr(fmt.Errorf("pusher exit unexcepted"))
			}
			c.setState(IDLE)
		}
	case state:
		c.setState(msg)
	}
}
