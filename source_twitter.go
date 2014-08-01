package main

import (
	"github.com/araddon/httpstream"
	"github.com/mrjones/oauth"
	"github.com/pgct1/ambian-monitor/notification"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"encoding/json"
	"time"
	"sync"
)

// terms

var keywords []string
var corporateSources []string

// access keys

const apiKey = TwitterApiKey
const apiSecret = TwitterApiSecret
const accessToken = TwitterAccessToken
const accessTokenSecret = TwitterAccessTokenSecret

// internal constants

const logLevel = "warn"
const users = ""

type TweetMedia struct{
	Id int64
	Url string
	Type string
}

type Tweet struct{
	Id int64
	UserId int64
	Username string
	Screenname string
	UserImageUrl string
	Followers int
	Text string
	Hashtags []string
	Media []TweetMedia
}

// http stream

var httpStreamClient* httpstream.Client

func TwitterStream(DataStream chan notification.Packet){

	// set corporate sources (why can't this just be a const declaration at the top... *sigh*)

	corporateSources = []string{"bbc","bbcworld","bbcone","bbcnew","bbcbreaking","washingtonpost","reutersworld","reuters","ajelive","ajenglish","guardian","guardiannews","nytimes","rt_com","theonion","cracked","techcrunch","verge","thenextweb"}

	// update our keyword list with the latest words from the global AmbianStreams

	keywords = make([]string,0,100)

	for _,stream := range AmbianStreams {

		for _,keyword := range stream.TwitterKeywords {

			keywords = append(keywords,keyword)

		}

	}

	rawStream := make(chan []byte)

	freshnessCheckInit()

	go startTwitterApiStream(rawStream,keywords)

	for{

		select{

			case rawTweetData := <- rawStream:

				// new incoming tweet data, so first parse it

				tweet,err := parseTweet(rawTweetData)

				if err == nil {

					// next, check the id to make sure we haven't already
					// sent it out recently.

					if isFresh(tweet) {

						n,err := TweetNotification(tweet)

						if err == nil {
							DataStream <- n
						}

					}

				}

		}

	}

}

func StopTwitterStream(){

	httpStreamClient.Close()

}

// analyze the tweet and construct a notification from it

type TweetNotificationError struct{
	Message string
}

func (e TweetNotificationError) Error() string {
	return e.Message
}

func TweetNotification(tweet Tweet) (notification.Packet,error){

	var n notification.Packet

	if tweet.Followers < 2000 {
		return n,TweetNotificationError{"Uninteresting"}
	}

	// determine which streams to assign this notification to

	AmbianStreamIds := make([]int,0,len(AmbianStreams))

	for _,stream := range AmbianStreams {

		// check for keyword matches

		keywordSearch: for _,keyword := range stream.TwitterKeywords {

			for _,hashtag := range tweet.Hashtags {

				if keyword == hashtag {
					AmbianStreamIds = append(AmbianStreamIds,stream.Id)
					break keywordSearch
				}

			}

			text := strings.ToLower(tweet.Text)

			if strings.Contains(text, keyword) {
				AmbianStreamIds = append(AmbianStreamIds,stream.Id)
				break keywordSearch
			}

		}

	}

	// determine if this is a corporate source or not

	isCorporateSource := false

	screenname := strings.ToLower(tweet.Screenname)

	determinedCorporateStatus: for _,corporateSource := range corporateSources {

		if screenname == corporateSource {
			isCorporateSource = true
			break determinedCorporateStatus
		}

	}

	sources := notification.Sources{Corporate:isCorporateSource,SocialMedia:true,Aggregate:false}

	metaData := notification.MetaData{AmbianStreamIds,sources}

	jsonTweet,err := json.Marshal(tweet)

	if err == nil {
		n = notification.Packet{notification.NotificationTypeTweet,string(jsonTweet),metaData}
	}

	return n,err

}

// freshness - since twitter sends us retweets, but we show originals,
// we need to check and make sure we aren't frequently pushing duplicates.
// to do this, just use a simple algorithm: two lists, which are cleared
// at 6 minute intervals (offset by 3 min). any time a new tweet comes in,
// if it is in either list, reject it; if it's not in either list,
// accept it and put it in both lists.

// this is not thread-safe, but it is memory access safe, and race conditions
// happen at most once per 6 minutes and have the side effect of possibly
// incorrectly judging a tweet as unfresh when it is, in fact, fresh again.

const cStaleListSize = 1000
var staleListMutex *sync.RWMutex

var staleIdsListA []int64
var staleIdsListB []int64

