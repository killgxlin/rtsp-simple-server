package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/AsynkronIT/protoactor-go/actor"
)

func getFfmpegArg(res string) string {
	wh := strings.Split(res, "x")

	macArgs := `-fflags nobuffer -avioflags direct
-f avfoundation -r 30 -capture_cursor 1 -capture_mouse_clicks 1
-i ${darwinDeviceIndex}
-c:v h264_videotoolbox -preset ultrafast -tune zerolatency -profile:v baseline  -allow_sw 1
-f rtsp -b:v 10000000
rtsp://localhost:8554/broadcaster/watcher/upstream`

	winArgs := fmt.Sprintf(`-fflags nobuffer -avioflags direct
-f gdigrab -framerate 30 -show_region 1 -draw_mouse 1
-i desktop
-c:v libx264 -vf scale=%s:-1 -crf 25 -preset ultrafast -tune zerolatency -profile:v baseline -pix_fmt yuv420p
-f rtsp -b:v 10000000
rtsp://localhost:8554/broadcaster/watcher/upstream`, wh[0])

	if runtime.GOOS == `windows` {
		return winArgs
	}

	return macArgs
}

type pusherMsg string

type pusher struct {
	cmd *exec.Cmd
	wg  sync.WaitGroup
	res string
}

func (p *pusher) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		execPath, err := filepath.Abs(path.Join(path.Dir(os.Args[0]), "bin", "ffmpeg.exe"))
		panicOnErr(err)

		args := strings.Fields(getFfmpegArg(p.res))
		cmd := exec.Command(execPath, args...)

		log.Println(args)

		serr, err := cmd.StderrPipe()
		panicOnErr(err)
		sout, err := cmd.StdoutPipe()
		panicOnErr(err)

		reader := func(r io.ReadCloser) {
			defer func() {
				context.Poison(context.Self())
			}()
			defer p.wg.Done()
			br := bufio.NewReader(r)
			for {
				data, _, err := br.ReadLine()
				if err == io.EOF {
					break
				}
				panicOnErr(err)
				context.Send(context.Self(), string(data))
			}
		}

		p.wg.Add(2)
		go reader(serr)
		go reader(sout)

		err = cmd.Start()
		panicOnErr(err)
		p.cmd = cmd

	case *actor.Stopping:
		if p.cmd != nil {
			p.cmd.Process.Kill()
		}
	case *actor.Stopped:
		if p.cmd != nil {
			p.cmd.Wait()
			p.cmd = nil
		}
		p.wg.Wait()
		log.Println("pusher finished")
	case *actor.Restarting:
		context.Poison(context.Self())
	case string:
		context.Send(context.Parent(), pusherMsg(msg))
	}
}

func spawnPusher(parent actor.Context, res string) *actor.PID {
	props := actor.PropsFromProducer(func() actor.Actor { return &pusher{res: res} }) //.WithReceiverMiddleware(middleware.Logger)
	return parent.SpawnPrefix(props, "pusher")
}
