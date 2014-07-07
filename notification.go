package main

// meta

type Feeds struct {
	WorldNews bool
	SocialEntertainment bool
}

type Sources struct {
	Corporate bool
	SocialMedia bool
	Aggregate bool
}

type NotificationMetaData struct {
	Feeds Feeds
	Sources Sources
}

const (
	cNotificationTypeDefault = iota
	cNotificationTypeOther = iota
)

// packets

type NotificationPacket struct {
	Type int
	Content string
	MetaData NotificationMetaData
}
