package main

import (
	"fmt"
	"github.com/araddon/httpstream"
	"github.com/mrjones/oauth"
	"log"
	"os"
	"strconv"
	"strings"
	"encoding/json"
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

type Tweet struct{
	Id string
	UserId string
	Username string
	Screenname string
	UserImageUrl string
	Followers int
	Text string
	Hashtags []string
}

// http stream

var httpStreamClient* httpstream.Client

func TwitterStream(DataStream chan NotificationPacket){

	// set corporate sources (why can't this just be a const declaration at the top... *sigh*)

	corporateSources = []string{"bbc","bbcworld","bbcone","bbcnew","bbcbreaking","washingtonpost","reutersworld","reuters","ajelive","ajenglish","guardian","guardiannews","nytimes","rt_com","theonion","cracked","techcrunch","verge","thenextweb"}

	// update our keyword list with the latest words from the global AmbianStreams

	keywords = make([]string,100)

	for _,stream := range AmbianStreams {

		for _,keyword := range stream.TwitterKeywords {

			keywords = append(keywords,keyword)

		}

	}

	rawStream := make(chan []byte)

	go startTwitterApiStream(rawStream,keywords)

	for{

		select{

			case rawTweetData := <- rawStream:
				tweet,err := parseTweet(rawTweetData)
				notification,err := TweetNotification(tweet)

				if err == nil {
					DataStream <- notification
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

func TweetNotification(tweet Tweet) (NotificationPacket,error){

	// determine which streams to assign this notification to

	AmbianStreamIds := make([]int,len(AmbianStreams))

	for _,stream := range AmbianStreams {

		// check for keyword matches

		keywordSearch: for _,keyword := range stream.TwitterKeywords {

			for _,hashtag := range tweet.Hashtags {

				if keyword == hashtag {
					AmbianStreamIds = append(AmbianStreamIds,stream.Id)
					break keywordSearch
				}

			}

			if strings.Contains(tweet.Text, keyword) {
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

	sources := Sources{Corporate:isCorporateSource,SocialMedia:true,Aggregate:false}

	metaData := NotificationMetaData{AmbianStreamIds,sources}

	jsonTweet,err := json.Marshal(tweet)

	var notification NotificationPacket

	if err == nil {
		notification = NotificationPacket{cNotificationTypeTweet,string(jsonTweet),metaData}
	}

	return notification,err

}

// parse the raw string into our Tweet object

type rawTweetEntityHashtag struct {
	Text string
}

type rawTweetEntities struct {
	Hashtags []rawTweetEntityHashtag
}

type rawTweetUser struct {
	Id_str string
	Name string
	Screen_name string
	Followers_count int
	Profile_image_url string
}

type rawTweetInterestingFields struct {
	Id_str string
	Text string
	User rawTweetUser
	Entities rawTweetEntities
}

func parseTweet(tw []byte) (Tweet,error) {

	var tweet Tweet

	var rawTweet rawTweetInterestingFields

	err := json.Unmarshal(tw, &rawTweet)

	if err == nil {
		tweet.Id = rawTweet.Id_str
		tweet.UserId = rawTweet.User.Id_str
		tweet.Username = rawTweet.User.Name
		tweet.Screenname = rawTweet.User.Screen_name
		tweet.UserImageUrl = rawTweet.User.Profile_image_url
		tweet.Followers = rawTweet.User.Followers_count
		tweet.Text = rawTweet.Text

		for i:=0;i<len(rawTweet.Entities.Hashtags);i++ {
			tweet.Hashtags = append(tweet.Hashtags,rawTweet.Entities.Hashtags[i].Text)
		}

	}else{
		fmt.Println(err)
	}

	return tweet,err

}

func startTwitterApiStream(output chan []byte, keywords []string){

	httpstream.SetLogger(log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile), logLevel)

	stream := make(chan []byte, 1000)
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
}
