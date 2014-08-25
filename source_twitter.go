package main

import (
	"github.com/araddon/httpstream"
	"github.com/mrjones/oauth"
	"github.com/pgct1/ambian-monitor/notification"
	"github.com/pgct1/ambian-monitor/tweet"
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

				t,err := parseTweet(rawTweetData)

				if err == nil {

					// next, check the id to make sure we haven't already
					// sent it out recently.

					if isRecent(t) && hasntBeenSentRecently(t) {

						n,err := TweetNotification(t)

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

func TweetNotification(t tweet.Tweet) (notification.Packet,error){

	var n notification.Packet

	if t.Followers < 2000 {
		return n,TweetNotificationError{"Uninteresting"}
	}

	// determine if this is a corporate source or not

	isCorporateSource := false

	screenname := strings.ToLower(t.Screenname)

	determinedCorporateStatus: for _,corporateSource := range corporateSources {

		if screenname == corporateSource {
			isCorporateSource = true
			break determinedCorporateStatus
		}

	}

	// determine which streams to assign this notification to

	AmbianStreamIds := make([]int,0,len(AmbianStreams))

	foundRelevantStream := false

	for _,stream := range AmbianStreams {

		if stream.Filter(t, stream.TwitterKeywords, isCorporateSource) {
			foundRelevantStream = true
			AmbianStreamIds = append(AmbianStreamIds,stream.Id)
		}

	}

	if !foundRelevantStream {
		return notification.Packet{},TweetNotificationError{Message:"Irrelevant"}
	}

	sources := notification.Sources{Corporate:isCorporateSource,SocialMedia:true,Aggregate:false}

	metaData := notification.MetaData{AmbianStreamIds,sources}

	jsonTweet,err := json.Marshal(t)

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

func hasntBeenSentRecently(t tweet.Tweet) bool {

	found := false

	staleListMutex.RLock()

	L1: for _,id := range(staleIdsListA) {
		if id == t.Id {
			found = true
			break L1
		}
	}

	if !found{
		L2: for _,id := range(staleIdsListB) {
			if id == t.Id {
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

	staleIdsListA = append(staleIdsListA,t.Id)
	staleIdsListB = append(staleIdsListB,t.Id)

	staleListMutex.Unlock()

	return true

}

func isRecent(t tweet.Tweet) bool {
	return t.CreatedAt > time.Now().Unix() - 24*60*60
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
	Verified bool
	Followers_count int
	Profile_image_url string
}


type rawTweetInterestingFields struct {
	Id int64
	Created_at string
	Retweeted_status *rawTweetInterestingFields
	Text string
	User rawTweetUser
	Entities rawTweetEntities
}

func parseTweet(tw []byte) (tweet.Tweet,error) {

	var t tweet.Tweet

	rawTweet := new(rawTweetInterestingFields)
	rawTweet.Retweeted_status = new(rawTweetInterestingFields)

	err := json.Unmarshal(tw, &rawTweet)

	if err == nil {

		// first, check and see if this is a retweet
		// if so, just parse the original and ignore the rt

		if rawTweet.Retweeted_status.Id > 0 {
			rawTweet = rawTweet.Retweeted_status
		}

		createdAt,err := time.Parse(time.RubyDate,rawTweet.Created_at)

		if err != nil {
			fmt.Println(err)
		}

		t.Id = rawTweet.Id
		t.CreatedAt = createdAt.Unix()
		t.UserId = rawTweet.User.Id
		t.Username = rawTweet.User.Name
		t.Screenname = rawTweet.User.Screen_name
		t.Verified = rawTweet.User.Verified
		t.UserImageUrl = rawTweet.User.Profile_image_url
		t.Followers = rawTweet.User.Followers_count
		t.Text = rawTweet.Text

		for i:=0;i<len(rawTweet.Entities.Hashtags);i++ {
			t.Hashtags = append(t.Hashtags,rawTweet.Entities.Hashtags[i].Text)
		}

		for i:=0;i<len(rawTweet.Entities.Media);i++ {
			t.Media = append(t.Media,tweet.TweetMedia{
				Id:rawTweet.Entities.Media[i].Id,
				Url:rawTweet.Entities.Media[i].Media_url,
				Type:rawTweet.Entities.Media[i].Type,
			})
		}

	}else{
		fmt.Println(err)
	}

	return t,err

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
