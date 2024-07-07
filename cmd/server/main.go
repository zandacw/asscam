package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/langlandsbrogram/asscam/pkg/message"
)

type Config struct {
	Port int
	Ip   string
}

// argsParsing parses CLI arguments and returns Config or error
func argsParsing() (Config, error) {
	var config Config

	// Define flags
	flag.StringVar(&config.Ip, "ip", "127.0.0.1", "Ip to listen on (default: 127.0.0.1)")
	flag.IntVar(&config.Port, "port", 6969, "Port to listen on (default: 6969)")

	// Parse command-line arguments
	flag.Parse()

	// Validate required arguments

	if config.Port >= 65535 || config.Port < 1000 {
		flag.PrintDefaults()
		return config, errors.New("Error: 1000 < PORT <= 65535")
	}

	return config, nil
}

func main() {

	args, err := argsParsing()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	addr := net.UDPAddr{
		Port: args.Port,
		IP:   net.ParseIP(args.Ip),
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Printf("Listening on %s ...\n", conn.LocalAddr())

	handleConns(conn)
}

func handleConns(conn *net.UDPConn) {

	bros := NewBros()

	buf := make([]byte, 2048)

	stats := NewStats(1)

	go stats.Check()

	for {

		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			fmt.Println("Error: ", err)
			continue
		}

		if bros.isRoomFull(addr) {
			msg := message.MakeError("full")
			conn.WriteTo(msg, addr)
			continue
		}

		switch data, msg := message.Parse(buf[:n]); msg {
		case message.Info:
			name := string(data)
			bros.add(addr, name)
			msg := message.MakeInfo("ok")
			conn.WriteTo(msg, addr)
		case message.Frame:
			stats.ProcessBytes(n)
			if otherAddr, ok := bros.otherBro(addr); ok {
				msg := message.MakeFrame(data)
				conn.WriteTo(msg, otherAddr)
			} else {
				msg := message.MakeError("empty")
				conn.WriteTo(msg, addr)
			}
		case message.Audio:
			stats.ProcessBytes(n)
			if otherAddr, ok := bros.otherBro(addr); ok {
				msg := message.MakeAudio(data)
				conn.WriteTo(msg, otherAddr)
			} else {
				msg := message.MakeError("empty")
				conn.WriteTo(msg, addr)
			}
		case message.Error:
			bros.remove(addr)
		case message.Unknown:
			fmt.Printf("received unknown message byte: %d; skipping\n", buf[0])
		}

	}
}
