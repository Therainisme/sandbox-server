package main

import (
	"encoding/hex"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sandbox-server/sandbox"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type RunRequest struct {
	Code  string `json:"code"`
	Stdin string `json:"stdin"`
}

func wait(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade err:", err)
		return
	}
	defer c.Close()

	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read err:", err)
			break
		}

		var req RunRequest
		json.Unmarshal(message, &req)

		generatorName := GeneratorFilename()
		tempFilepath := filepath.Join(getCurrentAbPath(), "workspace", generatorName)
		println(tempFilepath)
		f, _ := os.OpenFile(
			tempFilepath+".cpp",
			os.O_RDWR|os.O_CREATE,
			0666,
		)
		f.Write([]byte(req.Code))
		f.Close()

		task := sandbox.Task{
			Filename: generatorName,
			Stdin:    req.Stdin,
			Result:   make(chan *sandbox.TaskResult),
		}
		dispatch <- task

		// wait for sandbox result
		res := <-task.Result
		bytes, _ := json.Marshal(res)
		// os.Remove(tempFilepath)
		// os.Remove(tempFilepath + ".cpp")

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
	b := make([]byte, 8)
	rand.Read(b)
	rand_str := hex.EncodeToString(b)
	return rand_str
}
