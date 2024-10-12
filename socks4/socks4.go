package socks4

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
)

type Server struct {
	l net.Listener
}

func NewServer() (*Server, error) {
	l, err := net.Listen("tcp", ":1080")
	if err != nil {
		return nil, err
	}
	return &Server{
		l: l,
	}, nil
}

func (s *Server) Close() error {
	return s.l.Close()
}

func (s *Server) Serve() error {
	for {
		conn, err := s.l.Accept()
		if err != nil {
			return err
		}
		go handle(conn)
	}
}

type Socks4Command int8

const (
	Socks4CommandCONNECT Socks4Command = 1
)

type Socks4Packet struct {
	Version  int8
	Command  Socks4Command
	DestPort int16
	DestIP   [4]uint8
	UserID   []byte
	raw      []byte
}

func (p Socks4Packet) DestAddressPort() string {
	log.Println(p.DestIP)
	return fmt.Sprintf("%d.%d.%d.%d:%d",
		p.DestIP[0], p.DestIP[1], p.DestIP[2], p.DestIP[3],
		p.DestPort,
	)
}

func handle(conn net.Conn) {
	defer func() {
		log.Println("close handle")
		conn.Close()
	}()
	packetBytes := make([]byte, 1024)
	n, err := bufio.NewReader(conn).Read(packetBytes)
	if err != nil {
		log.Println("failed to read", err)
		return
	}

	log.Printf("read %d bytes", n)
	debug(n, packetBytes)

	destPort := int16(packetBytes[2])<<8 + int16(packetBytes[3])
	log.Println(destPort)

	if packetBytes[0] != 4 {
		log.Println("invalid version:", packetBytes[0])
		return
	}
	packet := &Socks4Packet{
		Version:  int8(packetBytes[0]),
		Command:  Socks4Command(packetBytes[1]),
		DestPort: int16(destPort),
		DestIP: [4]uint8{
			uint8(packetBytes[4]),
			uint8(packetBytes[5]),
			uint8(packetBytes[6]),
			uint8(packetBytes[7]),
		},
		UserID: packetBytes[8 : n-1],
		raw:    packetBytes[:n-1],
	}

	log.Println(packet)

	switch packet.Command {
	case Socks4CommandCONNECT:
		if err := commandCONNECT(conn, packet); err != nil {
			log.Println(err)
			return
		}
	default:
		log.Println("invalid command:", packet.Command)
		return
	}

	log.Println("done")
}

func connectRequestGranted(conn net.Conn, packet *Socks4Packet) error {
	n, err := conn.Write(append(
		[]byte{
			0,
			90,
		},
		packet.raw[2:8]...,
	))
	if err != nil {
		return err
	}
	log.Printf("conn: write %d bytes", n)

	log.Println("request granted")
	return nil
}

func pipe(logName string, src, dest net.Conn) error {
	for {
		n, err := io.CopyN(dest, src, 1024)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		log.Printf("pipe %s: write %d bytes", logName, n)
	}
	log.Printf("pipe %s: done", logName)
	return nil
}

func commandCONNECT(conn net.Conn, packet *Socks4Packet) error {
	log.Println("dial", packet.DestAddressPort())
	dialConn, err := net.Dial("tcp", packet.DestAddressPort())
	if err != nil {
		return err
	}
	defer dialConn.Close()

	if err := connectRequestGranted(conn, packet); err != nil {
		return err
	}

	go func() {
		if err := pipe("client -> dest", conn, dialConn); err != nil {
			log.Println(err)
		}
	}()

	if err := pipe("dest -> client", dialConn, conn); err != nil {
		return err
	}
	log.Println("CONNECT: done")
	return nil
}

func debug(n int, packet []byte) {
	for i := 0; i < n; i += 16 {
		msg := ""
		hex := ""
		ascii := ""
		for j := i; j < i+16 && j < n; j++ {
			b := string(packet[j])
			if b == "\r" {
				b = "\\r"
			}
			if b == "\n" {
				b = "\\n"
			}
			if b == "\r\n" {
				b = "\\r\\n"
			}
			if packet[j] < 32 || packet[j] > 126 {
				b = "."
			}
			hex = fmt.Sprintf("%s%02X", hex, packet[j])
			ascii += fmt.Sprint(b)
			if (j+1)%2 == 0 {
				hex += " "
			}
		}
		space := 40 - len(hex)
		msg = hex
		for j := 0; j < space; j++ {
			msg += " "
		}
		msg += " " + ascii
		fmt.Println(msg)
	}
}
