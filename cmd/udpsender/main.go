package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

const port = ":42069"

func main() {
	addr, err := net.ResolveUDPAddr("udp", port)
	if err != nil {
		log.Fatalf("couldn't resolve UDP address: %s", err.Error())
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatalf("couldn't establish connection: %s", err.Error())
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println(">")
		str, err := reader.ReadString('\n') // Reads until a newline character
		if err != nil {
			fmt.Printf("Error reading input: %s", err.Error())
		}

		numBytes, err := conn.Write([]byte(str))
		if err != nil {
			fmt.Printf("error: %s, number of bytes written: %d", err.Error(), numBytes)
		}
	}
}