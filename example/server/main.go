package main

import "github.com/ophum/socks-go/socks4"

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	s, err := socks4.NewServer()
	if err != nil {
		return err
	}
	defer s.Close()

	return s.Serve()
}
