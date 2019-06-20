/**
 * @file main.go
 * @author Mikhail Klementyev jollheef<AT>riseup.net
 * @license GNU GPLv3
 * @date July, 2016
 * @brief Tiny non-interactive cli browser
 */

package main

import (
	"context"
	"net/http"
	"os"
	"strings"

	"code.dumpstack.io/tools/wi/commands"
	"code.dumpstack.io/tools/wi/storage"

	cookiejar "github.com/juju/persistent-cookiejar"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/cretz/bine/process"
	"github.com/cretz/bine/tor"
	"github.com/ipsn/go-libtor"
)

var creator = libtor.Creator

type LibTorWrapper struct{}

func (LibTorWrapper) New(ctx context.Context, args ...string) (process.Process, error) {
	return creator.New(ctx, args...)
}

type searchList []string

func (l *searchList) Set(value string) (err error) {
	*l = append(*l, value)
	return
}

func (l *searchList) String() (s string) {
	return ""
}

func (l *searchList) IsCumulative() bool {
	return true
}

func SearchList(settings kingpin.Settings) (target *[]string) {
	target = new([]string)
	settings.SetValue((*searchList)(target))
	return
}

var (
	useTor = kingpin.Flag("tor", "Use embedded tor").Default("false").Bool()
	ua     = kingpin.Flag("ua", "User-Agent").Default("Wi 0.1").String()

	get    = kingpin.Command("get", "Get url")
	getUrl = get.Arg("url", "Url").Required().String()

	form     = kingpin.Command("form", "Fill form")
	formID   = form.Arg("id", "Form ID").Required().Int64()
	formArgs = SearchList(form.Arg("args", "Form arguments"))

	link            = kingpin.Command("link", "Get link")
	linkNo          = link.Arg("no", "Number").Required().Int64()
	linkFromHistory = link.Flag("history", "Item from history").Bool()

	historyList      = kingpin.Command("history", "List history")
	historyListItems = historyList.Arg("items", "Amount of items").Int64()
	historyListAll   = historyList.Flag("all", "Show all items").Bool()

	search     = kingpin.Command("search", "Search by duckduckgo")
	searchArgs = SearchList(search.Arg("string", "String for search"))
)

func main() {
	homePath, exists := os.LookupEnv("HOME")
	var wiDir, widbPath, wijarPath string
	if exists {
		wiDir = homePath + "/.wi"
		widbPath = wiDir + "/wi.db"
		wijarPath = wiDir + "/wi.jar"
	} else {
		wiDir = "/tmp"
		widbPath = "/tmp/wi.db"
		wijarPath = "/tmp/wi.jar"
	}

	err := os.MkdirAll(wiDir, 0700)
	if err != nil {
		panic(err)
	}

	db, err := storage.OpenDB(widbPath)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	os.Setenv("GOCOOKIES", wijarPath)

	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}

	defer jar.Save()

	kingpin.Parse()

	var t *tor.Tor
	if *useTor {
		t, err = tor.Start(nil, &tor.StartConf{
			ProcessCreator: LibTorWrapper{},
			DataDir:        wiDir + "/tor",
		})
		if err != nil {
			panic(err)
		}

		defer t.Close()

		dialer, err := t.Dialer(nil, nil)
		if err != nil {
			panic(err)
		}

		commands.Transport = &http.Transport{DialContext: dialer.DialContext}
	}

	commands.UserAgent = *ua

	switch kingpin.Parse() {
	case "get":
		commands.Get(db, jar, *getUrl)
	case "form":
		commands.Form(db, jar, *formID, *formArgs)
	case "link":
		commands.Link(db, jar, *linkNo, *linkFromHistory)
	case "history":
		commands.History(db, *historyListItems, 20, *historyListAll)
	case "search":
		commands.Get(db, jar, "https://duckduckgo.com/html/?kd=-1&q="+strings.Join(*searchArgs, "+"))
	}
}
