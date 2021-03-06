package main

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/taskqueue"
	"google.golang.org/appengine/urlfetch"

	"github.com/mjibson/goon"

	"golang.org/x/net/context"

	"github.com/joho/godotenv"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/line/line-bot-sdk-go/linebot/httphandler"

	"github.com/mituoh/line-bot-gamebook/lgscript"
)

// Msg represents a push message.
type Msg struct {
	UserID string          `json:"userId"`
	Script lgscript.Script `json:"script"`
}

// User is datastore object
type User struct {
	ID       string `datastore:"-" goon:"id"`
	Tappable []string
}

var botHandler *httphandler.WebhookHandler

func init() {
	err := godotenv.Load("line.env")
	if err != nil {
		panic(err)
	}

	botHandler, err = httphandler.New(
		os.Getenv("LINE_BOT_CHANNEL_SECRET"),
		os.Getenv("LINE_BOT_CHANNEL_TOKEN"),
	)
	botHandler.HandleEvents(handleCallback)

	http.Handle("/callback", botHandler)
	http.HandleFunc("/task", handleTask)
	http.HandleFunc("/push", handlePush)
}

// handleCallback is Webgook endpoint
func handleCallback(evs []*linebot.Event, r *http.Request) {
	c := newContext(r)
	ts := make([]*taskqueue.Task, len(evs))
	for i, e := range evs {
		j, err := json.Marshal(e)
		if err != nil {
			errorf(c, "json.Marshal: %v", err)
			return
		}
		data := base64.StdEncoding.EncodeToString(j)
		t := taskqueue.NewPOSTTask("/task", url.Values{"data": {data}})
		ts[i] = t
	}
	taskqueue.AddMulti(c, ts, "")
}

// handleTask is process event handler
func handleTask(w http.ResponseWriter, r *http.Request) {
	c := newContext(r)
	data := r.FormValue("data")
	if data == "" {
		errorf(c, "No data")
		return
	}

	j, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		errorf(c, "base64 DecodeString: %v", err)
		return
	}

	e := new(linebot.Event)
	err = json.Unmarshal(j, e)
	if err != nil {
		errorf(c, "json.Unmarshal: %v", err)
		return
	}

	logf(c, "EventType: %s\nMessage: %#v", e.Type, e.Message)

	switch e.Type {
	case linebot.EventTypeMessage:
		switch message := e.Message.(type) {
		case *linebot.TextMessage:
			if message.Text != "はじめる" {
				w.WriteHeader(200)
				return
			}
			startAddress := "*start"
			addPushTask(c, startAddress, e.Source.UserID)
		}
	case linebot.EventTypePostback:
		pbd := e.Postback.Data
		uid := e.Source.UserID
		// Get user from Datastore
		u := &User{ID: uid}
		getUser(r, u)
		for _, t := range u.Tappable {
			if t == pbd {
				addPushTask(c, pbd, uid)
				// Clear user to Datastore
				u = &User{ID: uid, Tappable: []string{}}
				putUser(r, u)
				return
			}
		}
	case linebot.EventTypeFollow:
		startAddress := "*start"
		addPushTask(c, startAddress, e.Source.UserID)
	}

	w.WriteHeader(200)
}

func addPushTask(c context.Context, a string, id string) {
	scripts, _ := lgscript.Load(a)
	oldTm := time.Duration(0)
	for _, script := range scripts {
		log.Debugf(c, script.Text)
		if script.Action.Token == lgscript.WAIT {
			delay, err := time.ParseDuration(script.Action.Wait)
			if err != nil {
				errorf(c, "time.ParseDuration: %v", err)
				return
			}
			oldTm = delay + oldTm
			continue
		}
		lm := Msg{UserID: id, Script: script}
		j, err := json.Marshal(lm)
		if err != nil {
			errorf(c, "json.Marshal: %v", err)
			return
		}
		d := base64.StdEncoding.EncodeToString(j)
		t := taskqueue.NewPOSTTask("/push", url.Values{"data": {d}})
		// Time cal
		l := len([]rune(lm.Script.Text))
		tm := delayTime(l)
		delay := time.Duration(tm)*time.Millisecond + oldTm
		oldTm = delay
		t.Delay = delay
		// Add task queue
		_, err = taskqueue.Add(c, t, "")
		if err != nil {
			errorf(c, "taskqueue.Add: %v", err)
			return
		}
	}
}

func delayTime(l int) (tm int) {
	tm = 2000
	if l < 5 {
		tm = 2000
	} else if l < 10 {
		tm = 2500
	} else if l < 15 {
		tm = 3000
	} else {
		tm = 4000
	}
	return tm
}

// handlePush is process event handler
func handlePush(w http.ResponseWriter, r *http.Request) {
	c := newContext(r)

	data := r.FormValue("data")
	if data == "" {
		errorf(c, "No data")
		return
	}

	j, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		errorf(c, "base64 DecodeString: %v", err)
		return
	}

	lm := new(Msg)
	err = json.Unmarshal(j, lm)
	if err != nil {
		errorf(c, "json.Unmarshal: %v", err)
		return
	}

	bot, err := newLINEBot(c)
	if err != nil {
		errorf(c, "newLINEBot: %v", err)
		return
	}

	switch lm.Script.Action.Token {
	case lgscript.TEXT:
		m := linebot.NewTextMessage(lm.Script.Text)
		_, err = bot.PushMessage(lm.UserID, m).WithContext(c).Do()
		if err != nil {
			errorf(nil, "PushMessage: %v", err)
			return
		}
	case lgscript.BUTTONS:
		btns := linebot.NewButtonsTemplate("", "", lm.Script.Text)
		var t []string // tappable data
		for _, br := range lm.Script.Action.Branches {
			if br.Text == "" {
				break
			}
			btn := linebot.NewPostbackTemplateAction(br.Text, br.Address, br.Text)
			btns.Actions = append(btns.Actions, btn)
			t = append(t, br.Address)
		}
		btnsMsg := linebot.NewTemplateMessage(lm.Script.Text, btns)
		_, err = bot.PushMessage(lm.UserID, btnsMsg).WithContext(c).Do()
		if err != nil {
			errorf(nil, "PushMessage: %v", err)
			return
		}
		// Put user to Datastore
		u := &User{ID: lm.UserID, Tappable: t}
		putUser(r, u)
	}

	w.WriteHeader(200)
}

func putUser(r *http.Request, u *User) {
	c := newContext(r)
	g := goon.NewGoon(r)
	_, err := g.Put(u)
	if err != nil {
		errorf(c, "Goon put: %v", err)
		return
	}
}

func getUser(r *http.Request, u *User) {
	c := newContext(r)
	g := goon.NewGoon(r)
	err := g.Get(u)
	if err != nil {
		errorf(c, "Goon get: %v", err)
		return
	}
}

func logf(c context.Context, format string, args ...interface{}) {
	log.Infof(c, format, args...)
}

func errorf(c context.Context, format string, args ...interface{}) {
	log.Errorf(c, format, args...)
}

func newContext(r *http.Request) context.Context {
	return appengine.NewContext(r)
}

func newLINEBot(c context.Context) (*linebot.Client, error) {
	return botHandler.NewClient(
		linebot.WithHTTPClient(urlfetch.Client(c)),
	)
}

func isDevServer() bool {
	return os.Getenv("RUN_WITH_DEVAPPSERVER") != ""
}
