package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
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
		price, _ := json.Marshal(data[2])
		t := time.Now()

		//os.O_RDWRを渡しているので、同時に読み込みも可能
		file, err := os.OpenFile("./log.csv", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
		if err != nil {
			//エラー処理
			fmt.Println(err)
		}
		fmt.Printf("%v\n", strconv.FormatInt(t.Unix(), 10)+","+string(price)+",coincheck")

		if t.Unix() != lastunixtime {
			lastunixtime = t.Unix()
			//fmt.Fprintln(file, strconv.FormatInt(t.Unix(), 10)+","+string(price)+",coincheck") //書き込み
		}

		file.Close()
	}
}
