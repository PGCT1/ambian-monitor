package main

import (
  "github.com/gorilla/websocket"
  "github.com/go-martini/martini"
  "github.com/martini-contrib/cors"
  "github.com/pgct1/ambian-monitor/connection"
  "github.com/pgct1/ambian-monitor/notification"
  "github.com/pgct1/ambian-monitor/tweet"
  "net/http"
  "encoding/json"
  "fmt"
)

// globals

var AmbianStreams []AmbianStream	// active streams (world news, social & entertainment, etc.)

func main() {

  tweet.Test()

	// just hardcode some database entries for now

	CreateAmbianStream(AmbianStream{
		Name:"World News",
		TwitterKeywords:[]string{"syria","egypt","hamas","idf","palestine","gaza","putin","snowden","russia","benghazi","isil","merkel","kerry","clinton","ferguson","brussels","moscow","washington"},
    Filter:func(t tweet.Tweet, keywords []string, isCorporateSource bool) bool {

      // make sure this isn't irrelevant shit like cam girl ads or something

      sexyKeywords := []string{"webcam","girls","hawt","sex","camgirls","horny","cam","kinky"}

      if !isCorporateSource {  // automatically trust major news sources

        match := t.KeywordMatch(sexyKeywords)

        if match == true {
          jsonTweet,_ := json.Marshal(t)
          fmt.Println(string(jsonTweet))
          fmt.Println(",")
          return false
        }

      }

      // can't detect anything in our blacklists, so if we have a keyword
      // match, let it through

      return t.KeywordMatch(keywords)

    },
	})

	CreateAmbianStream(AmbianStream{
		Name:"Social & Entertainment",
		TwitterKeywords:[]string{"harhar"},
    Filter:func(t tweet.Tweet, keywords []string, isCorporateSource bool) bool {
      return true
    },
	})

	AmbianStreams,_ = GetAmbianStreams()

  DataStream := make(chan notification.Packet)

  go connection.InitializeConnectionManager(SubscriptionPassword,DataStream)

  go TwitterStream(DataStream)

  go ArticleStream(DataStream)

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

  martiniServerSetup.Get("/currentTopHeadlines", func(res http.ResponseWriter) (int,string) {

    json,err := CurrentNewsSourceTopHeadlinesAsJson()

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
