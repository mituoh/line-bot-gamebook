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
		addPushTask(c, pbd, e.Source.UserID)
	case linebot.EventTypeFollow:
		startAddress := "*start"
		addPushTask(c, startAddress, e.Source.UserID)
	}

	w.WriteHeader(200)
}

func addPushTask(c context.Context, a string, id string) {
	scripts, _ := lgscript.Load(a)
	i := 0
	for script := range scripts {
		i++
		lm := Msg{UserID: id, Script: scripts[script]}
		j, err := json.Marshal(lm)
		if err != nil {
			errorf(c, "json.Marshal: %v", err)
			return
		}
		d := base64.StdEncoding.EncodeToString(j)
		t := taskqueue.NewPOSTTask("/push", url.Values{"data": {d}})
		// delay, err := time.ParseDuration("2s")
		delay := time.Duration(i) * time.Duration(3) * time.Second
		t.Delay = delay
		taskqueue.Add(c, t, "")
	}
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

	ctx := appengine.NewContext(r)
	log.Debugf(ctx, string(j))

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

	if lm.Script.Action.Token == 0 {
		m := linebot.NewTextMessage(lm.Script.Text)
		_, err = bot.PushMessage(lm.UserID, m).WithContext(c).Do()
		if err != nil {
			errorf(nil, "PushMessage: %v", err)
			return
		}
	} else {
		br1 := lm.Script.Action.Branch1
		br2 := lm.Script.Action.Branch2
		btn1 := linebot.NewPostbackTemplateAction(br1.Text, br1.Address, br1.Text)
		btn2 := linebot.NewPostbackTemplateAction(br2.Text, br2.Address, br2.Text)
		btns := linebot.NewButtonsTemplate("", "", lm.Script.Text, btn1, btn2)
		btnsMsg := linebot.NewTemplateMessage(lm.Script.Text, btns)
		_, err = bot.PushMessage(lm.UserID, btnsMsg).WithContext(c).Do()
		if err != nil {
			errorf(nil, "PushMessage: %v", err)
			return
		}
	}

	w.WriteHeader(200)
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
