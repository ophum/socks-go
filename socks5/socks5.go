package socks5

import (
	"bufio"
	"log"
	"net"
	"slices"

	"github.com/ophum/socks-go/internal/debug"
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

var availableAuthMethods = []uint8{
	//0, // NO AUTHENTICATION REQUIRED
	//1, // GSSAPI
	2, // USERNAME/PASSWORD
}

type AuthPacket struct {
	Version     uint8
	MethodCount uint8
	Methods     []uint8
}

type AuthUsernamePasswordPacket struct {
	Version        uint8
	UsernameLength uint8
	Username       []byte
	PasswordLength uint8
	Password       []byte
}

func handle(conn net.Conn) {
	defer conn.Close()

	packetBytes := make([]byte, 1024)
	n, err := bufio.NewReader(conn).Read(packetBytes)
	if err != nil {
		log.Println("failed to read", err)
		return
	}

	log.Printf("read %d bytes", n)
	debug.Dump(n, packetBytes)

	if packetBytes[0] != 5 {
		log.Println("invalid version", packetBytes[0])
		return
	}
	auth := AuthPacket{
		Version:     packetBytes[0],
		MethodCount: packetBytes[1],
		Methods:     packetBytes[2:],
	}

	selectedAuthMethod := uint8(0xff)
	for i := 0; i < int(auth.MethodCount); i++ {
		if slices.Contains(availableAuthMethods, auth.Methods[i]) {
			selectedAuthMethod = auth.Methods[i]
			break
		}
	}

	n, err = conn.Write([]byte{
		auth.Version,
		selectedAuthMethod,
	})
	if err != nil {
		log.Println("failed to write", err)
		return
	}

	log.Printf("write %d bytes", n)

	n, err = conn.Read(packetBytes)
	if err != nil {
		log.Println("failed to read", err)
		return
	}

	log.Printf("read %d bytes", n)
	debug.Dump(n, packetBytes)

	usernamePasswordPacket := &AuthUsernamePasswordPacket{
		Version:        packetBytes[0],
		UsernameLength: packetBytes[1],
		Username:       packetBytes[2 : 2+packetBytes[1]],
		PasswordLength: packetBytes[2+packetBytes[1]],
		Password:       packetBytes[2+packetBytes[1]+1 : 2+packetBytes[1]+packetBytes[2+packetBytes[1]]+1],
	}

	log.Println("username:", string(usernamePasswordPacket.Username), "password:", string(usernamePasswordPacket.Password))
	status := uint8(0)
	if string(usernamePasswordPacket.Username) != "test" {
		status = 1
	}
	if string(usernamePasswordPacket.Password) != "test-password" {
		status = 1
	}

	n, err = conn.Write([]byte{
		1,
		status,
	})
	if err != nil {
		log.Println("failed to write", err)
		return
	}

	log.Printf("write %d bytes", n)
}
