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

  updateNewsSourceTopHeadline(newsSource,ch.Items,ambianStreamId)

  pushableItems := filterPushableNewsItems(newitems)

  for _,item := range(pushableItems) {

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

func isFresh(newsItem *rss.Item) bool {

  yesterday := time.Now().Add(-1 * time.Hour * 24)

  publishDate,err := extractPublishDate(newsItem)

  if err == nil && publishDate.Before(yesterday) {
    return false
  }else{
    return true
  }

}

func filterPushableNewsItems(newsItems []*rss.Item) []*rss.Item {

  filteredItems := make([]*rss.Item,0,len(newsItems))

  for _,newsItem := range(newsItems) {

    if isFresh(newsItem) {
      filteredItems = append(filteredItems,newsItem)
    }

  }

  return filteredItems

}

func extractPublishDate(newsItem *rss.Item) (time.Time,error) {

  publishDate,err := time.Parse(time.RFC1123,newsItem.PubDate)

  if err != nil {
    publishDate,err = time.Parse(time.RFC1123Z,newsItem.PubDate)
  }

  return publishDate,err
}

func chooseTopHeadline(newsItems []*rss.Item) *rss.Item {

  /*
    choose the first element by default, unless it's > 24 hours old, in which
    case god knows what order their RSS feed is in, so we choose the newest
    article instead
  */

  firstItemPublishDate,err := extractPublishDate(newsItems[0])

  if err != nil {
    fmt.Println("PING2")
    fmt.Println(err)
    return newsItems[0]
  }

  if !isFresh(newsItems[0]) {

    mostRecentItem := newsItems[0]
    mostRecentPublishDate := firstItemPublishDate

    for i,newsItem := range(newsItems) {

      publishDate,err := extractPublishDate(newsItem)

      if err == nil {

        if publishDate.After(mostRecentPublishDate) {

          mostRecentItem = newsItems[i]
          mostRecentPublishDate = publishDate

        }

      }else{
        fmt.Println("PING1")
        fmt.Println(err)
      }

    }

    return mostRecentItem

  }

  return newsItems[0]

}

var currentTopHeadlinesMutex *sync.RWMutex

var currentTopHeadlines map[int]map[string]*rss.Item

func updateNewsSourceTopHeadline(newsSource NewsSource, newsItems []*rss.Item, ambianStreamId int){

  chosenTopHeadline := chooseTopHeadline(newsItems)

  currentTopHeadlinesMutex.Lock()

  currentTopHeadlines[ambianStreamId][newsSource.Name] = chosenTopHeadline

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
