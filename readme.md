# Ambian Monitor
Backend data monitoring service used by Ambian
## Flow
1. A new connection C is received and handled by websocketConnectionHandler
2. C must provide an AuthorizationPacket immediately

    - if C doesn't or the metadata is invalid

	    - close C
	
	- if yes
		- acquire a new broadcastChannel from AvailableChannels or from the BroadcastSecretary if none are available
		- set the broadcastChannel to the AuthorizationPacket.DesiredStreams
		- listen to the broadcastChannel and push anything that flows in over the websocket