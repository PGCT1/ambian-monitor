package main

import (
	"fmt"
	"github.com/araddon/httpstream"
	"github.com/mrjones/oauth"
	"log"
	"os"
	"strconv"
	"strings"
	"math/rand"
	"encoding/json"
)

// terms

const search = "obama"
var keywords []string

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
	UserImageUrl string
	Followers int
	Text string
	Hashtags []string
}

func TwitterStream(DataStream chan NotificationPacket){

	keywords := []string{"obama","putin"}

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

// analyze the tweet and construct (if desirable) a notification from it

func TweetNotification(tweet Tweet) (NotificationPacket,error){

	feeds := Feeds{WorldNews:rand.Intn(2) != 0,SocialEntertainment:rand.Intn(2) != 0}
	sources := Sources{Corporate:rand.Intn(2) != 0,SocialMedia:rand.Intn(2) != 0,Aggregate:rand.Intn(2) != 0}

	metaData := NotificationMetaData{feeds,sources}

	jsonTweet,err := json.Marshal(tweet)

	var notification NotificationPacket

	if err == nil {
		notification = NotificationPacket{cNotificationTypeDefault,string(jsonTweet),metaData}
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

	client := httpstream.NewOAuthClient(&at, httpstream.OnlyTweetsFilter(func(line []byte) {
		stream <- line
	}))

	// parse userIds

	userIds := make([]int64, 0)

	for _, userId := range strings.Split(users, ",") {
		if id, err := strconv.ParseInt(userId, 10, 64); err == nil {
			userIds = append(userIds, id)
		}
	}

	err := client.Filter(userIds, keywords, []string{"en"}, nil, false, done)

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
