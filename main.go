package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const inputFilePath = "messages.txt"

func main() {
	file, err := os.Open(inputFilePath)
	if err != nil {
		log.Fatalf("Error opening text file: %v", err)
	}
	defer file.Close()
	
	ch := getLinesChannel(file)
	for value := range ch {
		fmt.Printf("read: %s\n", value)
	}

}

func getLinesChannel(f io.ReadCloser) <-chan string {
	ch := make(chan string)
	go func() {
		defer close(ch)
		currentLineContents := ""

		for {
			buffer := make([]byte, 8)
			n, err := f.Read(buffer)
			if err != nil {
				if errors.Is(err, io.EOF) {
					if currentLineContents != "" {
						ch <- currentLineContents
					}
					break
				}
				fmt.Println("Error reading file:", err.Error())
				break
			}
		str := string(buffer[:n])
		parts := strings.Split(str, "\n")
		for i := 0; i < len(parts)-1; i++ {
			ch <- currentLineContents + parts[i]
			currentLineContents = ""
		}
		currentLineContents += parts[len(parts)-1]
		}
	}()
	return ch
}