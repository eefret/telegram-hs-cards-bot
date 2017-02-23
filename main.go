package main

import (
	"github.com/eefret/hsapi"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	log "github.com/sirupsen/logrus"
	"flag"
	"fmt"
	"os"
	"strings"
	"github.com/eefret/telegram-hs-cards-bot/bot"
	"github.com/garyburd/redigo/redis"
)

//----------------------------------------------Initialization
const (
	MASHAPE_KEY string = "tntkXJyM7EmshBgQYsXtCHHEX8Izp1uHrN1jsnTpw7tNCxEZIN"
)

var (
	Token           string
	HsapiToken      string
	RedisConnString string
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.StringVar(&HsapiToken, "h", "", "Hearthstone API Token")
	flag.StringVar(&RedisConnString, "r", ":6379", "Redis connection string")
	flag.Parse()
}

func main() {
	//Tokens
	if len(Token) == 0 {
		log.Panic("No token provided.. Exiting...")
		return
	}
	if len(HsapiToken) == 0 {
		HsapiToken = MASHAPE_KEY;
	}
	//Logging setup
	f, err := os.OpenFile("errors.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Println("there was an error opening the file...")
		return
	}
	defer f.Close()
	log.SetOutput(f)
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.JSONFormatter{})

	//getting bot, redis and api
	tgBot, err := tgbotapi.NewBotAPI(Token);
	if checkErr(err) {
		return
	}
	log.Printf("Authorised on account %s", tgBot.Self.UserName)
	api := hsapi.NewHsAPI(HsapiToken)

	u := tgbotapi.NewUpdate(1)
	updates, err := tgBot.GetUpdatesChan(u);
	if checkErr(err) {
		return
	}
	redisConn, err := redis.Dial("tcp", RedisConnString);
	if checkErr(err) {
		return
	}

	hsBot := bot.NewBot(api, tgBot, redisConn)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		log.Printf("[%s] %s ", update.Message.From.UserName, update.Message.Text)
		if strings.HasPrefix(update.Message.Text, "/") {
			hsBot.HandleMessage(update)
		}
	}
}

func checkErr(err error) bool {
	if err != nil {
		log.Error(err)
		return true
	}
	return false
}
