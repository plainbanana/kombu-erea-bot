# How to run
Regist binary to crontab.

# Usage 
- get new ClientID and ClientSecret
`% env MASTODONSERVER=XXX MASTODONAPPWEBSITE=XXX go run main.go`
- and then run
`% env USERAGENT=XXX MASTODONSERVER=XXX MASTODONCLIENTID=XXX MASTODONCLIENTSECRET=XXX MASTODONAPPYOUREMAIL=XXX MASTODONAPPYOURPASSWORD=XXX go run main.go`
# Notice
This Bot uses [Spla2API](https://spla2.yuu26.com/). You need to set UserAgent which contains program name and your contacts info (default is this repo).