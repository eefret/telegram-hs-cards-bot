package hsimagebot

import (
	"appengine"
	"appengine/datastore"
	"appengine/urlfetch"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

//----------------------------------------------Initialization
const (
	TOKEN         string = "TOKEN"
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
	getAndPrint(w, r, BASE_URL+"getMe")
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	offset := r.URL.Query().Get("offset")
	getAndPrint(w, r, BASE_URL+"getUpdates?offset="+offset)
}

func setWebhookHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	getAndPrint(w, r, BASE_URL+"setWebhook?url="+url)
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	//Getting context
	c := appengine.NewContext(r)
	//decoding to json Message struct
	decoder := json.NewDecoder(r.Body)
	var update Update
	err := decoder.Decode(&update)
	if err != nil {
		c.Infof("%v", err)
	}

	// getting the string to respond
	body, _ := json.Marshal(update)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(body))

	//getting msg
	msg := update.Message
	c.Infof("%v", string(body))
	q := datastore.NewQuery(
		"RespondedMessage").Filter(
		"MessageId =", msg.MessageId).Filter(
		"ChatId =", msg.Chat.Id)
	count, _ := q.Count(c)
	if count > 0 {
		c.Infof("respondedmessage found %v with chatId %v and messageId %v ",
			count, msg.Chat.Id, msg.MessageId)
		return
	}
	//c.Infof("webhook body received: %v", string(body))
	respMessage := new(RespondedMessage)
	respMessage.MessageId = msg.MessageId
	respMessage.Success = true
	respMessage.ChatId = msg.Chat.Id

	if len(msg.Text) > 0 {
		if strings.HasPrefix(msg.Text, "/") {
			if strings.HasPrefix(msg.Text, "/hs") {
				if regexp.MustCompile("{").MatchString(msg.Text) {
					//extracting parameters
					var parameters []string
					result := regexp.MustCompile("\\{(.*?)\\}").FindAllStringSubmatch(msg.Text, -1)
					for _, v := range result {
						c.Infof("param: %v", v[1])
						parameters = append(parameters, v[1])
					}
					//extracting cardname
					rg := regexp.MustCompile("\\/hs(.*?)\\{")
					cardName := strings.TrimSpace(rg.FindStringSubmatch(msg.Text)[1])
					c.Infof("card name: %v", cardName)
					if len(cardName) == 0 {
						_, err := reply(r, msg.Chat.Id, msg.MessageId, "must provide cardname", "")
						checkErr(w, err)
						c.Infof("no cardname provided")
					}

					if len(parameters) > 2 {
						_, err := reply(r, msg.Chat.Id, msg.MessageId, "Too many params max 2", "")
						checkErr(w, err)
						c.Infof("too many params")
					} else {
						cards := getCard(r, cardName)
						c.Infof("card retrieved: %v", len(cards))
						cardIndex, isGold := extractParams(parameters)
						if cardIndex != -1 && len(cards) > 1 && (cardIndex <= len(cards)-1) {
							if isGold {
								_, err := reply(r,
									msg.Chat.Id,
									msg.MessageId,
									cards[cardIndex].ImgGold, "")
								checkErr(w, err)
								c.Infof("replied golden card at index %v", cardIndex)
							} else {
								_, err := reply(r,
									msg.Chat.Id,
									msg.MessageId,
									cards[cardIndex].Img, "")
								checkErr(w, err)
								c.Infof("replied card at index %v", cardIndex)
							}
						} else if cardIndex > (len(cards) - 1) {
							_, err := reply(r,
								msg.Chat.Id,
								msg.MessageId,
								fmt.Sprintf("index out of bounds, index %v in max %v , try again.",
									cardIndex, len(cards)-1), "")
							checkErr(w, err)
							c.Infof("index out of bounds")
						} else if cardIndex == -1 && len(cards) > 1 {
							_, err := reply(r,
								msg.Chat.Id,
								msg.MessageId,
								createMulticardError(cards, cardName), "")
							checkErr(w, err)
							c.Infof("multiple cards without index")
						} else if cardIndex == -1 && len(cards) == 1 {
							if isGold {
								_, err := reply(r,
									msg.Chat.Id,
									msg.MessageId,
									cards[0].ImgGold, "")
								checkErr(w, err)
								c.Infof("replied golden card")
							} else {
								_, err := reply(r,
									msg.Chat.Id,
									msg.MessageId,
									cards[0].Img, "")
								checkErr(w, err)
								c.Infof("replied card ")
							}
						} else if len(cards) == 0 {
							_, err := reply(r, msg.Chat.Id, msg.MessageId, "couldn't found that card sorry", "")
							checkErr(w, err)
							c.Infof("no cardname found")
						}
					}
				} else {
					cardName := strings.TrimSpace(msg.Text[3:])
					cards := getCard(r, cardName)

					if len(cards) > 1 {
						_, err := reply(r, msg.Chat.Id, msg.MessageId,
							createMulticardError(cards, cardName), "")
						checkErr(w, err)
						c.Infof("multiple cards with no params")
					} else if len(cards) == 1 {
						_, err := reply(r, msg.Chat.Id, msg.MessageId, cards[0].Img, "")
						checkErr(w, err)
						c.Infof("replied first image with no params")
					} else {
						_, err := reply(r, msg.Chat.Id, msg.MessageId, "couldn't found that card sorry", "")
						checkErr(w, err)
						c.Infof("no cardname found")
						return
					}
				}

			} else if strings.HasPrefix(msg.Text, "/hs") && len(strings.TrimSpace(msg.Text)) <= 3 {
				_, err := reply(r, msg.Chat.Id, msg.MessageId, "must provide card name", "")
				checkErr(w, err)
				c.Infof("please provide cardname")
			} else {
				_, err := reply(r, msg.Chat.Id, msg.MessageId, "What command ?", "")
				checkErr(w, err)
				c.Infof("What Command?")
			}
		} else if msg.Text == "who are you" {
			_, err := reply(r, msg.Chat.Id, msg.MessageId,
				"heartstone image bot by: Christopher T. Herrera(eefretsoul@gmail.com)", "")
			checkErr(w, err)
			c.Infof("replied who am I")
		}
	} else {
		reply(r, msg.Chat.Id, msg.MessageId, "What?", "")
		c.Infof("What?")
	}

	//Saving message cache to datastore
	key, err := datastore.Put(c, datastore.NewIncompleteKey(c, "RespondedMessage", nil), respMessage)
	c.Infof("responded message key: %v", key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

//----------------------------------------------Methods

func extractParams(parameters []string) (cardIndex int, isGold bool) {
	cardIndex = -1
	isGold = false
	for _, v := range parameters {
		if num, err := strconv.Atoi(v); err == nil {
			cardIndex = num
		}
		if strings.TrimSpace(strings.ToLower(v)) == "gold" {
			isGold = true
		}
	}
	return cardIndex, isGold
}

func createMulticardError(cards []Card, cardName string) string {
	response := "Your query returned several cards: \n"
	for k, v := range cards {
		response = response + "- " + v.Name + " (" + strconv.Itoa(k) + ") \n"
	}
	response = response + "Select which one you want with '/hs " + cardName + " {digit} and"
	response = response + " redo the same query or type a more specific"
	response = response + " query with the exact name in the list."
	return response
}

//reply sends a reply to the given chat quoting the given messageId
func reply(r *http.Request, chatId int64, messageId int64, message string, imgURL string) (resp *http.Response, err error) {
	c := appengine.NewContext(r)
	client := urlfetch.Client(c)
	if len(message) > 0 {
		parameters := url.Values{}
		parameters.Add("chat_id", strconv.FormatInt(chatId, 10))
		parameters.Add("text", message)
		parameters.Add("disable_web_page_preview", "false")
		parameters.Add("reply_to_message_id", strconv.FormatInt(messageId, 10))
		r, err := http.NewRequest("POST",
			BASE_URL+"sendMessage",
			strings.NewReader(parameters.Encode()))
		resp, err := client.Do(r)
		defer resp.Body.Close()
		return resp, err
	} else if len(imgURL) > 0 {
		//Get the image
		r, err := client.Get(imgURL)
		defer r.Body.Close()
		//create body and write
		reqBody := &bytes.Buffer{}
		writer := multipart.NewWriter(reqBody)
		part, err := writer.CreateFormFile("photo", "card.png")
		img, err := ioutil.ReadAll(r.Body)
		part.Write(img)
		writer.WriteField("chat_id", strconv.FormatInt(chatId, 10))
		writer.WriteField("reply_to_message_id", strconv.FormatInt(messageId, 10))
		//Close and send
		err = writer.Close()
		req, err := http.NewRequest("POST", BASE_URL+"sendPhoto", reqBody)
		resp, err := client.Do(req)
		defer resp.Body.Close()
		c.Infof("%v", resp.Status)
		return resp, err
	}
	return resp, err
}

func getCard(r *http.Request, name string) []Card {
	c := appengine.NewContext(r)
	client := urlfetch.Client(c)
	//parsing url
	URL, _ := url.Parse(HS_API_SEARCH + name)
	params := url.Values{}
	params.Add("collectible", "1")
	URL.RawQuery = params.Encode()

	req, err := http.NewRequest("GET",
		URL.String(),
		nil)
	c.Infof(HS_API_SEARCH + url.QueryEscape(name) + "?collectible=1)")
	req.Header.Set("X-Mashape-Key", MASHAPE_KEY)
	resp, err := client.Do(req)
	if err != nil {
		c.Infof("error during cards request")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		var cardResponse []Card
		return cardResponse
	}
	//body, _ := ioutil.ReadAll(resp.Body)
	//c.Infof("%v", string(body))
	decoder := json.NewDecoder(resp.Body)
	var cardResponse []Card
	err = decoder.Decode(&cardResponse)
	if err != nil {
		c.Infof("error decoding response %v", err)
	}
	return cardResponse
}

//getAndPrint executes getFromUrl and prints it into the ResponseWriter
func getAndPrint(w http.ResponseWriter, r *http.Request, url string) {
	fmt.Fprint(w, string(getFromUrl(w, r, url)))
}

//getFromUrl takes an url and perform a GET request to it then give backthe body
func getFromUrl(w http.ResponseWriter, r *http.Request, url string) []byte {
	client := urlfetch.Client(appengine.NewContext(r))
	resp, err := client.Get(url)
	checkErr(w, err)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	checkErr(w, err)
	return body
}

//checkErr takes a ResponseWriter and and Error if there is an error
//then it Writes it to the ResponseWriter and return a http.StatusInternalServerError
func checkErr(w http.ResponseWriter, err error) {
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
