package main

import (
  "code.google.com/p/go.net/websocket"
  "github.com/go-martini/martini"
)

// globals

var AmbianStreams []AmbianStream	// active streams (world news, social & entertainment, etc.)

func main() {

	// just hardcode some database entries for now

	CreateAmbianStream(AmbianStream{
		Name:"World News",
		TwitterKeywords:[]string{"obama","putin","snowden","russia","benghazi","isil","merkel","kerry","clinton","brussels","moscow","washington"},
	})

	CreateAmbianStream(AmbianStream{
		Name:"Social & Entertainment",
		TwitterKeywords:[]string{"funny","til"},
	})

	AmbianStreams,_ = GetAmbianStreams()

  InitializeConnectionManager()

  martiniServerSetup := martini.Classic()

  martiniServerSetup.Get("/json", websocket.Handler(websocketConnectionHandler).ServeHTTP)

  martiniServerSetup.Use(martini.Static("web"))

  martiniServerSetup.Run()
}
