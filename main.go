package main

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/mattn/go-mastodon"
)

// config
var (
	userAgent          = "kombu-erea-bot (repo https://github.com/plainbanana/kombu-erea-bot)"
	mastodonServer     = "https://mustardon.tokyo"
	mastodonAppWebsite = "https://github.com/plainbanana/kombu-erea-bot"

	// following conf should be set by environments
	mastodonClientID        = ""
	mastodonClientSecret    = ""
	mastodonAppYourEmail    = ""
	mastodonAppYourPassword = ""

	timezone = time.FixedZone("Asia/Tokyo", 9*60*60)
)

const (
	splatoon2API = "https://spla2.yuu26.com"
	chacheFile   = "./api_chache.gob"
)

type splatoonRespSchedules struct {
	Result []struct {
		Rule   string `json:"rule"`
		RuleEx struct {
			Key     string `json:"key"`
			Name    string `json:"name"`
			Statink string `json:"statink"`
		} `json:"rule_ex"`
		Maps   []string `json:"maps"`
		MapsEx []struct {
			ID      int    `json:"id"`
			Name    string `json:"name"`
			Image   string `json:"image"`
			Statink string `json:"statink"`
		} `json:"maps_ex"`
		Start    string    `json:"start"`
		StartUtc time.Time `json:"start_utc"`
		StartT   int       `json:"start_t"`
		End      string    `json:"end"`
		EndUtc   time.Time `json:"end_utc"`
		EndT     int       `json:"end_t"`
		Tooted   struct {
			First   bool
			Secound bool
		}
	} `json:"result"`
	Timestamp time.Time
}

func init() {
	if s := os.Getenv("USERAGENT"); s != "" {
		userAgent = s
	}
	if s := os.Getenv("MASTODONSERVER"); s != "" {
		mastodonServer = s
	}
	if s := os.Getenv("MASTODONAPPWEBSITE"); s != "" {
		mastodonAppWebsite = s
	}

	mastodonClientID = os.Getenv("MASTODONCLIENTID")
	mastodonClientSecret = os.Getenv("MASTODONCLIENTSECRET")

	if mastodonClientID == "" || mastodonClientSecret == "" {
		app, err := mastodon.RegisterApp(context.Background(), &mastodon.AppConfig{
			Server:     mastodonServer,
			ClientName: "kombu-erea-bot",
			Scopes:     "read write follow",
			Website:    mastodonAppWebsite,
		})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("For example, you can use following environments\n")
		fmt.Printf("env MASTODONCLIENTID=%s MASTODONCLIENTSECRET=%s\n", app.ClientID, app.ClientSecret)
		log.Fatalln("OMG! luck of environments for your app: MASTODONCLIENTID or MASTODONCLIENTSECRET")
	}

	mastodonAppYourEmail = os.Getenv("MASTODONAPPYOUREMAIL")
	mastodonAppYourPassword = os.Getenv("MASTODONAPPYOURPASSWORD")

	if mastodonAppYourEmail == "" || mastodonAppYourPassword == "" {
		log.Fatalln("OMG! luck of environments for your app: MASTODONAPPYOUREMAIL or MASTODONAPPYOURPASSWORD")
	}
}

func main() {
	var statusText string

	schedules := getSplatoon2GachiSchedules("gachi/schedule")
	for i, v := range schedules.Result {
		if v.Rule == "ガチエリア" && isContain(v.Maps, "コンブトラック") {
			if time.Now().Add(time.Hour*2).After(v.StartUtc) && v.EndUtc.After(time.Now()) && !schedules.Result[i].Tooted.First {
				statusText = statusText + "コンブエリア start at " +
					v.StartUtc.In(timezone).Format("2006-01-02 15:03 -07:00") + " \n"
				toot(statusText)
				schedules.Result[i].Tooted.First = true
				continue
			}

			if time.Now().Add(time.Minute*10).After(v.StartUtc) && v.EndUtc.After(time.Now()) && !schedules.Result[i].Tooted.Secound {
				statusText = statusText + "コンブエリア soon start at " +
					v.StartUtc.In(timezone).Format("2006-01-02 15:03 -07:00") + " \n"
				toot(statusText)
				schedules.Result[i].Tooted.Secound = true
			}
		}
	}
	storeRespToFile(schedules)
}

func toot(text string) {
	c := mastodon.NewClient(&mastodon.Config{
		Server:       mastodonServer,
		ClientID:     mastodonClientID,
		ClientSecret: mastodonClientSecret,
	})
	err := c.Authenticate(context.Background(), mastodonAppYourEmail, mastodonAppYourPassword)
	if err != nil {
		log.Fatal(err)
	}

	curUser, err := c.GetAccountCurrentUser(context.Background())
	if !errors.Is(err, nil) {
		log.Fatal(err)
	}
	curFollowers, err := c.GetAccountFollowers(context.Background(), curUser.ID, nil)
	if !errors.Is(err, nil) {
		log.Fatal(err)
	}

	for _, v := range strings.Split(parseAccountsToMention(curFollowers), " ") {
		if v != "" {
			c.PostStatus(context.Background(), &mastodon.Toot{
				Status:     v + " " + text,
				Visibility: "unlisted",
			})
			log.Println("toot", v+" "+text)
		}
	}
}

func getSplatoon2GachiSchedules(uri string) splatoonRespSchedules {
	if oldResp := restoreRespFromFile(); oldResp.Timestamp.Add(time.Hour*12).After(time.Now()) && !oldResp.Timestamp.IsZero() {
		log.Println("return old response", oldResp.Timestamp, time.Now())
		return oldResp
	}
	log.Println("call API")
	resp := getFromSpla2API(uri)
	resp.Timestamp = time.Now()
	return resp
}

func getFromSpla2API(uri string) splatoonRespSchedules {
	base, err := url.Parse(splatoon2API)
	if !errors.Is(err, nil) {
		log.Fatal(err)
	}
	base.Path = path.Join(base.Path, uri)
	req, err := http.NewRequest("GET", base.String(), nil)
	if !errors.Is(err, nil) {
		log.Fatal(err)
	}

	req.Header.Set("User-Agent", userAgent)

	c := http.DefaultClient
	resp, err := c.Do(req)
	if !errors.Is(err, nil) {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if !errors.Is(err, nil) {
		log.Fatal(err)
	}
	var result splatoonRespSchedules
	err = json.Unmarshal(body, &result)
	if !errors.Is(err, nil) {
		log.Fatal(err)
	}

	return result
}

func storeRespToFile(b splatoonRespSchedules) {
	f, err := os.Create(chacheFile)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	enc := gob.NewEncoder(f)
	if err := enc.Encode(b); !errors.Is(err, nil) {
		log.Fatal(err)
	}
}

func restoreRespFromFile() splatoonRespSchedules {
	f, err := os.Open(chacheFile)
	if err != nil {
		return splatoonRespSchedules{}
	}
	defer f.Close()
	var resp splatoonRespSchedules
	dec := gob.NewDecoder(f)
	if err := dec.Decode(&resp); !errors.Is(err, nil) {
		log.Fatal(err)
	}
	return resp
}

func isContain(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func parseAccountsToMention(accounts []*mastodon.Account) string {
	var result string
	for _, v := range accounts {
		if v.Bot {
			continue
		}
		result += parseAccountToMention(v)
	}
	return result
}

func parseAccountToMention(account *mastodon.Account) string {
	s := strings.Split(account.URL, "/")
	return s[len(s)-1] + "@" + s[len(s)-2] + " "
}
