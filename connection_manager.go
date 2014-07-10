package main

import(
	"github.com/gorilla/websocket"
  "fmt"
  "time"
  "encoding/json"
 )

/*

	Flow: 
		1. A new connection C is received and handled by websocketConnectionHandler
		2. C must provide an AuthorizationPacket immediately
			if C doesn't or the metadata is invalid
				close C
			if yes
				acquire a new broadcastChannel from AvailableChannels or from the BroadcastSecretary if none are available
				set the broadcastChannel to the AuthorizationPacket.DesiredStreams
				listen to the broadcastChannel and push anything that flows in over the websocket

*/

type AuthorizationPacket struct {
	Password string
	DesiredStreams NotificationMetaData
}

type BroadcastChannel struct{
	Channel chan NotificationPacket
	Active bool
	MetaData NotificationMetaData
}

// Global channels

var AvailableChannels chan *BroadcastChannel
var BroadcastSecretary chan bool

// Websocket connection handler

func websocketConnectionHandler(ws *websocket.Conn) {

  defer ws.Close()

  // they should send a password immediately; if they don't, close the connection

  authorized := false
  var DesiredMetaData NotificationMetaData

  password := func () chan string{

    channel := make(chan string)

    go func(){

      var authResponse AuthorizationPacket

      _, message, err := ws.ReadMessage()

      if err == nil {
      	err = json.Unmarshal(message,&authResponse)
      }

      //err := websocket.JSON.Receive(ws, &authResponse)

      if err == nil {

      	DesiredMetaData = authResponse.DesiredStreams

        channel <- authResponse.Password
      }

    }()

    return channel
  }()

  select {

    case pw := <- password:
      if pw == SubscriptionPassword {
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

    broadcastChannel.MetaData = DesiredMetaData

    broadcastChannel.Active = true

    L: for {

    	select{

    		case notification := <- NotificationChannel:

					raw, err := json.Marshal(notification)

					if err == nil {
	    			err = ws.WriteMessage(websocket.TextMessage,raw)
					}

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

// init

func InitializeConnectionManager(){

	DataStream := make(chan NotificationPacket)

	go monitorDataSources(DataStream)

	go channelManager(DataStream)

}

// pushes data to the channel each time it has a new notification from a source

func monitorDataSources(DataStream chan NotificationPacket){

	go TwitterStream(DataStream)

}

// handles adding / removing channels, broadcasting notifications from the DataStream

func channelManager(DataStream chan NotificationPacket){

	defaultFeeds := []int{1,2}
	defaultSources := Sources{Corporate:true,SocialMedia:true,Aggregate:true}

	DefaultSubscription := NotificationMetaData{defaultFeeds,defaultSources}

	AvailableChannels = make(chan *BroadcastChannel)
	BroadcastSecretary = make(chan bool)

	BroadcastChannels := make([]BroadcastChannel,0)

	for{
		
		select{

			case <- BroadcastSecretary:

				// someone is waiting for a channel, and none are available, so make a new one for them

				BroadcastChannels = append(BroadcastChannels,BroadcastChannel{make(chan NotificationPacket),false,DefaultSubscription})

				AvailableChannels <- &(BroadcastChannels[len(BroadcastChannels) - 1])

			case notification := <- DataStream:

				for _,channel := range BroadcastChannels {

					// make sure this channel actually has a listener

					if channel.Active{

						// make sure this notification is from a desired feed

						notificationIsFromDesiredFeed := false

						found: for _,desiredFeed := range channel.MetaData.AmbianStreamIds {
							for _,targetedFeed := range notification.MetaData.AmbianStreamIds{
								if desiredFeed == targetedFeed {
									notificationIsFromDesiredFeed = true
									break found
								}
							}
						}

						if (notificationIsFromDesiredFeed){

						  // make sure this notification is from a desired source

						  if (channel.MetaData.Sources.Corporate == true &&  notification.MetaData.Sources.Corporate == true) || 
						     (channel.MetaData.Sources.SocialMedia == true &&  notification.MetaData.Sources.SocialMedia == true) || 
						     (channel.MetaData.Sources.Aggregate == true &&  notification.MetaData.Sources.Aggregate == true){

								channel.Channel <- notification

							}

						}

					}

				}

				// TODO: delete unused channels after heavy traffic calms down

		}

	}

}
