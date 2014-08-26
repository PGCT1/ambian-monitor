package main

import (
  "github.com/gorilla/websocket"
  "github.com/go-martini/martini"
  "github.com/martini-contrib/cors"
  "github.com/pgct1/ambian-monitor/connection"
  "github.com/pgct1/ambian-monitor/notification"
  "net/http"
  "regexp"
  "strconv"
)

func main() {

	ambianStreams := InitializeAmbianStreams()

  DataStream := make(chan notification.Packet)

  go connection.InitializeConnectionManager(SubscriptionPassword,DataStream)

  go TwitterStream(DataStream,ambianStreams)

  go ArticleStream(DataStream,ambianStreams)

  martiniServerSetup := martini.Classic()

	martiniServerSetup.Use(cors.Allow(&cors.Options{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET"},
		AllowHeaders:     []string{"Origin"},
		AllowCredentials: true,
	}))

	martiniServerSetup.Get("/json", func(w http.ResponseWriter, r *http.Request) {

		ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)

		if _, ok := err.(websocket.HandshakeError); ok {

			http.Error(w, "Invalid websocket handshake", 400)

			return
		} else if err != nil {
			return
		}

		connection.WebsocketConnectionHandler(ws)

	})

  martiniServerSetup.Get("/currentTopHeadlines/:ambianStreamId", func(res http.ResponseWriter, params martini.Params) (int,string) {

    valid, _ := regexp.MatchString("^\\d+$", params["ambianStreamId"])

    if !valid {
      return 400,"Invalid stream ID."
    }

    ambianStreamId,_ := strconv.Atoi(params["ambianStreamId"])

    json,err := CurrentNewsSourceTopHeadlinesAsJson(ambianStreamId)

    if err != nil {
      return 500,err.Error()
    }else{
      res.Header().Set("Content-Type", "application/json")
      return 200,json
    }

  })

  martiniServerSetup.Use(martini.Static("web"))

  martiniServerSetup.Run()
}
