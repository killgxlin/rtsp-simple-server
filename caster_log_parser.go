package main

import (
	"bufio"
	"io"
	"regexp"
	"strings"
)

var logs = `2021/05/28 15:46:19 I [0/0] rtsp-simple-server v0.0.0
2021/05/28 15:46:19 I [0/0] [RTSP] UDP/RTP listener opened on :8000
2021/05/28 15:46:19 I [0/0] [RTSP] UDP/RTCP listener opened on :8001
2021/05/28 15:46:19 I [0/0] [RTSP] TCP listener opened on :8554
2021/05/28 15:46:19 I [0/0] [RTMP] listener opened on :1935
2021/05/28 15:46:19 I [0/0] [HLS] listener opened on :8888
2021/05/28 15:46:26 I [0/0] [RTMP] [conn 10.224.215.34:47264] opened
2021/05/28 15:46:26 I [0/0] [RTMP] [conn 10.224.215.34:47264] ERR: no one is publishing to path 'broadcaster/watcher/upstream'
2021/05/28 15:46:26 I [0/0] [RTMP] [conn 10.224.215.34:47264] closed
2021/05/28 15:46:38 I [0/0] [RTMP] [conn 10.224.215.34:47268] opened
2021/05/28 15:46:38 I [0/0] [RTMP] [conn 10.224.215.34:47268] ERR: no one is publishing to path 'broadcaster/watcher/upstream'
2021/05/28 15:46:38 I [0/0] [RTMP] [conn 10.224.215.34:47268] closed
2021/05/28 15:46:50 I [0/0] [RTMP] [conn 10.224.215.34:47272] opened
2021/05/28 15:46:50 I [0/0] [RTMP] [conn 10.224.215.34:47272] ERR: no one is publishing to path 'broadcaster/watcher/upstream'
2021/05/28 15:46:50 I [0/0] [RTMP] [conn 10.224.215.34:47272] closed
2021/05/28 15:47:03 I [0/0] [RTMP] [conn 10.224.215.34:47276] opened
2021/05/28 15:47:03 I [0/0] [RTMP] [conn 10.224.215.34:47276] ERR: no one is publishing to path 'broadcaster/watcher/upstream'
2021/05/28 15:47:03 I [0/0] [RTMP] [conn 10.224.215.34:47276] closed
2021/05/28 15:47:16 I [0/0] [RTMP] [conn 10.224.215.34:47282] opened
2021/05/28 15:47:16 I [0/0] [RTMP] [conn 10.224.215.34:47282] ERR: no one is publishing to path 'broadcaster/watcher/upstream'
2021/05/28 15:47:16 I [0/0] [RTMP] [conn 10.224.215.34:47282] closed
2021/05/28 15:47:29 I [0/0] [RTSP] [conn [::1]:50336] opened
2021/05/28 15:47:29 I [0/0] [RTSP] [session 1593976251] opened by [::1]:50336
2021/05/28 15:47:29 I [1/0] [RTSP] [session 1593976251] is publishing to path 'broadcaster/watcher/upstream', 1 track with UDP
2021/05/28 15:47:30 I [1/0] [RTMP] [conn 10.224.215.34:47286] opened
2021/05/28 15:47:30 I [1/1] [RTMP] [conn 10.224.215.34:47286] is reading from path 'broadcaster/watcher/upstream'
2021/05/28 15:47:48 I [0/0] [RTMP] [conn 10.224.215.34:47286] closed
2021/05/28 15:47:48 I [0/0] [RTSP] [session 1593976251] closed
2021/05/28 15:47:48 I [0/0] [RTSP] [conn [::1]:50336] ERR: read tcp [::1]:8554->[::1]:50336: wsarecv: An established connection was aborted by the software in your host machine.
2021/05/28 15:47:48 I [0/0] [RTSP] [conn [::1]:50336] closed
2021/05/28 15:48:00 I [0/0] [RTMP] [conn 10.224.215.34:47294] opened
2021/05/28 15:48:00 I [0/0] [RTMP] [conn 10.224.215.34:47294] ERR: no one is publishing to path 'broadcaster/watcher/upstream'
2021/05/28 15:48:00 I [0/0] [RTMP] [conn 10.224.215.34:47294] closed
2021/05/28 15:48:02 I [0/0] [RTMP] [conn 10.224.215.34:47296] opened
2021/05/28 15:48:02 I [0/0] [RTMP] [conn 10.224.215.34:47296] ERR: no one is publishing to path 'broadcaster/watcher/upstream'
2021/05/28 15:48:02 I [0/0] [RTMP] [conn 10.224.215.34:47296] closed
2021/05/28 15:48:03 I [0/0] [RTSP] [conn [::1]:50341] opened
2021/05/28 15:48:03 I [0/0] [RTSP] [session 917001550] opened by [::1]:50341
2021/05/28 15:48:03 I [1/0] [RTSP] [session 917001550] is publishing to path 'broadcaster/watcher/upstream', 1 track with UDP
2021/05/28 15:48:05 I [1/0] [RTMP] [conn 10.224.215.34:47300] opened
2021/05/28 15:48:05 I [1/1] [RTMP] [conn 10.224.215.34:47300] is reading from path 'broadcaster/watcher/upstream'
2021/05/28 15:48:08 I [1/1] [RTMP] [conn 10.224.215.34:47302] opened
2021/05/28 15:48:08 I [1/1] [RTMP] [conn 10.224.215.34:47300] ERR: write tcp 10.224.72.95:1935->10.224.215.34:47300: wsasend: An existing connection was forcibly closed by the remote host.

2021/05/28 15:48:08 I [1/0] [RTMP] [conn 10.224.215.34:47300] closed
2021/05/28 15:48:08 I [1/1] [RTMP] [conn 10.224.215.34:47302] is reading from path 'broadcaster/watcher/upstream'
2021/05/28 15:48:23 I [1/1] [RTMP] [conn 10.224.215.34:47302] ERR: write tcp 10.224.72.95:1935->10.224.215.34:47302: wsasend: An established connection was aborted by the software in your 
host machine.
2021/05/28 15:48:23 I [1/0] [RTMP] [conn 10.224.215.34:47302] closed
2021/05/28 15:48:29 I [1/0] [RTMP] [conn 10.224.215.34:47310] opened
2021/05/28 15:48:29 I [1/1] [RTMP] [conn 10.224.215.34:47310] is reading from path 'broadcaster/watcher/upstream'
2021/05/28 15:48:45 I [1/1] [RTMP] [conn 10.224.215.34:47310] ERR: write tcp 10.224.72.95:1935->10.224.215.34:47310: wsasend: An existing connection was forcibly closed by the remote host.

2021/05/28 15:48:45 I [1/0] [RTMP] [conn 10.224.215.34:47310] closed
2021/05/28 15:49:17 I [0/0] [RTSP] [session 917001550] closed
2021/05/28 15:49:17 I [0/0] [RTSP] [conn [::1]:50341] ERR: read tcp [::1]:8554->[::1]:50341: wsarecv: An established connection was aborted by the software in your host machine.
2021/05/28 15:49:17 I [0/0] [RTSP] [conn [::1]:50341] closed`

