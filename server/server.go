package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sys/hardware"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shirou/gopsutil/process"
)

/*---Variable and type declaration---*/
var (
	// Web socket upgrader
	wsUpgrader = websocket.Upgrader{
		WriteBufferSize: 1024,
		ReadBufferSize:  1024,
		//For simplicity, we will want to return the Check Origin true for all clients connect to the server
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	//Broadcast channel for the server to send message to
	Broadcast = make(chan []byte, 1024)
)

type Server struct {
	sync.Mutex                  //Embedding mutex to avoid race condition
	mux        http.ServeMux    //The server multiplxer
	clients    map[*Client]bool //Map used to keep track of all clients currently connecting to the server
	done       chan struct{}    //Done channel, used for graceful shutdown (not implemented yet)
}

func NewServer() *Server {
	return &Server{
		mux:     *http.NewServeMux(),
		clients: make(map[*Client]bool),
		done:    make(chan struct{}),
	}
}

/*---Handle websocket---*/

func (server *Server) Serve_WebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("Failed to upgrade web socket request\nError: %v\n", err)
		return
	}

	//If a connection upgrade success, add new publisher
	client := server.AddClient(conn)

	//Start independent goroutine
	go client.SendMessages()
}

func (server *Server) AddClient(conn *websocket.Conn) *Client {
	//Lock the server struct to avoid race condition
	server.Lock()
	defer server.Unlock()

	//Create new client and registered it to the server map
	client := NewClient(conn, server)
	server.clients[client] = true

	fmt.Println("New client join the server")

	return client
}

func (server *Server) RemoveClient(client *Client) {
	//Lock the server struct to avoid race condition
	server.Lock()
	defer server.Unlock()

	//Check if the current client has been registered
	if _, hasRegistered := server.clients[client]; hasRegistered {
		//Close the channel message
		close(client.msgs)
		//Close the connection
		client.conn.Close()
		//Remove publisher from the list of publisher
		delete(server.clients, client)

		fmt.Println("Client lost connection")

		return
	}

	//If this client has not been registered, then print out the message
	fmt.Println("Cannot found this client!")
}

func (server *Server) HandleProcessAction(w http.ResponseWriter, r *http.Request) {
	/*
	 * Handler for handling the process action: kill, terminate or send signal
	 * We don't need a web socket connection here, since we fetch the data and send to all clients in a fixed interval
	 */

	//Get the request parameter
	params := r.URL.Query()
	action := params.Get("action")
	pidRaw := params.Get("pid")

	//Sanitize action and parse pid
	action = strings.ToLower(action)
	pid, err := strconv.Atoi(pidRaw)
	if err != nil {
		fmt.Printf("Failed to parse PID\nError: %v\n", err)
		http.Error(w, "Invalid PID", http.StatusBadRequest)
		return
	}

	//Get the process by PID
	proc, err := process.NewProcess(int32(pid))
	if err != nil {
		fmt.Printf("Failed to get process with PID %d to perform action\nError: %v\n", pid, err)
		http.Error(w, "Fail to execute action to process", http.StatusInternalServerError)
		return
	}

	//Check for each action
	switch action {
	case "kill":
		err := proc.Kill()
		if err != nil {
			http.Error(w, "Internal server error: Failed to kill process", http.StatusInternalServerError)
		} else {
			//Send success response message
			w.WriteHeader(http.StatusOK)
			w.Write(fmt.Appendf(nil, "Process with PID %d kill sucessfully", pid))
		}
	case "terminate":
		err := proc.Terminate()
		if err != nil {
			http.Error(w, "Internal server error: Failed to terminate process", http.StatusInternalServerError)
		} else {
			//Send success response message
			w.WriteHeader(http.StatusOK)
			w.Write(fmt.Appendf(nil, "Process with PID %d terminate sucessfully", pid))
		}
	case "send_signal":
		//Get the signal value
		signalRaw := params.Get("signal")
		signal, err := strconv.Atoi(signalRaw)
		if err != nil {
			fmt.Printf("Failed to parse Signal\nError: %v\n", err)
			http.Error(w, "Invalid Signal", http.StatusBadRequest)
			return
		}
		err = proc.SendSignal(syscall.Signal(signal))
		if err != nil {
			http.Error(w, "Internal server error: Failed to send signal to process", http.StatusInternalServerError)
		} else {
			//Send success response message
			w.WriteHeader(http.StatusOK)
			w.Write(fmt.Appendf(nil, "Process with PID %d receive signal sucessfully", pid))
		}
	}
}

/*---Config server---*/
func (server *Server) Start() {
	/*---Serve the static files---*/

	//Serve the index.html file
	fs := http.FileServer(http.Dir("./web"))
	server.mux.Handle("/", fs)

	//Serve the static resources
	fs = http.FileServer(http.Dir("./web/static"))
	server.mux.Handle("/static/", http.StripPrefix("/static", fs))

	//Handler for upgrading from HTTP to Web Socket
	server.mux.HandleFunc("/ws", server.Serve_WebSocket)
	server.mux.HandleFunc("/process", server.HandleProcessAction)

	//Start the goroutine for collecting system data
	go func() {
		//We used ticker (which has a channel as a field) for fetching data internally instead of using time.Sleep
		ticker := time.NewTicker(time.Second) //Set interval is 1 second
		defer ticker.Stop()

		//Create the hardware struct
		hw := hardware.NewHardware()

		//Fetch data continously until we receive some data in done channel (most of the time is server manually shutdown)
		for {
			select {
			case <-ticker.C:
				hw.CollectData()
				html, err := hw.ToHtml(hardware.TMPL)
				if err == nil {
					//If we success to get the data, then send it to broadcast
					Broadcast <- []byte(html)
				} else {
					//If fail, let's just log the error and ignore the current fetching
					fmt.Printf("Failed to collect system data\nError: %v\n", err)
				}
			case <-server.done: //Not implemented yet
				return
			}
		}

	}()

	//Listen and serve
	fmt.Println("Start server at http://localhost:8800 ...")
	err := http.ListenAndServe(":8800", &server.mux)
	if err != nil {
		fmt.Printf("Failed to start server\nError: %v\n", err)
		os.Exit(1)
	}
}
