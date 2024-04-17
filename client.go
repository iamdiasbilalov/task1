package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

const (
	SERVER_PORT = ":3335"
	SERVER_TYPE = "tcp"
)

var wg sync.WaitGroup

// Read messages from the server and print them out
func readMessages(conn net.Conn) {
	defer wg.Done()
	reader := bufio.NewReader(conn)

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from server:", err)
			break
		}
		fmt.Print(message)
	}
}

// Write messages to the server
func writeMessages(conn net.Conn) {
	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(conn)

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from stdin:", err)
			continue
		}

		_, err = writer.WriteString(message)
		if err != nil {
			fmt.Println("Error writing to server:", err)
			continue
		}
		writer.Flush()

		if strings.TrimSpace(message) == "/exit" {
			fmt.Println("Exiting chat...")
			break
		}
	}
}

func main() {
	conn, err := net.Dial(SERVER_TYPE, SERVER_PORT)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	wg.Add(2)

	go func() {
		defer wg.Done()
		readMessages(conn)
	}()

	go func() {
		defer wg.Done()
		writeMessages(conn)
	}()

	wg.Wait()
}