func freshnessCheckInit (){

	staleListMutex = new(sync.RWMutex)

	staleIdsListA = make([]int64,0,cStaleListSize)
	staleIdsListB = make([]int64,0,cStaleListSize)

	go func(){

		staleList := false

		for {

			select {
				case <- time.After(3*time.Minute):

					staleListMutex.Lock()

					if staleList {
						staleIdsListA = staleIdsListA[:0]
					}else{
						staleIdsListB = staleIdsListB[:0]
					}

					staleListMutex.Unlock()

					staleList = !staleList
			}

		}

	}()

}

func isFresh(tweet Tweet) bool {

	found := false

	staleListMutex.RLock()

	L1: for _,id := range(staleIdsListA) {
		if id == tweet.Id {
			found = true
			break L1
		}
	}

	if !found{
		L2: for _,id := range(staleIdsListB) {
			if id == tweet.Id {
				found = true
				break L2
			}
		}
	}

	staleListMutex.RUnlock()

	if found {
		return false
	}

	staleListMutex.Lock()

	staleIdsListA = append(staleIdsListA,tweet.Id)
	staleIdsListB = append(staleIdsListB,tweet.Id)

	staleListMutex.Unlock()

	return true

}

// parse the raw string into our Tweet object

type rawTweetEntityHashtag struct {
	Text string
}

type rawTweetEntityMedia struct {
	Id int64
	Media_url string
	Type string
}

type rawTweetEntities struct {
	Hashtags []rawTweetEntityHashtag
	Media []rawTweetEntityMedia
}

type rawTweetUser struct {
	Id int64
	Name string
	Screen_name string
	Followers_count int
	Profile_image_url string
}


type rawTweetInterestingFields struct {
	Id int64
	Retweeted_status *rawTweetInterestingFields
	Text string
	User rawTweetUser
	Entities rawTweetEntities
}

func parseTweet(tw []byte) (Tweet,error) {

	var tweet Tweet

	rawTweet := new(rawTweetInterestingFields)
	rawTweet.Retweeted_status = new(rawTweetInterestingFields)

	err := json.Unmarshal(tw, &rawTweet)

	if err == nil {

		// first, check and see if this is a retweet
		// if so, just parse the original and ignore the rt

		if rawTweet.Retweeted_status.Id > 0 {
			rawTweet = rawTweet.Retweeted_status
		}

		tweet.Id = rawTweet.Id
		tweet.UserId = rawTweet.User.Id
		tweet.Username = rawTweet.User.Name
		tweet.Screenname = rawTweet.User.Screen_name
		tweet.UserImageUrl = rawTweet.User.Profile_image_url
		tweet.Followers = rawTweet.User.Followers_count
		tweet.Text = rawTweet.Text

		for i:=0;i<len(rawTweet.Entities.Hashtags);i++ {
			tweet.Hashtags = append(tweet.Hashtags,rawTweet.Entities.Hashtags[i].Text)
		}

		for i:=0;i<len(rawTweet.Entities.Media);i++ {
			tweet.Media = append(tweet.Media,TweetMedia{
				Id:rawTweet.Entities.Media[i].Id,
				Url:rawTweet.Entities.Media[i].Media_url,
				Type:rawTweet.Entities.Media[i].Type,
			})
		}

	}else{
		fmt.Println(err)
	}

	return tweet,err

}

func startTwitterApiStream(output chan []byte, keywords []string){

	httpstream.SetLogger(log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile), logLevel)

	stream := make(chan []byte, 2000)
	done := make(chan bool)

	httpstream.OauthCon = oauth.NewConsumer(
		apiKey,
		apiSecret,
		oauth.ServiceProvider{
			RequestTokenUrl:   "http://api.twitter.com/oauth/request_token",
			AuthorizeTokenUrl: "https://api.twitter.com/oauth/authorize",
			AccessTokenUrl:    "https://api.twitter.com/oauth/access_token",
		})

	at := oauth.AccessToken{
		Token:  accessToken,
		Secret: accessTokenSecret,
	}

	httpStreamClient = httpstream.NewOAuthClient(&at, httpstream.OnlyTweetsFilter(func(line []byte) {
		stream <- line
	}))

	// parse userIds

	userIds := make([]int64, 0)

	for _, userId := range strings.Split(users, ",") {
		if id, err := strconv.ParseInt(userId, 10, 64); err == nil {
			userIds = append(userIds, id)
		}
	}

	err := httpStreamClient.Filter(userIds, keywords, []string{"en"}, nil, false, done)

	if err != nil {
		httpstream.Log(httpstream.ERROR, err.Error())
	} else {

		go func() {
			for tw := range stream {
				output <- tw
			}
		}()
		_ = <-done
	}

	fmt.Println("TWEET STREAM STOPPED")
}
