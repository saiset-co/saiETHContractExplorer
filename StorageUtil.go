package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/webmakom-com/mycointainer/src/Storage/src/gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"net/http"
	"time"
)

type Token struct {
	Id    string `json:"_id"`
	Token string `json:"token"`
}

type Login struct {
	email    string
	password string
}

type StorageRequest struct {
	login        Login
	collection   string
	selectString interface{}
	options      string
	token        string
	data         interface{}
}

var myToken Token

func (s StorageRequest) toJson() ([]byte, error) {
	if (Login{}) != s.login {
		return json.Marshal(bson.M{"email": s.login.email, "password": s.login.password})
	}

	if s.selectString != nil {
		return json.Marshal(bson.M{"collection": s.collection, "select": s.selectString, "token": s.token, "data": s.data})
	}

	return json.Marshal(bson.M{"collection": s.collection, "token": s.token, "data": s.data})
}

func login() {
	request := StorageRequest{login: Login{email: config.Storage.Auth.Email, password: config.Storage.Auth.Password}}
	token := CallSaiStrorage("login", request)
	_ = json.Unmarshal([]byte(token), &myToken)
}

func CallSaiStrorage(method string, request StorageRequest) string {
	if method != "login" && myToken.Token == "" {
		login()
	}

	request.token = myToken.Token

	return callSaiStrorage(config.Storage.Url+"/"+method, request)
}

func callSaiStrorage(url string, request StorageRequest) string {

	jsonStr, jsonErr := request.toJson()
	fmt.Println(string(jsonStr))

	if jsonErr != nil {
		panic(jsonErr)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	_ = time.AfterFunc(5*time.Second, func() {
		resp.Body.Close()
	})
	body, _ := ioutil.ReadAll(resp.Body)
	return string(body)
}
