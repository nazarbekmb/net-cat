package server

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func checkValideness(conn net.Conn) (string, error) {
	name, _ := bufio.NewReader(conn).ReadString('\n')
	name = strings.TrimSpace(name)
	for _, user := range clients {
		if name == user.name {
			return "", fmt.Errorf("this name is already in use, try another")
		}
	}
	if len(name) == 0 || len(name) > 20 {
		return "", fmt.Errorf("the length of the name must be at least one character and less than 20 characters")
	}
	return name, nil
}

func printGreeting(conn net.Conn) {
	greeting, err := os.ReadFile("greeting.txt")
	if err != nil {
		panic(err)
	}

	fmt.Fprintln(conn, "Welcome to TCP-Chat!\n"+string(greeting))
}

func checkMsg(msg string) bool {
	for _, v := range msg {
		if v >= '!' && v <= '~' {
			return true
		}
	}
	return false
}

func checkLen(msg string) bool {
	if len(msg) > 0 && len(msg) <= 200 {
		return true
	}
	return false
}
