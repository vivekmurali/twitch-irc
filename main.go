package main

import (
	"fmt"
	"log"
	"os"
	"sync"

	ws "github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

type environment struct {
	conn *ws.Conn
}

func init() {
	godotenv.Load()
}

func connect() environment {
	c, _, err := ws.DefaultDialer.Dial("ws://irc-ws.chat.twitch.tv:80", nil)
	checkErr(err)
	return environment{c}
}

func initChat(env environment) {
	authCode := os.Getenv("TWITCH_TOKEN")
	env.conn.WriteMessage(ws.TextMessage, []byte("PASS "+authCode+"\r\n"))
	env.conn.WriteMessage(ws.TextMessage, []byte("NICK dyggaming\r\n"))
	env.conn.WriteMessage(ws.TextMessage, []byte("JOIN #dyggaming\r\n"))
	env.conn.WriteMessage(ws.TextMessage, []byte("CAP REQ :twitch.tv/tags\r\n"))
}

func main() {
	e := connect()
	var wg sync.WaitGroup
	go func() {
		wg.Add(1)
		defer wg.Done()
		readMessages(e)
	}()
	initChat(e)
	wg.Wait()
}

func readMessages(e environment) {
	for {
		_, msg, err := e.conn.ReadMessage()
		checkErr(err)
		fmt.Println(string(msg))
		//TODO
		// - [ ] Parse message
		// - [ ] Reply with pong
	}
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
	return
}
