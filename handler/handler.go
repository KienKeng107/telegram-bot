package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

// Define a few constants and variable to handle different commands
const punchCommand string = "/punch"

var lenPunchCommand int = len(punchCommand)

const startCommand string = "/start"

var lenStartCommand int = len(startCommand)

const botTag string = "@UncleK"

var lenBotTag int = len(botTag)

// Pass token and sensible APIs through environment variables
const telegramApiBaseUrl string = "https://api.telegram.org/bot"
const telegramApiSendMessage string = "/sendMessage"
const telegramTokenEnv string = "1705894753:AAEOhf0xjdlLfkecsoMImH_2t4kbdfxXmfQ"

var telegramApi string = telegramApiBaseUrl + os.Getenv(telegramTokenEnv) + telegramApiSendMessage

const rapLyricsApiEnv string = "3hu8pok7H6fCJlejQuOXJ1FrznOyMkF64ZyAjotGoBJM6laVvzEeF2lxpYFRqpAm"

var rapLyricsApi string = os.Getenv(rapLyricsApiEnv)

// Update is a Telegram object that the handler receives every time an user interacts with the bot
type Update struct {
	UpdateId int     `json:"update_id"`
	Message  Message `json:"message"`
}

// Implements the fmt.String interface to get the representation of an Update as a string.
func (u Update) String() string {
	return fmt.Sprintf("(update id: %d, message:%s)", u.UpdateId, u.Message)
}

// Message is a Telegram object that can be found in an update
// Note that not all Update contains a Message. Update for an Inline Query doesn't.
type Message struct {
	Text string `json:"text"`
	Chat Chat   `json:"chat"`
}

func (m Message) String() string {
	return fmt.Sprintf("(text: %s, chat: %s)", m.Text, m.Chat)
}

// type Audio struct {
// 	FileId   string `json:"file_id"`
// 	Duration string `json:"duration"`
// }

// func (a Audio) String() string {
// 	return fmt.Sprintf("(file id: %s, duration: %s)", a.FileId, a.Duration)
// }

// type Voice Audio

// type Document struct {
// 	FileId   string `json:"file_id"`
// 	FileName string `json:"file_name"`
// }

// func (d Document) String() string {
// 	return fmt.Sprintf("(file id: %s, file name: %s)", d.FileId, d.FileName)
// }

// A Telegram Chat indicates the converstation to which the message belongs
type Chat struct {
	Id int `json:"id"`
}

func (c Chat) String() string {
	return fmt.Sprintf("(id: %d)", c.Id)
}

type Lyric struct {
	Punch string `json:"output"`
}

// HandleTelegramWebHook sends a message back to the chat with a punchline starting by the message provided by the user.
func HandleTelegramWebHook(w http.ResponseWriter, r *http.Request) {

	//Parse incoming request
	log.Printf("Starting server......")
	var update, err = parseTelegramRequest(r)
	if err != nil {
		log.Printf("Error parsing update, %s", err.Error())
		return
	}

	// Sanitize input
	var sanitizedSeed = sanitize(update.Message.Text)

	// Call RepLyrics to get a punchline
	var lyric, errRaplyrics = getPunchline(sanitizedSeed)
	if errRaplyrics != nil {
		log.Printf("Got error when calling RapLyrics API %s", errRaplyrics.Error())
		return
	}

	// Send the punchline back to Telegram
	var telegramReponseBody, errTelegram = sendTextToTelegramChat(update.Message.Chat.Id, lyric)
	if errTelegram != nil {
		log.Printf("Got error %s from telegram, reponse body is %s", errTelegram.Error(), telegramReponseBody)
	} else {
		log.Printf(" Punchline %s successfuly distributed to chat id %d", lyric, update.Message.Chat.Id)
	}
}

// parseTelegramRequest handles incoming update from the Telegram web hook
func parseTelegramRequest(r *http.Request) (*Update, error) {
	var update Update
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		log.Printf("Counld not decode incoming update %s", err.Error())
		return nil, err
	}
	if update.UpdateId == 0 {
		log.Printf("invalid update id, got update id = 0")
		return nil, errors.New("Invalid update id of 0 indicates failure to parse incoming update")
	}
	return &update, nil
}

// sanitize remove clutter like/start /punch or the bot name from the string s passed as input
func sanitize(s string) string {
	if len(s) >= lenStartCommand {
		if s[:lenStartCommand] == startCommand {
			s = s[lenStartCommand:]
		}
	}

	if len(s) >= lenPunchCommand {
		if s[:lenPunchCommand] == punchCommand {
			s = s[lenPunchCommand:]
		}
	}
	if len(s) >= lenBotTag {
		if s[:lenBotTag] == botTag {
			s = s[lenBotTag:]
		}
	}

	return s
}

// getPunchline calls the RapLyrics API to get a punchline back
func getPunchline(seed string) (string, error) {
	rapLyricResp, err := http.PostForm(rapLyricsApi, url.Values{"input": {seed}})
	if err != nil {
		log.Printf("Error while calling raplyrics %s", err.Error())
		return "", err
	}
	var punchline Lyric
	if err := json.NewDecoder(rapLyricResp.Body).Decode(&punchline); err != nil {
		log.Printf("Could not decode incoming punchline %s", err.Error())
		return "", err
	}

	defer rapLyricResp.Body.Close()
	return punchline.Punch, nil
}

// sendTextToTelegramChat sends a text message to the Telegram chat indentified by its chat Id
func sendTextToTelegramChat(chatId int, text string) (string, error) {
	log.Printf("Sending %s to chat_id: %d", text, chatId)
	response, err := http.PostForm(
		telegramApi,
		url.Values{
			"chat_id": {strconv.Itoa(chatId)},
			"text":    {text},
		})
	if err != nil {
		log.Printf("Error when posting text to the chat: %s", err.Error())
		return "", err
	}
	defer response.Body.Close()
	var bodyBytes, errRead = ioutil.ReadAll(response.Body)
	if errRead != nil {
		log.Printf("Error in parsing telegram answer %s", errRead.Error())
		return "", err
	}
	bodyString := string(bodyBytes)
	log.Printf("Body of Telegram Response: %s", bodyString)

	return bodyString, nil
}
