package main

import (
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func getLocalIp(ipPrefix string) []string {
	ips := []string{}
	ifaces, err := net.Interfaces()
	panicOnErr(err)
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		panicOnErr(err)

		// fmt.Println("IPNnet", i.Name)
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if !ip.IsGlobalUnicast() {
				continue
			}
			ipStr := ip.String()
			if len(ipPrefix) > 0 && strings.Index(ipStr, ipPrefix) < 0 {
				continue
			}
			// fmt.Println("IPNnet", ip)
			ips = append(ips, ipStr)
		}
	}
	return ips
}

func getBinPath(name string) (string, error) {
	postFix := ""
	if runtime.GOOS == `windows` {
		postFix = ".exe"
	}
	binpath, err := filepath.Abs(filepath.Join(filepath.Dir(os.Args[0]), name+postFix))
	panicOnErr(err)
	if fileExists(binpath) {
		return binpath, err
	}

	return filepath.Abs(filepath.Join(filepath.Dir(os.Args[0]), "bin", name+postFix))
}
