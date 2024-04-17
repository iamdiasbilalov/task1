package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	CONN_PORT = ":3335"
	CONN_TYPE = "tcp"
)

var (
	clientNames   = make(map[net.Conn]string)
	historyMutex  sync.Mutex
	clientList    sync.Map
	activeClients = make(chan net.Conn)
)

// Log a message to the history.log file
func logMessage(message string) {
	historyMutex.Lock()
	defer historyMutex.Unlock()

	file, err := os.OpenFile("history.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Error opening history log:", err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(message)
	if err != nil {
		log.Println("Error writing to history log:", err)
	}
}

// Broadcast message to everyone
func broadcastMessage(msg string) {
	clientList.Range(func(key, value interface{}) bool {
		conn, ok := key.(net.Conn)
		if !ok {
			return true
		}

		_, err := conn.Write([]byte(msg))
		if err != nil {
			log.Println("Broadcast error:", err)
			conn.Close()
			clientList.Delete(conn)
		}
		return true
	})
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	activeClients <- conn

	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Client disconnected: %v\n", conn.RemoteAddr())
			leaveMessage := fmt.Sprintf("%s Notice: \"%s\" left the chat\n", time.Now().Format("15:04"), clientNames[conn])
			broadcastMessage(leaveMessage)
			logMessage(leaveMessage)
			clientList.Delete(conn)
			delete(clientNames, conn)
			break
		}

		handleMessage(conn, strings.TrimSpace(msg))
	}
}

func handleMessage(conn net.Conn, msg string) {
	if strings.HasPrefix(msg, "/join") {
		nickname := strings.TrimSpace(strings.TrimPrefix(msg, "/join"))
		joinMessage := fmt.Sprintf("%s Notice: \"%s\" joined the chat\n", time.Now().Format("15:04"), nickname)
		clientNames[conn] = nickname

		clientList.Store(conn, nickname)
		broadcastMessage(joinMessage)
		logMessage(joinMessage)
	} else {
		nickname, ok := clientNames[conn]
		if !ok {
			conn.Write([]byte("You need to /join with a nickname first.\n"))
			return
		}

		chatMessage := fmt.Sprintf("%s - %s: %s\n", time.Now().Format("15:04"), nickname, msg)

		broadcastMessage(chatMessage)
		logMessage(chatMessage)
	}
}

func main() {
	ln, err := net.Listen(CONN_TYPE, CONN_PORT)
	if err != nil {
		log.Fatal("Error starting TCP server:", err)
	}
	defer ln.Close()
	go func() {
		for {
			fmt.Println(len(clientNames))
			time.Sleep(time.Second * 5)
		}
	}()
	log.Println("Chat server started on " + CONN_PORT)

	// Handle new active clients
	go func() {
		for conn := range activeClients {
			log.Printf("New client connected: %v\n", conn.RemoteAddr())
		}
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn)
	}
}
