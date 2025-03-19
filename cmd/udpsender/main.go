package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

const serverAddr = ":42069"

func main() {
	udpaddr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		log.Fatalf("couldn't resolve UDP address: %s", err.Error())
	}

	conn, err := net.DialUDP("udp", nil, udpaddr) // Connects to address
	if err != nil {
		log.Fatalf("couldn't establish connection: %s", err.Error())
	}
	defer conn.Close()

	fmt.Printf("Sending to %s. Type your message and press Enter to send. Press Ctrl+C to exit.\n", serverAddr)

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println(">")
		message, err := reader.ReadString('\n') // Reads until a newline character
		if err != nil {
			fmt.Printf("Error reading input: %s", err.Error())
			os.Exit(1)
		}

		numBytes, err := conn.Write([]byte(message))
		if err != nil {
			fmt.Printf("error: %s, number of bytes written: %d", err.Error(), numBytes)
			os.Exit(1)
		}

		fmt.Printf("Message sent: %s", message)
	}
}