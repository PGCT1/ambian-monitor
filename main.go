package main

import (
  "code.google.com/p/go.net/websocket"
  "github.com/go-martini/martini"
)

func main() {

  InitializeConnectionManager()

  martiniServerSetup := martini.Classic()

  martiniServerSetup.Get("/json", websocket.Handler(websocketConnectionHandler).ServeHTTP)

  martiniServerSetup.Use(martini.Static("web"))

  martiniServerSetup.Run()
}
