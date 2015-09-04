package hsimagebot
import (
	"strconv"
	"appengine"
	botApi "github.com/Syfaro/telegram-bot-api"
	"github.com/technoweenie/multipartstreamer"
	"appengine/urlfetch"
	"net/http"
	"encoding/json"
	"appengine/datastore"
	"net/url"
)


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


func replyCard(messageID int, chatID int, card Card, c appengine.Context, isgold bool) (error) {
	ms := multipartstreamer.New()
	client := urlfetch.Client(c)
	var err error
	//Writing other params
	params := make(map[string]string)
	params["chat_id"] = strconv.Itoa(chatID)
	params["reply_to_message_id"] = strconv.Itoa(messageID)
	params["caption"] = card.Name
	ms.WriteFields(params)

	//Getting the image
	var resp *http.Response
	var extension string
	var paramType string
	var methodName string
	if isgold {
		resp , err = client.Get(card.ImgGold)
		c.Debugf("downloading image status: %v", resp.Status)
		extension = "gif"
		paramType = "document"
		methodName = "sendDocument"
	} else {
		resp, err = client.Get(card.Img)
		c.Debugf("downloading image status: %v", resp.Status)
		extension = "png"
		paramType = "photo"
		methodName = "sendPhoto"
	}

	defer resp.Body.Close()

	contentLength, _ := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
	c.Debugf("%v", contentLength)
	ms.WriteReader(paramType, "card." + extension, contentLength, resp.Body)
	req, _ := http.NewRequest("POST", BASE_URL + methodName, nil)
	ms.SetupRequest(req)

	resp, err = client.Do(req)
	//getting the response
	decoder := json.NewDecoder(resp.Body)
	response := botApi.APIResponse{}
	err = decoder.Decode(&response)

	if response.Ok {
		var message botApi.Message
		json.Unmarshal(response.Result, &message)
		insertSentCard(c, card, message.Photo[0].FileID, isgold)
		c.Infof("sent card image: %v", card.Name)
	} else {
		c.Errorf("ok: %v, error_code : %v, description: %v",
			response.Ok, response.ErrorCode, response.Description)
	}
	return err
}

func sentCardExist(c appengine.Context, card Card, isGold bool) (fileID string){
	q := datastore.NewQuery(
		"SentCard").Filter(
		"CardID =", card.CardId).Filter("IsGold =", isGold)
	count, _ := q.Count(c)
	if count != 0 {
		t := q.Run(c)
		sc := SentCard{}
		if _, err := t.Next(&sc); err != nil {
			return sc.FileID
		}
	}
	return ""
}

func insertSentCard (c appengine.Context, card Card, fileID string, isGold bool) {
	sentCard := SentCard{}
	sentCard.CardID = card.CardId
	sentCard.FileID = fileID
	sentCard.IsGold = isGold
	datastore.Put(c, datastore.NewIncompleteKey(c, "SentCard", nil), sentCard)
}

func isMessageRepeated (c appengine.Context, msgID int, chatID int) bool{
	//Checking datastore and confirming its not a repeated message
	q := datastore.NewQuery(
		"RespondedMessage").Filter(
		"MessageId =", msgID).Filter(
		"ChatId =", chatID)
	count, _ := q.Count(c)
	if count == 0 {
		return false
	}
	return true
}

func isActiveBot(c appengine.Context, chatID int) bool{
	key, err := getStatusKey(c, chatID)
	if err == nil {
		stat := BotStatus{}
		if err := datastore.Get(c, key, &stat); err == nil {
			return stat.Active
		}
		return false
	}
	return false
}

func getStatusKey(c appengine.Context, chatID int) (*datastore.Key, error) {
	q:= datastore.NewQuery("BotStatus").Filter("ChatID =", chatID).KeysOnly()
	var statuses []BotStatus
	keys , err := q.GetAll(c, &statuses)
	if err == nil && len(keys) > 0{
		return keys[0], nil
	}
	return nil, err
}

func insertStatus(c appengine.Context, chatID int, status bool) error{
	stat := new(BotStatus)
	stat.Active = status
	stat.ChatID = chatID
	//If existing updating
	key, err := getStatusKey(c, chatID)
	if key != nil {
		_, err = datastore.Put(c, key, stat)
		return err
	}
	//else inserting new
	_, err = datastore.Put(c, datastore.NewIncompleteKey(c, "BotStatus", nil), stat)
	if err != nil {
		return err
	}
	return nil
}

func insertRespondedMessage(c appengine.Context, msgID int, chatID int) error {
	respMessage := new(RespondedMessage)
	respMessage.MessageId = msgID
	respMessage.Success = true
	respMessage.ChatId = chatID
	_, err := datastore.Put(c, datastore.NewIncompleteKey(c, "RespondedMessage", nil), respMessage)
	return err
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

//checkErr takes a ResponseWriter and and Error if there is an error
//then it Writes it to the ResponseWriter and return a http.StatusInternalServerError
func checkErr(w http.ResponseWriter, err error) {
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}