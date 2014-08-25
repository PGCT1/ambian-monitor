package tweet

import "strings"

type TweetMedia struct{
  Id int64
  Url string
  Type string
}

type Tweet struct{
  Id int64
  CreatedAt int64
  UserId int64
  Username string
  Screenname string
  Verified bool
  UserImageUrl string
  Followers int
  Text string
  Hashtags []string
  Media []TweetMedia
}

func (tweet Tweet) KeywordMatch(keywords []string) bool {

  separators := " .,!;:`~#$%^&*()_-+={}[]'/\\<>"

  found := false

  keywordSearch: for _,keyword := range keywords {

    for _,hashtag := range tweet.Hashtags {

      if keyword == hashtag {
        found = true
        break keywordSearch
      }

    }

    text := strings.ToLower(tweet.Text)

    if strings.Contains(text, keyword) {

      if strings.HasSuffix(text, keyword) {

          found = true
          break keywordSearch

      }else{

        if strings.HasPrefix(text, keyword) {

          i := strings.Index(text,keyword) + len(keyword)

          charAfter := text[i:i+1]

          if strings.Contains(separators,charAfter) {

            found = true
            break keywordSearch

          }

        }else{

          i := strings.Index(text,keyword) - 1
          j := strings.Index(text,keyword) + len(keyword)

          charBefore := text[i:i+1]
          charAfter := text[j:j+1]

          if strings.Contains(separators,charBefore) && strings.Contains(separators,charAfter) {

            found = true
            break keywordSearch

          }

        }

      }

    }

  }

  return found

}
