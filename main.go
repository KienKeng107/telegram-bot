package main

import (
	"net/http"

	"github.com/KienKeng107/golang/handler"
)

func main() {
	http.ListenAndServe(":3000", http.HandlerFunc(handler.HandleTelegramWebHook))
}
