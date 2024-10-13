package main

import (
	"io"
	"log"
	"net/http"
	"os"

	"golang.org/x/net/proxy"
)

func main() {
	p, err := proxy.SOCKS5("tcp", "localhost:1080", &proxy.Auth{
		User:     "test",
		Password: "test-password",
	}, proxy.Direct)
	if err != nil {
		panic(err)
	}

	client := http.DefaultClient
	client.Transport = &http.Transport{
		Dial: p.Dial,
	}

	resp, err := client.Get(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	log.Println(string(b))
}
