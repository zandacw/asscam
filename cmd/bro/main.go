package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/langlandsbrogram/asscam/pkg/message"
	"github.com/langlandsbrogram/asscam/pkg/video"
)

const (
	udpChunkSize = 256
)

// Config holds configuration parsed from CLI arguments
type Config struct {
	ServerAddr     string
	Name           string
	Width          int
	HideVideo      bool
	FrameChunkSize int
}

// argsParsing parses CLI arguments and returns Config or error
func argsParsing() (Config, error) {
	var config Config

	// Define flags
	flag.StringVar(&config.ServerAddr, "server", "", "Server address (e.g., 198.1.1.8:6969)")
	flag.StringVar(&config.Name, "name", "", "Your name")
	flag.IntVar(&config.Width, "width", 0, "Width of the video")
	flag.IntVar(&config.FrameChunkSize, "chunksize", 256, "Frame chunk size (default: 256)")
	flag.BoolVar(&config.HideVideo, "hide", false, "Flag to indicate whether to show video or not")

	// Parse command-line arguments
	flag.Parse()

	// Validate required arguments
	if config.ServerAddr == "" {
		flag.PrintDefaults()
		return config, fmt.Errorf("server address is required")
	}

	if config.Name == "" {
		flag.PrintDefaults()
		return config, fmt.Errorf("name is required")
	}

	if config.Width == 0 || config.Width >= 255 {
		config.Width = 255
	}

	if config.FrameChunkSize > 1024 {
		config.FrameChunkSize = 1024
	} else if config.FrameChunkSize < 128 {
		config.FrameChunkSize = 128
	}

	return config, nil
}

func main() {

	args, err := argsParsing()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	addr, err := net.ResolveUDPAddr("udp", args.ServerAddr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Dial to the address with UDP
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer removeMe(conn)

	msg := message.MakeInfo(args.Name)
	_, err = conn.Write(msg)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go handleInterupt(cancel)

	frames, err := video.Start(ctx, args.Width)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	datas := dataStream(ctx, conn, args.FrameChunkSize)

	chunkCatcher := video.NewFrameCatcher()

	var frameId uint32
	var oldFrame video.Frame
	for {
		select {
		case frame := <-frames:
			encoded := frame.RunLengthEncode()
			chunks := video.ChunkFrameData(encoded, args.FrameChunkSize, frameId)
			for _, c := range chunks {
				data := c.Encode()
				msg := message.MakeFrame(data)
				conn.Write(msg)
			}

			frameId++

		case data := <-datas:
			switch data, msg := message.Parse(data); msg {
			case message.Info:
			case message.Frame:
				if len(data) < 5 {
					continue
				}

				frame := chunkCatcher.Catch(data)
				if frame != nil {
					frame.Display(oldFrame)
					oldFrame = frame
				}
			case message.Error:
			case message.Unknown:
			}
		case <-ctx.Done():
			return
		}
	}

}

func handleInterupt(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	cancel()

}

func dataStream(ctx context.Context, conn *net.UDPConn, chunkSize int) chan []byte {
	c := make(chan []byte)
	go func() {
		buffer := make([]byte, chunkSize+128)
		for {
			select {
			case <-ctx.Done():
			default:
				n, _, err := conn.ReadFromUDP(buffer)
				if err != nil {
					continue
				}
				data := make([]byte, n)
				copy(data, buffer[:n])
				c <- data
			}
		}
	}()
	return c
}

func sendName(conn *net.UDPConn, name string) error {
	msg := message.MakeInfo(name)
	_, err := conn.Write(msg)
	return err
}

func removeMe(conn *net.UDPConn) {
	msg := []byte{99}
	conn.Write(msg)
}

func printExit(err error) {
	fmt.Println(err)
	os.Exit(1)
}
