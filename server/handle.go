package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

var (
	Messages        = make(chan Message)
	welcomeMessages = make(chan User)
	quitMessages    = make(chan User)
)

func StartServer(port string) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}

	defer listener.Close()

	fmt.Printf("Server started, listening on the port %s\n", port)

	go broadcaster()
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}

		mu.Lock()
		if len(clients) == 10 {
			fmt.Fprint(conn, "Too many clients, disconnecting")
			conn.Close()
			continue
		}

		mu.Unlock()
		go handleConn(conn)
	}
}

func broadcaster() {
	file, err := os.Create("logs.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	logger := log.New(file, "", log.LstdFlags)
	// Map to keep track of messages sent to each client
	clientMessages := make(map[net.Conn]map[string]struct{})

	for {
		select {
		case inUser := <-welcomeMessages:
			// Initialize the message set for the new client
			clientMessages[inUser.conn] = make(map[string]struct{})
			logger.SetPrefix("[JOIN] ")
			logger.Printf("%s has joined the chat\n", inUser.name)

			for _, c := range clients {
				if inUser.conn != c.conn {
					fmt.Fprintf(c.conn, "\x1B[1G\x1B[2K%s has joined the chat\n[%s][%s]:", inUser.name, time.Now().Format("2006-01-02 15:04:05"), c.name)
				}
			}

		case msg := <-Messages:

			for i, c := range clients {
				frmt := fmt.Sprintf("[%s][%s]:%s", time.Now().Format("2006-01-02 15:04:05"), msg.sender.name, msg.text)
				if i == 0 {
					logger.SetPrefix("[MESSAGE] ")
					logger.Printf("%s %s", msg.sender.name, msg.text)
				}
				mu.Lock()
				messages = append(messages, frmt)
				mu.Unlock()
				if msg.sender.conn != c.conn {
					if _, ok := clientMessages[c.conn][frmt]; !ok {
						fmt.Fprintf(c.conn, "\x1B[1G\x1B[2K%s[%s][%s]:", frmt, time.Now().Format("2006-01-02 15:04:05"), c.name)
						clientMessages[c.conn][frmt] = struct{}{}
					}
				}
			}

		case outUser := <-quitMessages:
			logger.SetPrefix("[QUIT] ")
			logger.Printf("%s has left the chat\n", outUser.name)

			for _, c := range clients {
				if outUser.conn != c.conn {
					fmt.Fprintf(c.conn, "\x1B[1G\x1B[2K%s has left the chat\n[%s][%s]:", outUser.name, time.Now().Format("2006-01-02 15:04:05"), c.name)
				}
			}

		}
	}
}

func handleConn(conn net.Conn) {
	printGreeting(conn)
	var name string
	for {
		fmt.Fprint(conn, "[ENTER YOUR NAME]:")
		n, err := checkValideness(conn)
		if err == nil {
			name = n
			break
		}
		fmt.Fprintf(conn, "%v\n", err)
	}

	user := User{
		name: name,
		conn: conn,
	}
	mu.Lock()
	clients = append(clients, user)
	mu.Unlock()
	// Create a set to store messages sent to this client
	clientMessages := make(map[string]struct{})

	for _, msg := range messages {
		if _, ok := clientMessages[msg]; !ok {
			fmt.Fprint(conn, msg) // Send the message to the new client
			clientMessages[msg] = struct{}{}
		}
	}

	welcomeMessages <- user

	for {
		fmt.Fprintf(conn, "[%s][%s]:", time.Now().Format("2006-01-02 15:04:05"), user.name)
		userInput, err := bufio.NewReader(user.conn).ReadString('\n')
		if err != nil {
			// fmt.Printf("Error reading from client: %s\n", err)
			break
		}
		if checkMsg(userInput) && checkLen(userInput) {
			Messages <- Message{sender: user, text: userInput}
		}
	}

	defer func() {
		for i, u := range clients {
			if u.conn == user.conn {
				clients = append(clients[:i], clients[i+1:]...)
				break
			}
		}

		quitMessages <- user
		user.conn.Close()
	}()
}
