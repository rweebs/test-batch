package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	batch "github.com/Deeptiman/go-batch"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"net/http"
	"sync"
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
	defer func() {
		log.Println("masuk")
		err:= client.WriteMessage(websocket.TextMessage,[]byte("Done"))
		if err !=nil{
			log.Println("error")
		}

	}()
	var wg sync.WaitGroup

		logs := log.New()

		logs.Infoln("Batch Processing Example !")

		b := batch.NewBatch(batch.WithMaxItems(uint64(mFlag)))

		b.StartBatchProcessing()
		go func() {

			// Infinite loop to listen to the Consumer Client Supply Channel that releases
			// the []BatchItems for each iteration.
				for bt := range b.Consumer.Supply.ClientSupplyCh {
					wg.Add(1)
					//latlong := fmt.Sprintf("%f %f %d", val.Lat, val.Long,i)
					// send to every client that is currently connected
					//var result []interface{};
					//for _,data := range bt{
					//	temp:=data.Item
					//	result=append(result, temp)
					//}
					//file, _ := json.MarshalIndent(result, "", " ")
					xlFile := excelize.NewFile()
					sheetName := "transaction-report"
					xlFile.SetSheetName("Sheet1", "transaction-report")
					// set default column name
					xlFile.SetCellValue(sheetName, "A1", "ID")
					xlFile.SetCellValue(sheetName, "B1", "Name")
					xlFile.SetCellValue(sheetName, "C1", "Flag")

					for i, data := range bt {
						xlFile.SetCellValue(sheetName, fmt.Sprintf("A%d", i+2), data.Item.(Resources).Id)
						xlFile.SetCellValue(sheetName, fmt.Sprintf("B%d", i+2), data.Item.(Resources).Name)
						xlFile.SetCellValue(sheetName, fmt.Sprintf("C%d", i+2), data.Item.(Resources).Flag)
					}
					var b bytes.Buffer
					writr := bufio.NewWriter(&b)
					if err := xlFile.Write(writr);err!=nil{
						log.Println("error writing")
					}
					err := client.WriteMessage(websocket.BinaryMessage, b.Bytes())
					if err != nil {
						log.Printf("Websocket error: %s", err)
						client.Close()

					}
					wg.Done()


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
		wg.Wait()

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