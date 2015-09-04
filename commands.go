package hsimagebot
import (
	"strings"
	botApi "github.com/Syfaro/telegram-bot-api"
	"regexp"
	"strconv"
	"appengine"
	"net/http"
	"appengine/urlfetch"
)


func commandHs(w http.ResponseWriter, r *http.Request, msg botApi.Message) {
	bot, _ := botApi.NewBotAPIWithClient(TOKEN, urlfetch.Client(appengine.NewContext(r)))
	c := appengine.NewContext(r)
	var isGold bool
	var index int = -1
	var cardName string
	messageConfig := botApi.NewMessage(msg.Chat.ID, "")
	messageConfig.ReplyToMessageID  = msg.MessageID

	//Extracting if gold or not
	if strings.HasPrefix(msg.Text, "/hsg"){
		isGold = true
	}

	//Extracting index if any
	if regexp.MustCompile("{").MatchString(msg.Text){
		result := regexp.MustCompile("\\{(.*?)\\}").FindAllStringSubmatch(msg.Text, -1)
		for _, v := range result {
			if num, err := strconv.Atoi(v[1]); err == nil {
				index = num
				break;
			}
		}
	}

	//Extracting cardname
	if isGold {
		if index != -1 {
			cardName = strings.TrimSpace(regexp.MustCompile("\\/hsg(.*?)\\{").FindStringSubmatch(msg.Text)[1])
		}else {
			cardName = strings.TrimSpace(msg.Text[4:])
		}
	}else {
		if index != -1 {
			cardName = strings.TrimSpace(regexp.MustCompile("\\/hs(.*?)\\{").FindStringSubmatch(msg.Text)[1])
		} else {
			cardName = strings.TrimSpace(msg.Text[3:])
		}
	}

	//getting cards
	cards := getCard(r, cardName)
	c.Infof("cards found: %v, cardName: %v, is gold: %v, index: %v", len(cards), cardName, isGold, index)

	if len(cards) == 0 {
		messageConfig.Text = "Sorry, I couldn't find that card"
		bot.SendMessage(messageConfig)
		c.Infof("card not found")
	} else if len(cards) > 1 && index == -1 {
		messageConfig.Text = createMulticardError(cards, cardName)
		bot.SendMessage(messageConfig)
		c.Infof("multicardError")
	} else if len(cards) == 1 && index == -1 {
		if fileID := sentCardExist(c, cards[0], isGold); len(fileID) != 0 {
			config := botApi.NewPhotoShare(msg.Chat.ID, fileID)
			config.Caption = cards[0].Name
			config.ReplyToMessageID = msg.MessageID
			bot.SendPhoto(config)
			c.Infof("replied an already uploaded image card %v", config.Caption)
		} else{
			err := replyCard(msg.MessageID, msg.Chat.ID, cards[0], c, isGold)
			checkErr(w, err)
		}
	} else if len(cards) > 1 && index >= 0 {
		if fileID := sentCardExist(c, cards[index], isGold); len(fileID) != 0 {
			config := botApi.NewPhotoShare(msg.Chat.ID, fileID)
			config.Caption = cards[index].Name
			config.ReplyToMessageID = msg.MessageID
			bot.SendPhoto(config)
			c.Infof("replied an already uploaded image card %v", config.Caption)
		} else{
			err := replyCard(msg.MessageID, msg.Chat.ID, cards[index], c, isGold)
			checkErr(w, err)
		}
	}
}