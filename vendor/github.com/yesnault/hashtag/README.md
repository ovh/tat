# hashtag
Package hashtag implements extraction of Twitter type hashtags, mentions and
replies form text in Go. This package partially ports extraction routines from
[Twitter's official Java package](https://github.com/twitter/twitter-text) 
to Go and runs most of the standard twitter-text conformance tests. It does not
yet implement URL extraction (and hence URL/Hashtag overlaps), cashtags and lists

Since the package attempts to closely follow the Twitter-Text Java API, function 
names may be longer than typical Go package function names

Installation
------------
Note: As the package matures, I plan to move this to gopkg.in
```
go get github.com/srinathh/hashtag
```

Usage
-----
Import the package as
```
import "github.com/srinathh/hashtag"
```

This package supports the following functions to extract hashtags and mentions 
with or without position markers. The functions omit # and @ characters 
(and also their higher unicode number counterparts  ＠ and ＃) from return values
- `ExtractHashtags(text string) []string`
- `ExtractHashtagsWithIndices(text string) []Entity`
- `ExtractMentions(text string) []string`
- `ExtractMentionsWithIndices(text string) []Entity`
- `ExtractReply(text string) string`

Documentation
-------------
Read the full documentation and examples on [GoDoc](http://godoc.org/github.com/srinathh/hashtag)



