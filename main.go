package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	ws "github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	. "github.com/logrusorgru/aurora"
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

	var name string
	if len(os.Args) < 2 {
		name = "JOIN #dyggaming\r\n"
	} else {
		name = fmt.Sprintf("JOIN #%s \r\n", os.Args[1])
	}

	authCode := os.Getenv("TWITCH_TOKEN")
	env.conn.WriteMessage(ws.TextMessage, []byte("PASS "+authCode+"\r\n"))
	env.conn.WriteMessage(ws.TextMessage, []byte("NICK dyggaming\r\n"))
	env.conn.WriteMessage(ws.TextMessage, []byte(name))
	// env.conn.WriteMessage(ws.TextMessage, []byte("CAP REQ :twitch.tv/tags\r\n"))
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

		msgData, err := parse(string(msg))
		if err != nil {
			// fmt.Printf("Error")
			if err.Error() == "PING" {
				e.conn.WriteMessage(ws.TextMessage, []byte("PONG :tmi.twitch.tv"))
				continue
			}

			if err.Error() == "INIT" {
				continue
			}
			fmt.Println(err)
			continue
		}
		fmt.Printf("%s: %s", Bold(Green(msgData.user)), msgData.msg)
		// fmt.Printf("%s", msg)
	}

}

type message struct {
	user string
	msg  string
}

// 0 : user info
// 1 : msg type (PRIVMSG)
// 2 : Channel name
// 3... : message
func parse(msg string) (*message, error) {

	msgComponents := strings.Split(msg, " ")

	if len(msgComponents) < 3 {
		if msgComponents[0] == "PING" {
			return nil, errors.New("PING")
		}
		return nil, errors.New("ERROR SLICE LESS THAN 3")
	}

	if msgComponents[0] == ":tmi.twitch.tv" || msgComponents[0] == ":dyggaming.tmi.twitch.tv" {
		return nil, errors.New("INIT")
	}

	if msgComponents[0] == ":dyggaming!dyggaming@dyggaming.tmi.twitch.tv" && msgComponents[1] == "JOIN" {
		return nil, errors.New("INIT")
	}
	userMsg := strings.Join(msgComponents[3:], " ")

	username, _, ok := strings.Cut(msgComponents[0], "!")
	if !ok {
		return nil, errors.New("ERROR with cut string")
	}

	username = strings.Trim(username, ":")
	userMsg = strings.Trim(userMsg, ":")
	return &message{user: username, msg: userMsg}, nil

}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
	return
}
