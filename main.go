package main

import (
	"net/http"
	"io/ioutil"
	"fmt"
)

var testUrl = "https://egghead.io/lessons/tools-share-a-tmux-session-for-pair-programming-with-ssh"

func main() {
	resp, err := http.Get(testUrl)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	res := string(body)
	fmt.Println(res)
}
