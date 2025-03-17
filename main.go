package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
)

const inputFilePath = "messages.txt"

func main() {
	file, err := os.Open(inputFilePath)
	if err != nil {
		log.Fatalf("Error opening text file: %v", err)
	}
	defer file.Close()

	for {
		buffer := make([]byte, 8)
		n, err := file.Read(buffer)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			fmt.Println("Error reading file:", err.Error())
			break
		}
		str := string(buffer[:n])
		fmt.Printf("read: %s\n", str)
	}

}