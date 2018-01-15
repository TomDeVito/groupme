package groupme

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/pborman/uuid"
)

const apiUrl = "https://api.groupme.com/v3"

var httpClient = http.DefaultClient

func SetHttpClient(client *http.Client) {
	httpClient = client
}

type AttachmentType string

const (
	ImageType    AttachmentType = "image"
	LocationType                = "location"
	SplitType                   = "split"
	EmojiType                   = "emoji"
	MentionsType                = "mentions"
)

type SenderType string

const (
	SenderUser SenderType = "user"
	SenderBot             = "bot"
)

type App struct {
	token string
}

type Group struct {
	Id            string  `json:"id"`
	GroupId       string  `json:"group_id"`
	Name          string  `json:"name"`
	PhoneNumber   string  `json:"phone_number"`
	Type          string  `json:"type"`
	Description   string  `json:"description"`
	ImageUrl      string  `json:"image_url"`
	CreatorUserId string  `json:"creator_user_id"`
	CreatedAt     uint64  `json:"created_at"`
	UpdatedAt     uint64  `json:"updated_at"`
	OfficeMode    bool    `json:"office_mode"`
	ShareUrl      string  `json:"share_url"`
	Members       []User  `json:"members"`
	Messages      Message `json:"messages"`
	MaxMembers    uint    `json:"max_members"`
}

type ShortGroup struct {
	GroupId string `json:"group_id"`
	Name    string `json:"name"`
}

type User struct {
	Id          string `json:"id"`
	UserId      string `json:"user_id"`
	Name        string `json:"name"`
	Nickname    string `json:"nickname"`
	PhoneNumber string `json:"phone_number"`
	Email       string `json:"email"`
	Sms         bool   `json:"sms"`
	ImageUrl    string `json:"image_url"`
	Muted       bool   `json:"muted"`
	Autokicked  bool   `json:"autokicked"`
	CreatedAt   uint64 `json:"created_at"`
	UpdatedAt   uint64 `json:"updated_at"`
}

type MessagePreview struct {
	Nickname    string       `json:"nickname"`
	Text        string       `json:"text"`
	ImageUrl    string       `json:"image_url"`
	Attachments []Attachment `json:"attachments"`
}

type Attachment struct {
	/* common to all attachments */
	Type AttachmentType `json:"type"`

	/* image */
	Url string `json:"url,omitempty"`

	/* location */
	Name string `json:"name,omitempty"`
	Lat  string `json:"lat,omitempty"`
	Lng  string `json:"lng,omitempty"`

	/* split */
	Token string `json:"token,omitempty"`

	/* emoji */
	Placeholder string    `json:"placeholder,omitempty"`
	Charmap     []Charmap `json:"charmap,omitempty"`

	/* mentions */
	UserIds []string `json:"user_ids,omitempty"`
	Loci    []Locus  `json:"loci,omitempty"`
}

type Charmap [2]int

type Locus [2]int

type Message struct {
	Id          string       `json:"id"`
	SourceGuid  string       `json:"source_guid"`
	CreatedAt   uint32       `json:"created_at"`
	UserId      string       `json:"user_id"`
	GroupId     string       `json:"group_id"`
	Name        string       `json:"name"`
	AvatarUrl   string       `json:"avatar_url"`
	Text        string       `json:"text"`
	System      bool         `json:"system"`
	FavoritedBy []string     `json:"favorited_by"`
	Attachments []Attachment `json:"attachments,omitempty"`
	SenderType  SenderType   `json:"sender_type"`
}

type Bot struct {
	BotId       string `json:"bot_id"`
	GroupId     string `json:"group_id"`
	Name        string `json:"name"`
	AvatarUrl   string `json:"avatar_url"`
	CallbackUrl string `json:"callback_url"`
}

type BotMessage struct {
	BotId       string       `json:"bot_id"`
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

func NewApp(token string) *App {
	return &App{
		token: token,
	}
}

func Get(url string, respEnv interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header = map[string][]string{
		"Content-Type": {"application/json"},
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return fmt.Errorf("%s returned %d\n", url, resp.StatusCode)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	return json.Unmarshal(body, respEnv)
}

func post(url string, data string) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, strings.NewReader(data))
	if err != nil {
		return nil, err
	}

	req.Header = map[string][]string{
		"Content-Type": {"application/json"},
	}

	return httpClient.Do(req)
}

