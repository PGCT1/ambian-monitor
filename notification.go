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

// packets

type NotificationPacket struct {
	Type string
	Content string
	MetaData NotificationMetaData
}
