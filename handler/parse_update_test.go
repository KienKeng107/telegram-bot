package handler

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
)

var chat = Chat{Id: 1}

func TestParseUpdateMessageWithText(t *testing.T) {
	var msg = Message{
		Text: "Hello World",
		Chat: chat,
	}
	var update = Update{
		UpdateId: 1,
		Message:  msg,
	}
	requestBody, err := json.Marshal(update)
	if err != nil {
		t.Errorf("Failed to marshal update in json, got %s", err.Error())
	}

	req := httptest.NewRequest("POST", "http://myTelegramWebHookHandler.com/1705894753:AAEOhf0xjdlLfkecsoMImH_2t4kbdfxXmfQ", bytes.NewBuffer(requestBody))

	var updateToTest, errParse = parseTelegramRequest(req)
	if errParse != nil {
		t.Errorf("Expected a <nil> error, got %s", errParse.Error())
	}
	if *updateToTest != update {
		t.Errorf("Expected update %s, got %s", update, updateToTest)
	}
}

func TestParseUpdateInvalid(t *testing.T) {
	var msg = map[string]string{
		"bipbip": "12345",
		"bopbop": "wrong input",
	}
	requestBody, _ := json.Marshal(msg)
	req := httptest.NewRequest("POST", "http://myTelegramWebHookHandler.com/1705894753:AAEOhf0xjdlLfkecsoMImH_2t4kbdfxXmfQ", bytes.NewBuffer(requestBody))
	var _, err = parseTelegramRequest(req)
	if err == nil {
		t.Errorf("Expected an error, got <nil>")
	}
}
