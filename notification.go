package main

// meta

type Sources struct {
	Corporate bool
	SocialMedia bool
	Aggregate bool
}

type NotificationMetaData struct {
	AmbianStreamIds []int
	Sources Sources
}

const (
	cNotificationTypeDefault = iota
	cNotificationTypeTweet = iota
)

// packets

type NotificationPacket struct {
	Type int
	Content string
	MetaData NotificationMetaData
}
