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
	"regexp"
	"runtime"
	"strings"
	"sync"

	"github.com/AsynkronIT/protoactor-go/actor"
)

func getBinPath(name string) (string, error) {
	postFix := ""
	if runtime.GOOS == `windows` {
		postFix = ".exe"
	}
	return filepath.Abs(path.Join(path.Dir(os.Args[0]), "bin", name+postFix))
}

func getScreenDeviceIndex() (string, error) {

	binPath, err := getBinPath("ffmpeg")
	panicOnErr(err)

	args := `-f avfoundation -list_devices true -i ""`
	cmd := exec.Command(binPath, strings.Fields(args)...)
	data, err := cmd.Output()

	if ee, ok := err.(*exec.ExitError); ok {
		output := string(ee.Stderr)
		res := regexp.MustCompile(`\[(\d)+\] Capture screen`).FindStringSubmatch(output)
		if len(res) > 0 {
			return res[1], nil
		}
	}

	fmt.Println(string(data))
	return "", fmt.Errorf("error get capture device index")
}

func getFfmpegArg(res string) string {
	wh := strings.Split(res, "x")

	winArgs := fmt.Sprintf(`-fflags nobuffer -avioflags direct
-f gdigrab -framerate 30 -show_region 1 -draw_mouse 1
-i desktop
-c:v libx264 -vf scale=%s:-1 -crf 25 -preset ultrafast -tune zerolatency -profile:v baseline -pix_fmt yuv420p
-f rtsp -b:v 10000000
rtsp://localhost:8554/broadcaster/watcher/upstream`, wh[0])

	if runtime.GOOS == `windows` {
		return winArgs
	}

	index, err := getScreenDeviceIndex()
	panicOnErr(err)

	macArgs := fmt.Sprintf(`-fflags nobuffer -avioflags direct
-f avfoundation -r 30 -capture_cursor 1 -capture_mouse_clicks 1
-i %s
-c:v h264_videotoolbox -preset ultrafast -tune zerolatency -profile:v baseline  -allow_sw 1
-f rtsp -b:v 10000000
rtsp://localhost:8554/broadcaster/watcher/upstream`, index)

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
		execPath, err := getBinPath("ffmpeg")
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
				if err != nil {
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
