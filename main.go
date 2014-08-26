package main

import (
  "github.com/gorilla/websocket"
  "github.com/go-martini/martini"
  "github.com/martini-contrib/cors"
  "github.com/pgct1/ambian-monitor/connection"
  "github.com/pgct1/ambian-monitor/notification"
  "github.com/pgct1/ambian-monitor/tweet"
  "net/http"
	"regexp"
  "strconv"
)

// globals

var AmbianStreams []AmbianStream	// active streams (world news, social & entertainment, etc.)

func main() {

  tweet.Test()

	// just hardcode some database entries for now

	CreateAmbianStream(AmbianStream{
		Name:"World News",
		TwitterKeywords:[]string{"syria","egypt","hamas","idf","palestine","gaza","putin","snowden","russia","benghazi","isil","merkel","kerry","clinton","ferguson","brussels","moscow","washington"},
    NewsSources:[]NewsSource{
      // careful! these names (keys) are displayed in the client, and also map to
      // icons included in the client for these sources
      {"New York Times","http://rss.nytimes.com/services/xml/rss/nyt/InternationalHome.xml"},
      {"BBC International","http://feeds.bbci.co.uk/news/rss.xml?edition=int"},
      {"Reuters","http://mf.feeds.reuters.com/reuters/UKTopNews"},
      {"Al Jazeera","http://www.aljazeera.com/Services/Rss/?PostingId=2007731105943979989"},
      {"Washington Post","http://feeds.washingtonpost.com/rss/rss_blogpost"},
      {"Vice","https://news.vice.com/rss"},
      {"The Guardian","http://www.theguardian.com/world/rss"},
      {"Russia Today","http://rt.com/rss/"},
      {"The Wall Street Journal","http://online.wsj.com/xml/rss/3_7085.xml"},
      {"The Huffington Post","http://www.huffingtonpost.com/feeds/verticals/world/index.xml"},
      {"The Independent","http://rss.feedsportal.com/c/266/f/3503/index.rss"},
      {"The Telegraph","http://www.telegraph.co.uk/news/worldnews/rss"},
    },
    Filter:func(t tweet.Tweet, keywords []string, isCorporateSource bool) bool {

      // make sure this isn't irrelevant shit like cam girl ads or something

      sexyKeywords := []string{"webcam","girls","ass","fuck","bitch","boob","boobs","tit","tits","luv","hawt","sex","camgirls","horny","cam","kinky"}

      if !isCorporateSource {  // trust major news sources

        match := t.KeywordMatch(sexyKeywords)

        if match == true {
          return false
        }

      }

      // other irrelevant stuff that appears a lot

      irrelevant := []string{"concert","in-concert","1d","touchdown","football","home run","kardashian","kardashians","bieber","belieber","homerun","emmy","emmys","emmys2014"}

      if !isCorporateSource {  // trust major news sources

        match := t.KeywordMatch(irrelevant)

        if match == true {
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
    NewsSources:[]NewsSource{},
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
