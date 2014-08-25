package main

import (
	"github.com/pgct1/ambian-monitor/tweet"
	"strings"
)

// for now, just an interface with no actual foreign database connection

type AmbianStream struct {
	Id int
	Name string

	// source metadata

	TwitterKeywords []string
	Filter func(tweet.Tweet, []string, bool)bool
}

// our 'database'

var ambianStreams []AmbianStream

type DatabaseError struct {
	Message string
}

func (e DatabaseError) Error() string {
	return e.Message
}

func GetAmbianStreams() ([]AmbianStream,error) {
	return ambianStreams,nil
}

func CreateAmbianStream(ambianStream AmbianStream) (int,error) {

	// find highest Id and add one

	highestFound := 0

	for _,stream := range ambianStreams {
		if stream.Id > highestFound {
			highestFound = stream.Id
		}
	}

	newId := highestFound + 1

	newAmbianStream := ambianStream

	newAmbianStream.Id = newId

	ambianStreams = append(ambianStreams,newAmbianStream)

	applyCaseRules(newId)

	return newId,nil
}

func UpdateAmbianStream(ambianStream AmbianStream) error {

	found := false
	var foundId int

	for _,stream := range ambianStreams {

		if stream.Id == ambianStream.Id {

			found = true
			foundId = stream.Id

			stream = ambianStream

			break
		}

	}

	if found {
		applyCaseRules(foundId)
		return nil
	}else{
		return DatabaseError{"No stream found with the specified Id"}
	}

}

func DeleteAmbianStream(ambianStreamId int) error {

	// super shitty algorithm, but n is very small (as in... around 3), so linear time is fine
	// because it's readable to me

	newAmbianStreams := make([]AmbianStream,len(ambianStreams))

	for _,ambianStream := range ambianStreams {

		if ambianStream.Id != ambianStreamId {

			newAmbianStreams = append(newAmbianStreams,ambianStream)

		}

	}

	ambianStreams = newAmbianStreams

	return nil

}

func applyCaseRules(id int){

	// find this stream

	for i:=0;i<len(ambianStreams);i++ {

		if ambianStreams[i].Id == id {

			// set keywords and such to lower case

			for j,twitterKeyword := range ambianStreams[i].TwitterKeywords {
				ambianStreams[i].TwitterKeywords[j] = strings.ToLower(twitterKeyword)
			}

			break
		}

	}

}
