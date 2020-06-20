[![build](https://img.shields.io/github/workflow/status/plainbanana/kombu-erea-bot/release)](https://github.com/plainbanana/kombu-erea-bot/actions?query=workflow%3Arelease) [![release](https://img.shields.io/github/v/release/plainbanana/kombu-erea-bot)](https://github.com/plainbanana/kombu-erea-bot/releases/latest)

# kombu-area-bot
A simple bot for mastodon. This bot toot to followers when gachi splat_zones will hold at コンブトラック(kombu).
# Usage 
- get new ClientID and ClientSecret

    `% env MASTODONSERVER=XXX MASTODONAPPWEBSITE=XXX go run main.go`
- and then run

    `% env USERAGENT=XXX MASTODONSERVER=XXX MASTODONCLIENTID=XXX MASTODONCLIENTSECRET=XXX MASTODONAPPYOUREMAIL=XXX MASTODONAPPYOURPASSWORD=XXX go run main.go`
- Regist binary to crontab
# Notice
This Bot uses [Spla2API](https://spla2.yuu26.com/). You need to set UserAgent which contains program name and your contacts info (default is this repo).
