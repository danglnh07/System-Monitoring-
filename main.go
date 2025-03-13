package main

import "sys/server"

func main() {
	//Create server and start
	server := server.NewServer()
	server.Start()
}
