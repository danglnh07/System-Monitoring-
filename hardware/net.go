package hardware

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
	"golang.org/x/sys/unix"
)

var SocketType map[uint32]string = map[uint32]string{
	unix.SOCK_STREAM:    "TCP",
	unix.SOCK_DGRAM:     "UDP",
	unix.SOCK_SEQPACKET: "SEQUENCED_PACKET_SOCKETS",
	unix.SOCK_RAW:       "RAW_SOCKETS",
	unix.SOCK_RDM:       "RELIABLE_DATAGRAM",
}

type Address struct {
	IP   string
	Port uint32
}

func (add *Address) String() string {
	return fmt.Sprintf("%s :%d", add.IP, add.Port)
}

type ConnectionInfo struct {
	PID         int32   // Process PID that use the connection
	ProcessName string  // The name of the process that used the connection
	Type        uint32  // Socket type (SOCK_STREAM = TCP, SOCK_DGRAM = UDP)
	LocalAddr   Address // Local address (IP and Port)
	RemoteAddr  Address // Remote address (IP and Port)
	Status      string  // Connection status (e.g., "ESTABLISHED", "LISTEN")
}

func (connInfo *ConnectionInfo) String() string {
	return fmt.Sprintf("PID: %d\nProcess: %s\nConnection type: %s\nLocal address: %s\nRemote Address: %s\nStatus: %s",
		connInfo.PID,
		connInfo.ProcessName,
		SocketType[connInfo.Type],
		connInfo.LocalAddr.String(),
		connInfo.RemoteAddr.String(),
		connInfo.Status)
}

type Connections []ConnectionInfo

func NewConnections() *Connections {
	return &Connections{}
}

func (connections Connections) String() string {
	str := "\t\t---Connections information---\n"
	for _, connInfo := range connections {
		str += fmt.Sprintf("%s\n---\n", connInfo.String())
	}
	return str
}

func (connections *Connections) ToHtml(tmplPath string) (string, error) {
	//Func map
	funcMap := template.FuncMap{
		"DisplayAddress": func(add Address) string {
			return add.String()
		},
	}

	//Get the template
	tmpl, err := template.New("netTmpl.html").Funcs(funcMap).ParseFiles(tmplPath)
	if err != nil {
		return "", err
	}

	//Execute template
	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, connections)
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func (connections *Connections) GetAllConnection() error {
	//Clean the connections first
	*connections = (*connections)[:0]

	/*
	 * The 'kind' parameter filters network connections by protocol
	 * It can have these values:
	 * inet: All IPv4 and IPv6 network connections (both TCP & UDP).
	 * inet4:Only IPv4 connections (both TCP & UDP).
	 * inet6:Only IPv6 connections (both TCP & UDP).
	 * tcp: Both IPv4 and IPv6 TCP connections.
	 * tcp4: Only IPv4 TCP connections.
	 * tcp6: Only IPv6 TCP connections.
	 * udp:	Both IPv4 and IPv6 UDP connections.
	 * udp4: Only IPv4 UDP connections.
	 * udp6: Only IPv6 UDP connections.
	 * unix: Unix domain sockets.
	 */
	conns, err := net.Connections("inet") //Get all connections
	if err != nil {
		return err
	}

	for _, conn := range conns {
		//Filtering network connection (No supported socket type)
		if _, ok := SocketType[conn.Type]; ok {
			//Get the process name that use the connection
			proc, _ := process.NewProcess(conn.Pid)
			var name string
			name, err = proc.Name()
			if err != nil {
				name = "Idle Process" //In Linux, process with PID = 0 cannot get their name, so we assign a fallback value
			}

			connInfo := ConnectionInfo{
				PID:         conn.Pid,
				ProcessName: name,
				Type:        conn.Type,
				LocalAddr:   Address{IP: conn.Laddr.IP, Port: conn.Laddr.Port},
				RemoteAddr:  Address{IP: conn.Raddr.IP, Port: conn.Raddr.Port},
				Status:      conn.Status,
			}
			*connections = append(*connections, connInfo)
		}
	}

	return nil
}
