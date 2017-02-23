package bot

import (
	"github.com/eefret/hsapi"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"strings"
	"github.com/garyburd/redigo/redis"
	"fmt"
	log "github.com/sirupsen/logrus"
)

type Bot struct {
	Api   *hsapi.HsAPI
	Bot   *tgbotapi.BotAPI
	Redis redis.Conn
}

const (
	//commands
	COMMAND_NORMAL_CARD      = "/hs "
	COMMAND_GOLD_CARD        = "/ghs "
	COMMAND_CARD_SOUND       = "/shs "
	COMMAND_HIDDEN_CARD      = "/hhs "
	COMMAND_GOLD_HIDDEN_CARD = "/ghhs "
	COMMAND_START            = "/start"
	COMMAND_STOP             = "/stop"

	//params
	PARAM_GOLD       = "gold"
	PARAM_INDEX      = "index"
	PARAM_SOUND_TYPE = "sound_type"
	PARAM_COMMAND    = "command"
	MAX_PARAMS       = 3

	//redis keys
	KEY_STATUS = "status_chat-%d"
)

func NewBot(api *hsapi.HsAPI, bot *tgbotapi.BotAPI, redisConn redis.Conn) *Bot {
	return &Bot{
		Api:   api,
		Bot:   bot,
		Redis: redisConn,
	}
}

func (self *Bot) HandleMessage(update tgbotapi.Update) {

	config := hsapi.NewCardSearch("")
	var params map[string]interface{}
	switch true {
	case strings.HasPrefix(update.Message.Text, COMMAND_START):
		self.toggleBot(true, update)
		break
	case strings.HasPrefix(update.Message.Text, COMMAND_STOP):
		self.toggleBot(false, update)
		break
	case strings.HasPrefix(update.Message.Text, COMMAND_NORMAL_CARD):
		config.Name, params = parseParams(update.Message.Text, COMMAND_NORMAL_CARD)
		config.Collectible = true
		params[PARAM_GOLD] = false
		break
	case strings.HasPrefix(update.Message.Text, COMMAND_GOLD_CARD):
		config.Name, params = parseParams(update.Message.Text, COMMAND_GOLD_CARD)
		config.Collectible = true
		params[PARAM_GOLD] = true
		break
	case strings.HasPrefix(update.Message.Text, COMMAND_CARD_SOUND):
		config.Name, params = parseParams(update.Message.Text, COMMAND_CARD_SOUND)
		config.Collectible = true
		break
	case strings.HasPrefix(update.Message.Text, COMMAND_HIDDEN_CARD):
		config.Name, params = parseParams(update.Message.Text, COMMAND_HIDDEN_CARD)
		params[PARAM_GOLD] = false
		break
	case strings.HasPrefix(update.Message.Text, COMMAND_GOLD_HIDDEN_CARD):
		config.Name, params = parseParams(update.Message.Text, COMMAND_GOLD_HIDDEN_CARD)
		params[PARAM_GOLD] = true
		break
	}

	if !self.isActive(update.Message.Chat.ID) {
		log.Debug("bot not active in this chat")
		return
	}

	log.Debugf("%#v", params)
	log.Debugf("%#v", config)

	/*if soundType, ok := params[PARAM_SOUND_TYPE]; ok {
		//TODO implement sounds
	}*/

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	if cards, err := self.Api.Search(config); err != nil || len(cards) != 0 {
		if _, ok := params[PARAM_INDEX]; !ok && len(cards) > 1 {
			msg.Text = createMultiCardError(cards, config.Name, params[PARAM_COMMAND].(string))
			msg.ReplyToMessageID = update.Message.MessageID
			log.Debugf("%#v", msg)
			self.Bot.Send(msg)
			return
		}

		var card hsapi.Card
		if index, ok := params[PARAM_INDEX].(int); ok {
			card = cards[index]
		} else {
			card = cards[0]
		}

		if params[PARAM_GOLD].(bool) {
			msg.Text = card.ImgGold
		} else {
			msg.Text = card.Img
		}

		msg.ReplyToMessageID = update.Message.MessageID
		log.Debugf("%#v", msg)
		self.Bot.Send(msg)
		return
	}

	msg.Text = "couldn't find that card, try with another query"
	msg.ReplyToMessageID = update.Message.MessageID
	log.Debugf("%#v", msg)
	self.Bot.Send(msg)
}

func (self *Bot) toggleBot(flag bool, update tgbotapi.Update) {
	if flag == self.isActive(update.Message.Chat.ID) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("The bot status is already %t", flag))
		msg.ReplyToMessageID = update.Message.MessageID
		self.Bot.Send(msg)
		return
	}
	_, err := self.Redis.Do("SET", fmt.Sprintf(KEY_STATUS, update.Message.Chat.ID), flag)
	log.Debugf("%s set to %b", fmt.Sprintf(KEY_STATUS, update.Message.Chat.ID), flag)
	log.Error(err)
}

func (self *Bot) isActive(channelID int64) bool {
	stat, err := redis.Bool(self.Redis.Do("GET", fmt.Sprintf(KEY_STATUS, channelID)));
	log.Debug(stat)
	log.Error(err)
	return stat
}
