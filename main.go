package main

import (
  "github.com/gorilla/websocket"
  "github.com/go-martini/martini"
  "github.com/martini-contrib/cors"
  "net/http"
)

// globals

var AmbianStreams []AmbianStream	// active streams (world news, social & entertainment, etc.)

func main() {

	// just hardcode some database entries for now

	CreateAmbianStream(AmbianStream{
		Name:"World News",
		TwitterKeywords:[]string{"syria","egypt","hamas","idf","palestine","gaza","putin","snowden","russia","benghazi","isil","merkel","kerry","clinton","brussels","moscow","washington"},
	})

	CreateAmbianStream(AmbianStream{
		Name:"Social & Entertainment",
		TwitterKeywords:[]string{"harhar"},
	})

	AmbianStreams,_ = GetAmbianStreams()

  InitializeConnectionManager()

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

		websocketConnectionHandler(ws)

	})

  martiniServerSetup.Use(martini.Static("web"))

  martiniServerSetup.Run()
}
