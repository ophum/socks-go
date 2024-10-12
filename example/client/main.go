package main

import (
	"io"
	"log"
	"net/http"
	"os"

	"h12.io/socks"
)

func main() {
	p := socks.Dial("socks4://127.0.0.1:1080")

	client := http.DefaultClient
	client.Transport = &http.Transport{
		Dial: p,
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