func Post(url string, data string, respEnv interface{}) error {
	resp, err := post(url, data)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return fmt.Errorf("%s returned %d\n", url, resp.StatusCode)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	return json.Unmarshal(body, respEnv)
}

func (app *App) GetBot(botId string) (*Bot, error) {
	respEnv := struct {
		Bots []Bot `json:"response"`
	}{}

	err := Get(apiUrl+"/bots?token="+app.token, &respEnv)

	for i, _ := range respEnv.Bots {
		if respEnv.Bots[i].BotId == botId {
			return &respEnv.Bots[i], nil
		}
	}

	return nil, err
}

func (app *App) Groups() ([]*Group, error) {
	respEnv := struct {
		Groups []*Group `json:"response"`
	}{}

	err := Get(apiUrl+"/groups?token="+app.token, &respEnv)

	return respEnv.Groups, err
}

func (app *App) FormerGroups() ([]Group, error) {
	respEnv := struct {
		Groups []Group `json:"response"`
	}{}

	err := Get(apiUrl+"/groups/former?token="+app.token, &respEnv)

	return respEnv.Groups, err
}

func (app *App) Group(id string) (*Group, error) {
	respEnv := struct {
		Group Group `json:"response"`
	}{}

	err := Get(apiUrl+"/groups/"+id+"?token="+app.token, &respEnv)

	return &respEnv.Group, err
}

func (app *App) GetMessages(group *Group, limit uint) ([]Message, error) {
	var limitStr string

	respEnv := struct {
		Response struct {
			Count    int       `json:"count"`
			Messages []Message `json:"messages"`
		} `json:"response"`
	}{}

	if limit == 0 {
		/* If limit is 0 then use default */
		limitStr = ""
	} else {
		if limit > 100 {
			limit = 100 /* max is 100 */
		}
		limitStr = fmt.Sprintf("&limit=%d", limit)
	}

	err := Get(apiUrl+"/groups/"+group.GroupId+"/messages?token="+app.token+limitStr, &respEnv)

	return respEnv.Response.Messages, err
}

func (app *App) SendMessage(group *Group, msg *Message) (*Message, error) {
	msgBody := struct {
		Message Message `json:"message"`
	}{
		Message: *msg,
	}

	respEnv := struct {
		Response struct {
			Message Message `json:"message"`
		} `json:"response"`
	}{}

	data, err := json.Marshal(msgBody)
	if err != nil {
		return nil, err
	}

	err = Post(apiUrl+"/groups/"+group.GroupId+"/messages?token="+app.token, string(data), &respEnv)
	return &respEnv.Response.Message, err
}

func (app *App) SendMessageText(group *Group, text string) (*Message, error) {
	msg := &Message{
		SourceGuid: NewSourceGuid(),
		Text:       text,
	}

	return app.SendMessage(group, msg)
}

func (app *App) GetUserMe() (*User, error) {
	respEnv := struct {
		Me User `json:"response"`
	}{}

	err := Get(apiUrl+"/users/me?token="+app.token, &respEnv)
	return &respEnv.Me, err
}

func (bot *Bot) SendMessage(text string, attachments []Attachment) (*http.Response, error) {
	msg := &BotMessage{
		BotId:       bot.BotId,
		Text:        text,
		Attachments: attachments,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", apiUrl+"/bots/post", strings.NewReader(string(data)))
	if err != nil {
		return nil, err
	}

	req.Header = map[string][]string{
		"Content-Type": {"application/json"},
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, err
}

func (group *Group) GetUser(userId string) *User {
	for _, member := range group.Members {
		if member.UserId == userId {
			return &member
		}
	}

	return nil
}

func (group Group) String() string {
	shortGroup := ShortGroup{
		GroupId: group.GroupId,
		Name:    group.Name,
	}

	json, _ := json.Marshal(shortGroup)
	return string(json)
}

func FindMessage(msgs []Message, id string) *Message {
	for _, msg := range msgs {
		if msg.Id == id {
			return &msg
		}
	}

	return nil
}

func NewSourceGuid() string {
	return uuid.NewUUID().String()
}
