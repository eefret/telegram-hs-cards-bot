package hsimagebot

import (
	"appengine"
	"appengine/urlfetch"
	botApi "github.com/Syfaro/telegram-bot-api"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

//----------------------------------------------Initialization
const (
	TOKEN string = "TOKEN"
	BASE_URL      string = "https://api.telegram.org/bot" + TOKEN + "/"
	HS_API_SEARCH string = "https://omgvamp-hearthstone-v1.p.mashape.com/cards/search/"
	MASHAPE_KEY   string = "tntkXJyM7EmshBgQYsXtCHHEX8Izp1uHrN1jsnTpw7tNCxEZIN"
)


func init() {
	http.HandleFunc("/me", meHandler)
	http.HandleFunc("/update", updateHandler)
	http.HandleFunc("/set_webhook", setWebhookHandler)
	http.HandleFunc("/webhook", webhookHandler)
}

//----------------------------------------------Handlers
func meHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	bot, err := botApi.NewBotAPIWithClient(TOKEN, urlfetch.Client(appengine.NewContext(r)))
	user, err := bot.GetMe()
	resp, err := json.Marshal(user)
	checkErr(w, err)
	fmt.Fprint(w, string(resp))
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	bot, err := botApi.NewBotAPIWithClient(TOKEN, urlfetch.Client(appengine.NewContext(r)))
	num, err := strconv.Atoi(r.URL.Query().Get("offset"))
	updates, err := bot.GetUpdates(botApi.NewUpdate(num))
	resp, err := json.Marshal(updates)
	checkErr(w, err)
	fmt.Fprint(w, string(resp))
}

func setWebhookHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	bot, err := botApi.NewBotAPIWithClient(TOKEN, urlfetch.Client(appengine.NewContext(r)))
	resp, err := bot.SetWebhook(botApi.NewWebhook(r.URL.Query().Get("url")))
	str, err := json.Marshal(resp)
	checkErr(w, err)
	fmt.Fprint(w, string(str))
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	//Getting context
	c := appengine.NewContext(r)
	//decoding to json Message struct
	decoder := json.NewDecoder(r.Body)
	var update botApi.Update
	err := decoder.Decode(&update)
	if err != nil {
		c.Errorf("%v", err)
	}

	// getting the string to respond
	body, err := json.Marshal(update)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(body))
	checkErr(w, err)

	//getting msg
	msg := update.Message

	if isMessageRepeated(c, msg.MessageID, msg.Chat.ID) {
		c.Infof("repeated message")
		return
	}
	err = insertRespondedMessage(c, msg.MessageID, msg.Chat.ID)
	checkErr(w, err)

	active := isActiveBot(c, msg.Chat.ID)
	c.Infof("status is active: %v", active)
	//Working with commands
	if strings.HasPrefix(msg.Text, "/") {
		//Endpoints available when active
		if active {
			if strings.HasPrefix(msg.Text, "/hs") { //CARDS FETCHING PARAMS
				commandHs(w, r, msg)
			} else if strings.HasPrefix(msg.Text, "/stop") {
				err := insertStatus(c, msg.Chat.ID, false)
				if err != nil {
					c.Errorf("%v", err)
				}
			}
		} else { //Available when inactive
			if strings.HasPrefix(msg.Text, "/start") {
				err := insertStatus(c, msg.Chat.ID, true)
				if err != nil {
					c.Errorf("%v", err)
				}
			}
		}
	}
}

//----------------------------------------------Methods


