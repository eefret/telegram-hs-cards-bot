package hsimagebot

type RespondedMessage struct {
	MessageId int64
	ChatId    int64
	Success   bool
}

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

type Update struct {
	UpdateId int64   `json:"update_id"`
	Message  Message `json:"message"`
}

type Message struct {
	MessageId           int64       `json:"message_id"`
	From                User        `json:"from"`
	Date                int64       `json:"date"`
	Chat                GroupChat   `json:"chat"`
	ForwardFrom         User        `json:"forward_from"`
	ForwardDate         int64       `json:"forward_date"`
	ReplyToMessage      *Message    `json:"reply_to_message"`
	Text                string      `json:"text"`
	Audio               Audio       `json:"audio"`
	Document            Document    `json:"document"`
	Photo               []PhotoSize `json:"photo"`
	Sticker             Sticker     `json:"sticker"`
	Video               Video       `json:"video"`
	Voice               Voice       `json:"voice"`
	Caption             string      `json:"caption"`
	Contact             Contact     `json:"contact"`
	Location            Location    `json:"location"`
	NewChatParticipant  User        `json:"new_chat_participant"`
	LeftChatParticipant User        `json:"left_chat_participant"`
	NewChatTitle        string      `json:"new_chat_title"`
	NewChatPhoto        []PhotoSize `json:"new_chat_photo"`
	DeleteChatPhoto     bool        `json:"delete_chat_photo"`
	GroupChatCreated    bool        `json:"group_chat_created"`
}

type PhotoSize struct {
	FileId   string `json:"file_id"`
	Width    int64  `json:"width"`
	Height   int64  `json:"height"`
	FileSize int64  `json:"file_size"`
}

type Audio struct {
	FileId    string `json:"file_id"`
	Duration  int64  `json:"duration"`
	Performer string `json:"performer"`
	Title     string `json:"title"`
	MimeType  string `json:"mime_type"`
	FileSize  int64  `json:"file_size"`
}

type Document struct {
	FileId   string    `json:"file_id"`
	Thumb    PhotoSize `json:"thumb"`
	FileName string    `json:"file_name"`
	MimeType string    `json:"mime_type"`
	FileSize int64     `json:"file_size"`
}

type Sticker struct {
	FileId   string    `json:"file_id"`
	Width    int64     `json:"width"`
	Height   int64     `json:"height"`
	Thumb    PhotoSize `json:"thumb"`
	FileSize int64     `json:"file_size"`
}

type Video struct {
	FileId   string    `json:"file_id"`
	Width    int64     `json:"width"`
	Height   int64     `json:"height"`
	Duration int64     `json:"duration"`
	Thumb    PhotoSize `json:"thumb"`
	MimeType string    `json:"mime_type"`
	FileSize int64     `json:"file_size"`
}

type Voice struct {
	FileId   string `json:"file_id"`
	Duration int64  `json:"duration"`
	MimeType string `json:"mime_type"`
	FileSize int64  `json:"file_size"`
}

type Contact struct {
	PhoneNumber string `json:"phone_number"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	UserId      int64  `json:"user_id"`
}

type Location struct {
	Longitude float32 `json:"longitude"`
	Latitude  float32 `json:"latitude"`
}

type UserProfilePhotos struct {
	TotalCount int16         `json:"total_count"`
	Photos     [][]PhotoSize `json:"photos"`
}

type User struct {
	Id        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
}

type GroupChat struct {
	Id    int64  `json:"id"`
	Title string `json:"title"`
}
