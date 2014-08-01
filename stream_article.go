package main

import (
  "github.com/pgct1/ambian-monitor/notification"
  rss "github.com/jteeuwen/go-pkg-rss"
  "fmt"
  "time"
  "encoding/json"
)

type NewsSource struct {
  Name string
  RssUrl string
}

// careful! these names are displayed in the client, and also map to
// icons included in the client for these sources

var newsSources []NewsSource = []NewsSource{
  {"Reuters","http://mf.feeds.reuters.com/reuters/UKTopNews"},
  {"New York Times","http://rss.nytimes.com/services/xml/rss/nyt/InternationalHome.xml"},
  {"Russia Today","http://rt.com/rss/"},
  {"Washington Post","http://feeds.washingtonpost.com/rss/rss_blogpost"},
  {"Vice","https://news.vice.com/rss"},
  {"BBC International","http://feeds.bbci.co.uk/news/rss.xml?edition=int"},
  {"Al Jazeera","http://www.aljazeera.com/Services/Rss/?PostingId=2007731105943979989"},
  {"The Guardian","http://www.theguardian.com/world/rss"},
  {"The Telegraph","http://www.telegraph.co.uk/news/worldnews/rss"},
  {"The Huffington Post","http://www.huffingtonpost.com/feeds/verticals/world/index.xml"},
  {"The Wall Street Journal","http://online.wsj.com/xml/rss/3_7085.xml"},
  {"The Independent","http://rss.feedsportal.com/c/266/f/3503/index.rss"},
}

type NewsArticleNotification struct {
  Source string
  Content *rss.Item
}

var outputStream *chan notification.Packet

func PollFeed(newsSource NewsSource, timeout int) {

	feed := rss.New(timeout, true, chanHandler, func(feed *rss.Feed, ch *rss.Channel, newitems []*rss.Item){
    itemHandler(feed, ch, newitems, newsSource)
  })

	for {
		if err := feed.Fetch(newsSource.RssUrl, nil); err != nil {
			fmt.Println(err)
			return
		}

		<-time.After(time.Duration(feed.SecondsTillUpdate() * 1e9))
	}
}

func chanHandler(feed *rss.Feed, newchannels []*rss.Channel) {}

func itemHandler(feed *rss.Feed, ch *rss.Channel, newitems []*rss.Item, newsSource NewsSource) {

  for _,item := range(newitems) {

    newsArticleNotification := NewsArticleNotification{
      newsSource.Name,
      item,
    }

    jsonNewsItem,err := json.Marshal(newsArticleNotification)

    if err == nil {

      AmbianStreamIds := []int{1}

      sources := notification.Sources{
        Corporate:true,
        SocialMedia:false,
        Aggregate:false,
      }

      metaData := notification.MetaData{AmbianStreamIds,sources}

      n := notification.Packet{
        Type:notification.NotificationTypeOfficialNews,
        Content:string(jsonNewsItem),
        MetaData:metaData,
      }

      *outputStream <- n

    }

  }

}

func ArticleStream(DataStream chan notification.Packet) {

  outputStream = &DataStream

  for _,newsSource := range(newsSources) {

    go PollFeed(newsSource,10)

    <-time.After(time.Duration((600 / len(newsSources))) * time.Second)
  }
}
