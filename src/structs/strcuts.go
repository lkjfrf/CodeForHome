package structs

import (
	"net"
)

type PlayerLoginInfo struct {
	Conn *net.TCPConn
	Data map[string]interface{}
}
