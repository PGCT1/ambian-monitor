'use strict';

var Twitter = require('ntwitter');

var twitter = new Twitter({
  consumer_key: 'XyRF2uw8qnmO5SATkeZJDaJux',
  consumer_secret: '83SP95RAiq6QO4etRiPncQF1ZeTIpo9yippYx72DCsRp53BRVs',
  access_token_key: '2601992358-qAwvD43qmmBc8sJTaw5dsP4bSecCkINsRy5KDdN',
  access_token_secret: 'rOAUCdUfScb0hYkaJ6yyaZoZeJquFiB2FvnCsm7NHOnOY'
});

var capture = {
	stream:false,	// will be set to the active twitter stream as soon as we have one

	// these will also be set later; placeholders are given here, to show the structure

	textTerms:{
		'exampleTerm':[1,4]	// example term has been registered by feeds 1 and 4
	},
	domains:{},
	hashtags:{}
};

module.exports.start = function(){

	buildTermList(function(err,terms){

		if(err){
			return console.log(err);
		}

		console.log('Starting twitter monitor with terms:');
		console.log(terms);

		OpenTwitterStreamWithTerms(terms);

	});

};

module.exports.stop = function(){
	capture.stream.destroy();
};

module.exports.refreshTerms = function(){

	capture.stream.destroy();

	buildTermList(function(err,terms){

		if(err){
			return console.log(err);
		}

		console.log('Refreshing stream with new terms:');
		console.log(terms);

		OpenTwitterStreamWithTerms(terms);

	});

}

function handleNewTweet(incomingTweet){

	// assert validity

	if(!incomingTweet || !incomingTweet.user){
		return;
	}

	// parse this into a simpler object

	var tweet = {
		id:incomingTweet.id,
		text:incomingTweet.text,

		user:{
			id:incomingTweet.user.id,
			name:incomingTweet.user.name,
			image:incomingTweet.user.profile_image_url,
			followers:incomingTweet.user.followers_count
		}
	};

	if(incomingTweet.entities && incomingTweet.entities.hashtags){
		
		tweet.hashtags = [];

		for(var i=0;i<incomingTweet.entities.hashtags.length;++i){
			tweet.hashtags.push(incomingTweet.entities.hashtags[i].text);
		}

	}

	// decide which feeds this should be published to

	var feedsToPublish = [];

	// check for text matches

	var text = tweet.text.toLowerCase();

	var textTerms = Object.keys(capture.textTerms);

	for(var i=0;i<textTerms.length;++i){

		if(text.indexOf(textTerms[i]) != -1){

			// found a text term match, so add these feeds to our publish list

			var matchedFeedIds = capture.textTerms[textTerms[i]];

			for(var j=0;j<matchedFeedIds.length;++j){

				if(feedsToPublish.indexOf(matchedFeedIds[j]) == -1){
					feedsToPublish.push(matchedFeedIds[j]);
				}

			}

		}

	}

	// check for hashtag matches

	var tweetHashtags = [];

	for(var i=0;i<tweet.hashtags.length;++i){
		tweetHashtags.push(tweet.hashtags[i].toLowerCase());
	}

	var hashtags = Object.keys(capture.hashtags);

	for(var i=0;i<hashtags.length;++i){

		if(tweetHashtags.indexOf(hashtags[i]) != -1){

			// found a hashtag match, so add these feeds to our publish list

			var matchedFeedIds = capture.hashtags[hashtags[i]];

			for(var j=0;j<matchedFeedIds.length;++j){

				if(feedsToPublish.indexOf(matchedFeedIds[j]) == -1){
					feedsToPublish.push(matchedFeedIds[j]);
				}

			}

		}

	}

	// publish

	for(var i=0;i<feedsToPublish.length;++i){

		Feeds.publishUpdate(feedsToPublish[i],tweet);

	}

}

function buildTermList(callback){

	Feed.find({}).exec(function(err,feeds){

		if(err){
			return callback(err);
		}

		capture.textTerms = {};
		capture.domains = {};
		capture.hashtags = {};

		var terms = new Array();

		for(var i=0;i<feeds.length;++i){

			var feed = feeds[i];

			// text terms

			for(var j=0;j<feed.twitterTextTerms.length;++j){

				var term = feed.twitterTextTerms[j].toLowerCase();

				if(capture.textTerms[term]){
					capture.textTerms[term].push(feed.id);
				}else{
					terms.push(term);
					capture.textTerms[term] = [feed.id];
				}

			}

			// domains

			for(var j=0;j<feed.twitterDomains.length;++j){

				if(capture.domains[feed.twitterDomains[j]]){
					capture.domains[feed.twitterDomains[j]].push(feed.id);
				}else{
					capture.domains[feed.twitterDomains[j]] = [feed.id];
				}

			}

			// hash tags

			for(var j=0;j<feed.twitterHashTags.length;++j){

				var term = feed.twitterHashTags[j].toLowerCase();

				if(capture.hashtags[term]){
					capture.hashtags[term].push(feed.id);
				}else{
					terms.push(term);
					capture.hashtags[term] = [feed.id];
				}

			}

			callback(null,terms);

		}

	});

}

function OpenTwitterStreamWithTerms(terms){

	console.log('Opening twitter stream...');

	twitter.stream('user',{track:terms},function(stream){

		capture.stream = stream;

		stream.on('data', function(data){
			handleNewTweet(data);
		});

		stream.on('end', function (response){

			capture.stream = false;

			console.log('Twitter stream ended. Response:');
			console.log(response);
		});

		stream.on('destroy',function(response){

			capture.stream = false;

			console.log('Twitter stream destroyed. Response:');
			console.log(response);
		});

	});

}
