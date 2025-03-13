package server

import (
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	sync.Mutex                 //Embedding mutex to avoid race condition (can also use a mutex variable here)
	server     *Server         //The server pointer
	msgs       chan []byte     //Message channel for each client
	conn       *websocket.Conn //The client connection struct
}

func NewClient(conn *websocket.Conn, server *Server) *Client {
	return &Client{
		server: server,
		msgs:   make(chan []byte),
		conn:   conn,
	}
}

func (client *Client) SendMessages() {
	//If the loop is broken from (which means the connection is off for some reason), we want to remove the client
	defer client.server.RemoveClient(client)

	//Continously reading messages from Broadcast channel and write the data to client
	for {
		select {
		//If we receive the message from broadcast channel, then we send the message to client
		case msg, ok := <-Broadcast:
			//If fail to read message from broadcast (channel unexpectedly closing,...), send error message to clients
			if !ok {
				fmt.Println("Failed to read messages from broadcast channel")
				err := client.conn.WriteMessage(websocket.CloseMessage, nil)
				if err != nil {
					fmt.Println("Error sending error message to client")
					fmt.Printf("Error: %v\n", err)
				}
				return
			}

			//Send message to client
			err := client.conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				fmt.Println("Failed to send message to clients")
				fmt.Printf("Error: %v\n", err)
			}

		//If we receive message from individual client's message channel, then we process the request and only response
		//to the specific client
		case msg, ok := <-client.msgs:
			if !ok {
				fmt.Println("Failed to read messages")
				err := client.conn.WriteMessage(websocket.CloseMessage, nil)
				if err != nil {
					fmt.Println("Error sending error message to client")
					fmt.Printf("Error: %v\n", err)
				}
				return
			}

			//Send message to client
			err := client.conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				fmt.Println("Failed to write message to clients")
			}
		}
	}
}
