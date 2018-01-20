package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/websocket"
)

var (
	origin = "http://localhost:8787/"
	url    = "wss://ws-api.coincheck.com/"
)

// EchoMsg is Sample Websocket Message
type EchoMsg struct {
	transactionId int
	channel       string // ID
	price         float32
	quantity      float32
	kind          string
}

var lastunixtime int64

func main() {
	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		panic(err)
	}

	go receiveMsg(ws)

	sendMsg(ws, map[string]string{"type": "subscribe", "channel": "btc_jpy-trades"})

	go forever()
	fmt.Scanln()

	defer fmt.Printf("Web Socket Client Sample end.")
}

func forever() {
	for {
		time.Sleep(time.Second)
	}
}

func sendMsg(ws *websocket.Conn, shakeHands map[string]string) {
	websocket.JSON.Send(ws, shakeHands)
}

func receiveMsg(ws *websocket.Conn) {
	var data []interface{}
	for {
		websocket.JSON.Receive(ws, &data)
		priceBit, _ := json.Marshal(data[2])
		priceString := string(priceBit)
		priceString = strings.Replace(priceString, `"`, "", -1)
		priceFloat, _ := strconv.ParseFloat(priceString, 32)
		fmt.Printf("%v\n", priceFloat)
	}
}
