package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"

	"github.com/gorilla/websocket"
)

func main() {
	url := url.URL{
		Scheme: "ws",
		Host:   "localhost:8080",
		Path:   "/ws",
	}
	// connect to server
	conn, resp, err := websocket.DefaultDialer.Dial(url.String(), nil)
	if err != nil {
		if resp != nil {
			body, _ := io.ReadAll(resp.Body)
			fmt.Printf("Status: %d, Body: %s\n", resp.StatusCode, body)
		}
		log.Println("Failure to connect to server")
		return
	}

	defer conn.Close()

	// Loop Continuously
	// read message from websocket
	// print message to stdout
	// read from stdin
	// send response back via WebSocket
	// Loop until connection closes

	done := make(chan struct{})

	go func() {
		defer close(done) // when this function exits, close the channel
		for {
			_, p, err := conn.ReadMessage()
			if err != nil {
				log.Println("Connection closed by server")
				return
			}
			message := string(p)
			fmt.Printf("Server: %s\n", message)
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		err := conn.WriteMessage(websocket.TextMessage, []byte(line))
		if err != nil {
			fmt.Println("Error writing message: ", err)
			break
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error: reading from stdin: ", err)
	}

	<-done // blocks until the channel is closed
	fmt.Println("Game ended!")

}
