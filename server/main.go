package main

import (
	_ "encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type actionMessage struct {
	X string // "0" || "10" ||  "-10"
	Y string // "0" || "10" ||  "-10"
}

type screenWH struct {
	Width  int64 // "0" || "10" ||  "-10"
	Height int64 // "0" || "10" ||  "-10"
}

type Player struct {
	Id            uuid.UUID `json:"id"`
	ExceptionType string    `json:"exceptionType"`
	Color         [3]int    `json:"color"`
	X int64  `json:"x"`
	Y int64  `json:"y"`
	Show          bool      `json:"show"`
	windowH       int64
	windowW       int64
	Collision     bool `json:"collision"`
	Score         int  `json:"score"`
}

type Exception struct {
	ExceptionType string    `json:"exceptionType"`
	Show          bool      `json:"show"`
	X int64  `json:"x"`
	Y int64  `json:"y"`
}

type ElementsMsg struct {
	Self     *Player    `json:"self,omitempty"`
	Plyer    *Player    `json:"player,omitempty"`
	Excption *Exception `json:"exception,omitempty"`
}

var exceptionsTypes = [11]string{"IOException", "DivideByZeroException", "NullPointerException", "ArithmeticException", "FileNotFoundException", "IndexOutOfBoundsException",
	"InterruptedException", "ClassNotFoundException", "NoSuchFieldException", "NoSuchMethodException", "RuntimeException"}
var clients = make(map[*websocket.Conn]*Player) // connected clients
var broadcastMsg = make(chan ElementsMsg)
var upgrader = websocket.Upgrader{}
var exceptionsMap = struct {
	sync.RWMutex
	items [11]Exception
}{}

func initExceptionsList(){
	exceptionsMap.Lock()
	defer exceptionsMap.Unlock()
	println("init Exceptions: " )
	for i := 0; i < len(exceptionsTypes); i++ {
		exceptionsMap.items[i] = Exception{ExceptionType: exceptionsTypes[i], X: 0, Y: 0, Show: false}
		fmt.Println(" ", exceptionsMap.items[i])
	}
	fmt.Println("init Exceptions done " )
}

func Set()(Exception, bool) {
	var min, max int = 50, 600
	exceptionsMap.Lock()
	defer exceptionsMap.Unlock()
	if i := rand.Intn(len(exceptionsTypes)); !(exceptionsMap.items[i].Show){
		exceptionsMap.items[i].Show = true
		exceptionsMap.items[i].X= int64(rand.Intn(max - min + 1) + min)
		exceptionsMap.items[i].Y= int64(rand.Intn(max - min + 1) + min)
		return exceptionsMap.items[i],true
	}
	return Exception{},false
}

func RemoveRand() (Exception, bool) {
	exceptionsMap.Lock()
	defer exceptionsMap.Unlock()
	if i := rand.Intn(len(exceptionsTypes)); exceptionsMap.items[i].Show{
		exceptionsMap.items[i].Show = false
		return exceptionsMap.items[i], true
	}
	return Exception{},false
}

func doOverlap(playerX int64, playerY int64, exX int64, exY int64) bool{
	if playerX > (exX+130) || exX > (playerX+100){ return false }
	if (playerY+129) < exY || (exY +200) < playerY {return false }
	return true
}

func HandleExceptionCollision(newX int64, newY int64 ,player Player ) (Exception, bool) {
	exceptionsMap.Lock()
	defer exceptionsMap.Unlock()
	for i := 0; i < len(exceptionsMap.items); i++ {
		ex:=exceptionsMap.items[i]
		if ex.Show && doOverlap(newX, newY, ex.X, ex.Y) &&
			ex.ExceptionType == player.ExceptionType {
			exceptionsMap.items[i].Show=false
			return exceptionsMap.items[i], true
		}
	}
	return Exception{}, false
}

