package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	batch "github.com/Deeptiman/go-batch"
	log "github.com/sirupsen/logrus"
	"time"
)
type Resources struct {
	Id   int `json:"id"`
	Name string `json:"name"`
	Flag bool `json:"flag"`
}
// 1
type longLatStruct struct {
	Long float64 `json:"longitude"`
	Lat  float64 `json:"latitude"`
}
type client struct{
	Conn *websocket.Conn
	IdClient int
}
var clients = make(map[int]*websocket.Conn)
var broadcast = make(chan *longLatStruct)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
var rFlag, mFlag int



func main() {
	flag.IntVar(&rFlag, "r", 100, "No of resources")
	flag.IntVar(&mFlag, "m", 20, "Maximum items")
	flag.Parse()
	// 2
	router := mux.NewRouter()
	router.HandleFunc("/", rootHandler).Methods("GET")
	router.HandleFunc("/longlat", longLatHandler).Methods("POST")
	router.HandleFunc("/ws", wsHandler)
	//go echo()

	log.Fatal(http.ListenAndServe(":8844", router))
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "home")
}


func writer(coord *longLatStruct) {
	broadcast <- coord
}


func longLatHandler(w http.ResponseWriter, r *http.Request) {
	var coordinates longLatStruct
	if err := json.NewDecoder(r.Body).Decode(&coordinates); err != nil {
		log.Printf("ERROR: %s", err)
		http.Error(w, "Bad request", http.StatusTeapot)
		return
	}
	defer r.Body.Close()
	go writer(&coordinates)
}


func wsHandler(w http.ResponseWriter, r *http.Request) {

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()
	timeoutDuration:= 5 * time.Minute
	ws.SetReadDeadline(time.Now().Add(timeoutDuration))
	for {
		var coordinates longLatStruct
		if err := ws.ReadJSON(&coordinates); err == nil {
			go echo(ws)
		}
	}

	// register client




}

// 3
func echo(client *websocket.Conn) {

		logs := log.New()

		logs.Infoln("Batch Processing Example !")

		b := batch.NewBatch(batch.WithMaxItems(uint64(mFlag)))

		b.StartBatchProcessing()
		go func() {

			// Infinite loop to listen to the Consumer Client Supply Channel that releases
			// the []BatchItems for each iteration.


				for bt := range b.Consumer.Supply.ClientSupplyCh {
					//latlong := fmt.Sprintf("%f %f %d", val.Lat, val.Long,i)
					// send to every client that is currently connected
					var result []interface{};
					for _,data := range bt{
						temp:=data.Item
						result=append(result, temp)
					}
					file, _ := json.MarshalIndent(result, "", " ")

					err := client.WriteMessage(websocket.TextMessage, []byte(file))
					if err != nil {
						log.Printf("Websocket error: %s", err)
						client.Close()

					}

					//i=i+1
					//logs.WithFields(log.Fields{"Batch": bt}).Warn("Client")
					//log.Println("test")
					//log.Println(bt[0].Item)
					//file, _ := json.MarshalIndent(bt, "", " ")
					//
					//_ = ioutil.WriteFile("test"+strconv.Itoa(i)+".json", file, 0644)

				}


		}()

		for i := 1; i <= rFlag; i++ {
			b.Item <- Resources{
				Id:   i,
				Name: fmt.Sprintf("%s%d", "R-", i),
				Flag: false,
			}
		}
		b.Close()

		//latlong := fmt.Sprintf("%f %f %d", val.Lat, val.Long,i)
		//// send to every client that is currently connected
		//for client := range clients {
		//	err := client.WriteMessage(websocket.TextMessage, []byte(latlong))
		//	if err != nil {
		//		log.Printf("Websocket error: %s", err)
		//		client.Close()
		//		delete(clients, client)
		//	}
		//}


}
//func getResourceFromItem(data interface{}) Resources {
//	m := data.(Resources)
//	return resource
//}