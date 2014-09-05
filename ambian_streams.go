package main

import "github.com/pgct1/ambian-monitor/tweet"

type AmbianStream struct {
  Id int
  Name string

  // source metadata

  TwitterKeywords []string
  NewsSources []NewsSource
  Filter func(tweet.Tweet, []string, bool)bool
}

func InitializeAmbianStreams() []AmbianStream {

  CreateAmbianStream(AmbianStream{
    Id:1,
    Name:"World News",
    TwitterKeywords:[]string{"syria","egypt","sanction","gdp","hamas","idf","palestine","gaza","putin","snowden","russia","benghazi","isis","merkel","kerry","clinton","brussels","moscow"},
    NewsSources:[]NewsSource{
      // careful! these names (keys) are displayed in the client, and also map to
      // icons included in the client for these sources
      {"New York Times","http://rss.nytimes.com/services/xml/rss/nyt/InternationalHome.xml"},
      {"BBC International","http://feeds.bbci.co.uk/news/rss.xml?edition=int"},
      {"Reuters","http://mf.feeds.reuters.com/reuters/UKTopNews"},
      {"Al Jazeera","http://www.aljazeera.com/Services/Rss/?PostingId=2007731105943979989"},
      {"Washington Post","http://feeds.washingtonpost.com/rss/rss_blogpost"},
      {"Vice","https://news.vice.com/rss"},
      {"The Guardian","http://www.theguardian.com/world/rss"},
      {"Russia Today","http://rt.com/rss/"},
      {"The Wall Street Journal","http://online.wsj.com/xml/rss/3_7085.xml"},
      {"The Huffington Post","http://www.huffingtonpost.com/feeds/verticals/world/index.xml"},
      {"The Independent","http://rss.feedsportal.com/c/266/f/3503/index.rss"},
      {"The Telegraph","http://www.telegraph.co.uk/news/worldnews/rss"},
    },
    Filter:func(t tweet.Tweet, keywords []string, isCorporateSource bool) bool {

      // make sure this isn't irrelevant shit like cam girl ads or something

      sexyKeywords := []string{"webcam","girls","ass","fuck","bitch","boob","boobs","tit","tits","luv","hawt","sex","camgirls","horny","cam","kinky"}

      if !isCorporateSource {  // trust major news sources

        match := t.KeywordMatch(sexyKeywords)

        if match == true {
          return false
        }

      }

      // other irrelevant stuff that appears a lot

      irrelevant := []string{"concert","in-concert","1d","touchdown","football","home run","kardashian","kardashians","bieber","belieber","homerun","emmy","emmys","emmys2014"}

      if !isCorporateSource {  // trust major news sources

        match := t.KeywordMatch(irrelevant)

        if match == true {
          return false
        }

      }

      // can't detect anything in our blacklists, so if we have a keyword
      // match, let it through

      return t.KeywordMatch(keywords)

    },
  })

  CreateAmbianStream(AmbianStream{
    Id:2,
    Name:"Social & Entertainment",
    TwitterKeywords:[]string{"harhar"},
    NewsSources:[]NewsSource{},
    Filter:func(t tweet.Tweet, keywords []string, isCorporateSource bool) bool {
      return true
    },
  })

  ambianStreams,_ = GetAmbianStreams()

  return ambianStreams
}