type parsedLabel string

type parserItem struct {
	label parsedLabel
	reg   *regexp.Regexp
}

const (
	SessionOpened     parsedLabel = "SessionOpened"
	SessionPublishing parsedLabel = "SessionPublishing"
	SessionClosed     parsedLabel = "SessionClosed"
	ReceiverOpened    parsedLabel = "ReceiverOpened"
	ReceiverReading   parsedLabel = "ReceiverReading"
	ReceiverNoSender  parsedLabel = "ReceiverNoSender"
	ReceiverClosed    parsedLabel = "ReceiverClosed"
)

var parsers = []*parserItem{
	{
		label: SessionOpened,
		reg:   regexp.MustCompile(`\[session (\d+)\] opened by (.+)`),
	},
	{
		label: SessionPublishing,
		reg: regexp.MustCompile(`\[session (\d+)\] is publishing to path '(.+)', 1 track with UDP	`),
	},
	{
		label: SessionClosed,
		reg:   regexp.MustCompile(`\[session (\d+)\] closed`),
	},
	{
		label: ReceiverOpened,
		reg:   regexp.MustCompile(`\[conn (.*):(\d+)\] opened`),
	},
	{
		label: ReceiverReading,
		reg:   regexp.MustCompile(`\[conn (.*):(\d+)\] is reading from path '(.+)'`),
	},
	{
		label: ReceiverNoSender,
		reg:   regexp.MustCompile(`\[conn (.*):(\d+)\] ERR: no one is publishing to path '(.+)'`),
	},
	{
		label: ReceiverClosed,
		reg:   regexp.MustCompile(`\[conn (.*):(\d+)\] closed`),
	},
}

func parseLog(log string) (parsedLabel, []string) {
	for _, parser := range parsers {
		res := parser.reg.FindStringSubmatch(log)
		if len(res) == 0 {
			continue
		}
		return parser.label, res
	}
	return "", nil
}

func testParser() {
	br := bufio.NewReader(strings.NewReader(logs))
	for {
		data, _, err := br.ReadLine()
		if err == io.EOF {
			break
		}
		panicOnErr(err)
		_ = data
		parseLog(string(data))
	}
}
