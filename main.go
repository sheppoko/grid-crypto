package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gocarina/gocsv"
	"golang.org/x/net/websocket"
)

const (
	origin = "http://localhost:8787/"
	url    = "wss://ws-api.coincheck.com/"
)

//シミュレーション用パラメータ
var (
	GridRange          = 0.05      //どれくらい下落したら買い下がるか
	TakeProfitRange    = 0.05      //１ポジションあたり何%利益がでたら利益確定するか
	MaxPositionNum     = 10.0      //最大ポジション数
	InitialInvestiment = 1000000.0 //初期投資額
)

//トレード履歴
type TradeHistory struct {
	tradeDateTime time.Time //注文時間
	tradeSize     float64   //注文数
	orderType     int       //買0,売1
	profit        float64   //この注文による利益（売りの場合のみ）
}

//保有ポジション
type Position struct {
	dateTime time.Time //取得時間
	price    float64   //ポジションを取った際の値段
	size     float64   //量
}

//マーケット
type Market struct {
	price      float64   //価格
	lastUpdate time.Time //最終更新
}

//財布
type Wallet struct {
	btc float64
	jpy float64
}

var histories []TradeHistory //取引履歴
var positions []Position     //現在のポジション配列
var market Market            //マーケット情報
var wallet Wallet            //自分の財布

//positions配列から最低価格のポジションの値段を返却します
//ポジションがない場合はfalse,0を返却します
//ポジションがある場合はtrue,最低ポジション取得価格を返却します
func getLowestPostionPrice() (bool, float64) {
	hasPosition := false
	lowestPrice := 0.0
	for _, position := range positions {
		hasPosition = true
		if lowestPrice > position.price || lowestPrice == 0 {
			lowestPrice = position.price
		}
	}
	return hasPosition, lowestPrice
}

//市場価格からポジションを取るべきか判断し必要であればポジションをとります
func buyIfNeed() bool {
	hasPosition, lowestPositionPrice := getLowestPostionPrice()
	shouldBuy := false

	//市場価格が最低ポジションより指定レンジ下げた
	if lowestPositionPrice*(1-GridRange) >= market.price {
		shouldBuy = true
	}

	//ポジションがない
	if !hasPosition {
		shouldBuy = true
	}
	if shouldBuy {
		buy()
	}
	return shouldBuy
}

//ポジションごとに条件を満たした場合に利益確定します
func sellIfNeed() {
	for _, position := range positions {
		if position.price*(1+TakeProfitRange) < market.price {
			sell(position)
		}
	}
}

func buy() {
	position := new(Position)
	position.dateTime = time.Now()
	position.price = market.price
	amountJPYToBuy := wallet.jpy / (MaxPositionNum - float64(len(positions)))

	if wallet.jpy >= amountJPYToBuy {
		position.size = amountJPYToBuy / market.price
		wallet.jpy = wallet.jpy - amountJPYToBuy
		wallet.btc += position.size
		positions = append(positions, *position)
		log.Printf("購入条件成立：BTC%f円で、%fBTC購入します。(使用：%f円)", market.price, position.size, amountJPYToBuy)
		printWallet()
	}
	tradeHistory := new(TradeHistory)
	tradeHistory.orderType = 0
	tradeHistory.profit = 0
	tradeHistory.tradeDateTime = position.dateTime
	tradeHistory.tradeSize = position.size
	histories = append(histories, *tradeHistory)
	outputToCSV()

}

func sell(position Position) {
	newPositions := []Position{}
	for _, p := range positions {
		if p.price != position.price {
			newPositions = append(newPositions, p)
		} else {
			wallet.btc -= p.size
			wallet.jpy += market.price * p.size
			profit := (market.price - position.price) * position.size
			log.Printf("利益確定条件成立：BTCが%f円になったため%fBTCを利益確定します(利益:%f円)", market.price, position.size, profit)
			printWallet()
			tradeHistory := new(TradeHistory)
			tradeHistory.orderType = 1
			tradeHistory.profit = profit
			tradeHistory.tradeDateTime = time.Now()
			tradeHistory.tradeSize = position.size
			histories = append(histories, *tradeHistory)
			outputToCSV()
		}
	}
	positions = newPositions
}

//財布を投資開始状態に戻します
func initWallet() {
	wallet.btc = 0
	wallet.jpy = InitialInvestiment
}

func inputConfig() {

	fmt.Print("買い下がる幅を入力して下さい(ex:0.05)")
	_, err := fmt.Scanf("%f\n", &GridRange)
	if err != nil {
		panic("不正な値")
	}
	fmt.Print("利益確定幅を入力してください(ex:0.05)")
	_, err = fmt.Scanf("%f\n", &TakeProfitRange)
	if err != nil {
		panic("不正な値")
	}
	fmt.Print("最大ポジション数を入力してください(ex:10)")
	_, err = fmt.Scanf("%f\n", &MaxPositionNum)
	if err != nil {
		panic("不正な値")
	}
	fmt.Print("初期投資額を入力して下さい(ex:1000000)")
	_, err = fmt.Scanf("%f\n", &InitialInvestiment)
	if err != nil {
		panic("不正な値")
	}

}

//財布の状況をログに出力します
func printWallet() {
	log.Printf("\t------------------")
	log.Printf("\t保有BTC:%f 保有JPY:%f円", wallet.btc, wallet.jpy)
	log.Printf("\t総資産評価額:%f円", wallet.jpy+wallet.btc*market.price)
	log.Printf("\t------------------\n\n")
}

//設定内容を出力します
func printConfig() {
	log.Printf("------------------")
	log.Printf("購入下落幅%f", GridRange)
	log.Printf("利益確定率:%f", TakeProfitRange)
	log.Printf("最大ポジション数:%f", MaxPositionNum)
	log.Printf("初期投資額:%f", InitialInvestiment)
	log.Print("でシミュレートします")
	log.Printf("------------------\n\n")

}

func outputToCSV() {
	file, _ := os.OpenFile("sample.csv", os.O_RDWR|os.O_CREATE, os.ModePerm)
	defer file.Close()
	gocsv.MarshalFile(&histories, file)
}

func main() {
	initWallet()
	inputConfig()
	printConfig()

	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		panic(err)
	}
	go receiveMsg(ws)
	sendMsg(ws, map[string]string{"type": "subscribe", "channel": "btc_jpy-trades"})
	go forever()
	fmt.Scanln()
	defer fmt.Printf("Web Socket End")
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
		market.price = priceFloat
		market.lastUpdate = time.Now()
		buyIfNeed()
		sellIfNeed()
	}
}
