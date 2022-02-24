package main

import (
	"encoding/hex"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sandbox/sandbox"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func wait(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade err:", err)
		return
	}
	defer c.Close()
	println("wait")
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read err:", err)
			break
		}
		generatorName := GeneratorFilename()
		f, err := os.OpenFile(
			filepath.Join(getCurrentAbPath(), "workspace", generatorName+".cpp"),
			os.O_RDWR|os.O_CREATE,
			0666,
		)
		f.Write(message)
		f.Close()

		p := sandbox.Parcel{
			Filename: generatorName,
			Response: make(chan *sandbox.DispatchResult),
		}
		dispatch <- p
		res := <-p.Response
		bytes, _ := json.Marshal(res)

		err = c.WriteMessage(mt, bytes)
		if err != nil {
			log.Println("write err:", err)
			break
		}
	}
}

func RunWebsocket(port int) {
	http.HandleFunc("/sandbox", wait)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), nil))
}

func GeneratorFilename() string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, 32)
	rand.Read(b)
	rand_str := hex.EncodeToString(b)
	return rand_str
}
