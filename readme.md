# Ambian Monitor
Backend data monitoring service used by Ambian
## Connection Flow

1. A new connection C is received and handled by websocketConnectionHandler
2. C must provide an AuthorizationPacket immediately

    - if C doesn't or the metadata is invalid

	    - close C

	- if yes
		- acquire a new broadcastChannel from AvailableChannels or from the BroadcastSecretary if none are available
		- set the broadcastChannel to the AuthorizationPacket.DesiredStreams
		- listen to the broadcastChannel and push anything relevant that flows in over the websocket

## Internal Workings
#### Broadcasting
In Go, channels are queues, which means as soon as an item is consumed, it is no longer in the queue. Here, we need a different paradigm: we want to have something more similar to a cable service, where there are public channels which are published by a single company and subscribed to by many listeners (with some customization).

To do this, we create a set of AvailableChannels. Whenever a user disconnects, they place their channel into the AvailableChannels for someone else to use; conversely, whenever a user connects, they attempt to get a channel to use from the AvailableChannels. If none are available, this means all current channels are in use; in this case, a new channel is requested from the BroadcastSecretary (who will also cleanup excess channels in AvailableChannels after periods of heavy traffic). Once a user has acquired a channel, it is essentially synonymous with their websocket connection, and they will be subscribed to all broadcasts which meet their requested subscription configuration criteria.

#### Twitter
The Twitter streaming api is used to provide tweets, which are filtered as follows:

1. Parse away retweets into the original source
2. Discard anything we have already sent recently
3. Filter out tweets from users with low follower counts
4. Check for keyword matches for available streams

#### RSS
Many news sources provide RSS feeds, which are also monitored. Articles are not categorized by keywords, but by how the original news source classifies that particular RSS feed (e.g., a world news feed, a technology feed, etc.)
