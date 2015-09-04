package hsimagebot


//Persistence structs
type BotStatus struct {
	ChatID int
	Active bool
}

type RespondedMessage struct {
	MessageId int
	ChatId    int
	Success   bool
}

type SentCard struct {
	FileID string
	CardID string
	IsGold bool
}


//Api structs
type CardResponse struct {
	Cards []Card
}

type Card struct {
	CardId      string `json:"cardId"`
	Name        string `json:"name"`
	CardSet     string `json:"cardSet"`
	Type        string `json:"type"`
	Faction     string `json:"faction"`
	Rarity      string `json:"rarity"`
	Cost        int16  `json:"cost"`
	Attack      int16  `json:"attack"`
	Health      int16  `json:"health"`
	Text        string `json:"text, string"`
	Flavor      string `json:"flavor"`
	Artist      string `json:"artist"`
	Collectible bool   `json:"collectible"`
	Elite       bool   `json:"elite"`
	PlayerClass string `json:"playerClass"`
	Img         string `json:"img"`
	ImgGold     string `json:"imgGold"`
	Locale      string `json:"locale"`
	Mechanics   string `json:"-"`
}

//Conf struct

type Conf struct{
	token string
}