func handleNewPlayer(ws *websocket.Conn) {
	player := Player{Id: uuid.New(), X: int64(rand.Intn(600)), Y: int64(rand.Intn(300)), Score: 0, Show: true, ExceptionType: exceptionsTypes[rand.Intn(3)], Color: [3]int{rand.Intn(256), rand.Intn(256), rand.Intn(256)}, Collision: false}
	ws.WriteJSON(ElementsMsg{Self: &player}) //send to client active player as self

	//send all current players to the new player
	for key := range clients {
		if err := ws.WriteJSON(ElementsMsg{Plyer: clients[key]}); err != nil {
			log.Printf("124 error: %v", err)
			ws.Close()
			delete(clients, ws)
		}
	}

	broadcastMsg <- ElementsMsg{Plyer: &player} //broadcast new player to all clients
	clients[ws] = &player

	var msg screenWH
	if err := ws.ReadJSON(&msg); err != nil {
		log.Printf("135 error: %v", err)
		var plyr = clients[ws]
		plyr.Show = false
		broadcastMsg <- ElementsMsg{Plyer: plyr}
		delete(clients, ws)
	}else{
		clients[ws].windowH = msg.Height
		clients[ws].windowW = msg.Width
	}
}

func handlePlayerMovement(ws *websocket.Conn, newX int64, newY int64) {
	x := int64(clients[ws].X) + newX
	y := int64(clients[ws].Y) + newY
	player := *clients[ws]
	if y < 0 || x < 0 || x >= clients[ws].windowW || y >= clients[ws].windowH {
		player.Collision = true
	} else {
		value, ok:= HandleExceptionCollision(x,y,player)
		if ok {
			broadcastMsg <- ElementsMsg{Excption: &value}
			player.Score = player.Score + 1
			clients[ws].Score = player.Score
			player.Collision = false
		}
		player.X , player.Y = x, y
		clients[ws].X, clients[ws].Y = x,y
	}
	broadcastMsg <- ElementsMsg{Plyer: &player}
}

func exceptionsMapHandler() {
	addExTicker := time.NewTicker(2 * time.Second)
	go func() {
		for t := range addExTicker.C {
			_ = t // we don't print the ticker time, so assign this `t` variable to underscore `_` to avoid error
			newEx,ok := Set()
			if ok { broadcastMsg <-  ElementsMsg{Excption: &newEx} }
		}
	}()

	removeExTicker := time.NewTicker(3 * time.Second)
	go func() {
		for t := range removeExTicker.C {
			_ = t // we don't print the ticker time, so assign this `t` variable to underscore `_` to avoid error
			deletedValue,ok :=RemoveRand()
			if ok { broadcastMsg <- ElementsMsg{Excption: &deletedValue} }
		}
	}()
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true } //no cors
	ws, err := upgrader.Upgrade(w, r, nil) // Upgrade initial GET request to a websocket
	if err != nil { log.Fatal(err)}

	handleNewPlayer(ws)
	defer ws.Close() // Make sure we close the connection when the function returns

	for {
		var msg actionMessage
		if err := ws.ReadJSON(&msg); err != nil {
			log.Printf("197 error: %v", err)
			plyrMsg := clients[ws]
			plyrMsg.Show = false
			broadcastMsg <- ElementsMsg{Plyer: plyrMsg}
			delete(clients, ws)
			break
		}
		newX, _ := strconv.ParseInt(msg.X, 10, 64)
		newY, _ := strconv.ParseInt(msg.Y, 10, 64)
		handlePlayerMovement(ws, newX, newY)
	}
}

func broadcastMessages() {
	for {
		msg := <-broadcastMsg
		for client := range clients {
			if err := client.WriteJSON(msg); err != nil {
				log.Printf("215 error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	fmt.Println("ExceptionalMonkeys ... ")

	http.HandleFunc("/ws", handleConnections)
	initExceptionsList()
	go broadcastMessages()
	go exceptionsMapHandler()

	log.Println("http server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
