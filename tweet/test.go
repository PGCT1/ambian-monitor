package tweet

func Test(){

  checkTweetKeywords()

}

func checkTweetKeywords(){

  keywords := []string{"keyword"}

  tweet := Tweet{
    Hashtags:[]string{"foo","keyword","bar"},
    Text:"fjakwe/f.,./f ,asdfajwelkaj;l fka",
  }

  if !tweet.KeywordMatch(keywords) {
    panic("Unexpected behavior: keyword not found")
  }

  tweet = Tweet{
    Hashtags:[]string{"foo","keyword1","bar"},
    Text:"fjakwe/f.,./f ,asdfajwelkaj;l fka",
  }

  if tweet.KeywordMatch(keywords) {
    panic("Unexpected behavior: keyword incorrectly found")
  }

  tweet = Tweet{
    Hashtags:[]string{"foo","bar"},
    Text:"fjakwe/f.,./keywordf ,asdfajwelkaj;l fka",
  }

  if tweet.KeywordMatch(keywords) {
    panic("Unexpected behavior: keyword incorrectly found")
  }

  tweet = Tweet{
    Hashtags:[]string{"foo","bar"},
    Text:"fjakwe/f.,./ keywordf ,asdfajwelkaj;l fka",
  }

  if tweet.KeywordMatch(keywords) {
    panic("Unexpected behavior: keyword incorrectly found")
  }

  tweet = Tweet{
    Hashtags:[]string{"foo","bar"},
    Text:"fjakwe/f.,./akeyword f ,asdfajwelkaj;l fka",
  }

  if tweet.KeywordMatch(keywords) {
    panic("Unexpected behavior: keyword incorrectly found")
  }

  tweet = Tweet{
    Hashtags:[]string{"foo","bar"},
    Text:"fjakwe/f.,./keyword f ,asdfajwelkaj;l fka",
  }

  if !tweet.KeywordMatch(keywords) {
    panic("Unexpected behavior: keyword not found")
  }

  tweet = Tweet{
    Hashtags:[]string{"foo","bar"},
    Text:"keyword f ,asdfajwelkaj;l fka",
  }

  if !tweet.KeywordMatch(keywords) {
    panic("Unexpected behavior: keyword not found")
  }

  tweet = Tweet{
    Hashtags:[]string{"foo","bar"},
    Text:"f ,asdfajwelkaj;l fka #keyword",
  }

  if !tweet.KeywordMatch(keywords) {
    panic("Unexpected behavior: keyword not found")
  }
}
