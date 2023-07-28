package server

import (
	"net"
	"sync"
)

type Message struct {
	sender User
	text   string
}

var (
	mu       sync.Mutex
	clients  []User
	messages []string
)

type User struct {
	name string
	conn net.Conn
}
