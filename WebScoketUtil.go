package main

import (
	"bytes"
	"net/http"
)

func WebSocketMessage(message string, token string) {
	url := config.WebSocket.Url + "?method=broadcast&message=" + token + "|" + message
	req, err := http.NewRequest("GET", url, new(bytes.Buffer))

	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	_, err = client.Do(req)

	if err != nil {
		panic(err)
	}
}
