package main

import (
	"code.google.com/p/go.net/websocket"
	"github.com/go-martini/martini"
  "fmt"
  "time"
)

type AuthorizationPacket struct {
	Password  string
}

func websocketConnectionHandler(ws *websocket.Conn) {

	fmt.Printf("jsonServer %#v\n", ws.Config())

  defer ws.Close()

  // they should send a password immediately; if they don't, close the connection

  authorized := false

  password := func () chan string{

    channel := make(chan string)

    go func(){

      var authResponse AuthorizationPacket

      err := websocket.JSON.Receive(ws, &authResponse)

      if err == nil {
        channel <- authResponse.Password
      }

    }()

    return channel
  }()

  select {

    case pw := <- password:
      if pw == "valid" {
        authorized = true
      }

    case <- time.After(3*time.Second):
      fmt.Println("Authorization timeout.")
  }

  if !authorized {
    ws.Close()
  }else{

    // they're authorized now, so... do something

    stop := make(chan bool)

    for {

      select {
        case <-stop:
            return
      }

      // Send send a text message serialized T as JSON.

      // err = websocket.JSON.Send(ws, msg)

      // if err != nil {
      //   fmt.Println(err)
      //   break
      // }

      // fmt.Printf("send:%#v\n", msg)

    }

  }

}

func main() {

	martiniServerSetup := martini.Classic()

	martiniServerSetup.Get("/json", websocket.Handler(websocketConnectionHandler).ServeHTTP)

  martiniServerSetup.Use(martini.Static("web"))

	martiniServerSetup.Run()
}
