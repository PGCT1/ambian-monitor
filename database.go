package main

import (
	"strings"
)

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

	for _,stream := range ambianStreams {
		if stream.Id == ambianStream.Id {
			panic("Duplicate stream IDs found.")
		}
	}

	ambianStreams = append(ambianStreams,ambianStream)

	applyCaseRules(ambianStream.Id)

	return ambianStream.Id,nil
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
