'use strict';

// modules

var twitterStream = require('./SocialModules/twitter.js');

//constants

var startupDelay = 5000;	// must delay longer than sails lift; hack-ish, but sails has no onFinishedLoading hook...

// streams

setTimeout(function(){

	twitterStream.start();

},startupDelay);

// pings
