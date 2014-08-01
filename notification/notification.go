package notification

// meta

type Sources struct {
	Corporate bool
	SocialMedia bool
	Aggregate bool
}

type MetaData struct {
	AmbianStreamIds []int
	Sources Sources
}

const (
	NotificationTypeDefault = iota
	NotificationTypeTweet
	NotificationTypeOfficialNews
)

// packets

type Packet struct {
	Type int
	Content string
	MetaData MetaData
}
