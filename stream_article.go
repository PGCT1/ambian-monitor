package main

import (
  "github.com/pgct1/ambian-monitor/notification"
  rss "github.com/jteeuwen/go-pkg-rss"
  "fmt"
  "time"
  "encoding/json"
  "sync"
)

type NewsSource struct {
  Name string
  RssUrl string
}

type NewsArticleNotification struct {
  Source string
  Content *rss.Item
}

var outputStream *chan notification.Packet

func PollFeed(newsSource NewsSource, timeout int, ambianStreamId int) {

	feed := rss.New(timeout, true, chanHandler, func(feed *rss.Feed, ch *rss.Channel, newitems []*rss.Item){
    itemHandler(feed, ch, newitems, newsSource, ambianStreamId)
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

func itemHandler(feed *rss.Feed, ch *rss.Channel, newitems []*rss.Item, newsSource NewsSource, ambianStreamId int) {

  updateNewsSourceTopHeadline(newsSource,ch.Items[0],ambianStreamId)

  for _,item := range(newitems) {

    newsArticleNotification := NewsArticleNotification{
      newsSource.Name,
      item,
    }

    jsonNewsItem,err := json.Marshal(newsArticleNotification)

    if err == nil {

      AmbianStreamIds := []int{ambianStreamId}

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

var currentTopHeadlinesMutex *sync.RWMutex

var currentTopHeadlines map[int]map[string]*rss.Item

func updateNewsSourceTopHeadline(newsSource NewsSource, newsItem *rss.Item, ambianStreamId int){

  currentTopHeadlinesMutex.Lock()

  currentTopHeadlines[ambianStreamId][newsSource.Name] = newsItem

  currentTopHeadlinesMutex.Unlock()

}

func CurrentNewsSourceTopHeadlinesAsJson(ambianStreamId int) (string,error) {

  currentTopHeadlinesMutex.RLock()

  defer currentTopHeadlinesMutex.RUnlock()

  fmt.Println(ambianStreamId)

  jsonTopHeadlines,err := json.Marshal(currentTopHeadlines[ambianStreamId])

  if err != nil {
    fmt.Println(err)
    return "",err
  }else{
    return string(jsonTopHeadlines),err
  }

}

func ArticleStream(DataStream chan notification.Packet, ambianStreams []AmbianStream) {

  currentTopHeadlines = make(map[int]map[string]*rss.Item)

  currentTopHeadlinesMutex = new(sync.RWMutex)

  outputStream = &DataStream

  for _,stream := range ambianStreams {

    currentTopHeadlines[stream.Id] = make(map[string]*rss.Item)

    for _,newsSource := range(stream.NewsSources) {

      go PollFeed(newsSource,10,stream.Id)

      <-time.After(time.Duration((600 / len(stream.NewsSources))) * time.Second)
    }

  }

}
