package main

import(
	"code.google.com/p/go.net/websocket"
  "fmt"
  "time"
 )

// packets

type AuthorizationPacket struct {
	Password  string
}

type NotificationPacket struct {
	Type string
	Content string
}

type BroadcastChannel struct{
	Channel chan NotificationPacket
	Active bool
}

var AvailableChannels chan *BroadcastChannel
var BroadcastSecretary chan bool

func websocketConnectionHandler(ws *websocket.Conn) {

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

    // this connection is now authorized, so add a channel for it and listen

    var broadcastChannel* BroadcastChannel

    // acquire an available channel; if we don't get one immediately, request a new BroadcastChannel
    // from the BroadcastSecretary

    select{
	  	case broadcastChannel = <- AvailableChannels:
	  	case <- time.After(1*time.Second):
	  		BroadcastSecretary <- true
	  		broadcastChannel = <- AvailableChannels
	  }

    NotificationChannel := broadcastChannel.Channel
    broadcastChannel.Active = true

    L: for {

    	select{

    		case notification := <- NotificationChannel:
    			err := websocket.JSON.Send(ws,notification)
    			if err != nil {
    				break L
    			}

    	}

    }

    // this connection is closed, so set the channel to inactive and make it available to
    // whoever else wants to use it

    broadcastChannel.Active = false

    AvailableChannels <- broadcastChannel

  }

}

func InitializeConnectionManager(){

	greeting := NotificationPacket{Type:"greeting",Content:"Hi"}

	go func(){

		AvailableChannels = make(chan *BroadcastChannel)
		BroadcastSecretary = make(chan bool)

		BroadcastChannels := make([]BroadcastChannel,0)

		for{
			
			select{

				case <- BroadcastSecretary:

					// someone is waiting for a channel, and none are available, so make a new one for them

					BroadcastChannels = append(BroadcastChannels,BroadcastChannel{make(chan NotificationPacket),false})

					AvailableChannels <- &(BroadcastChannels[len(BroadcastChannels) - 1])

				case <- time.After(1*time.Second):

					for _,channel := range BroadcastChannels {
						if channel.Active{
							channel.Channel <- greeting
						}

					}

					// TODO: delete unused channels after heavy traffic calms down

			}

		}

	}()

}
