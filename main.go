package main

import (
	"net/http"
	//"io/ioutil"
	"fmt"
	"golang.org/x/net/html"
)

var testUrl = "https://egghead.io/lessons/tools-share-a-tmux-session-for-pair-programming-with-ssh"

func getFileRef(n *html.Node) string {
	nodeFound := false
	var fileRef string

	var f func(n *html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode  && n.Data == "meta"{
			//fmt.Println(n)
			for _, a := range n.Attr {
				if a.Key == "itemprop" && a.Val == "contentURL" {
					//fmt.Println(a)
					//fmt.Println(n)
					nodeFound = true
				}
				if nodeFound {
					for _, a := range n.Attr {
						if a.Key == "content"{
							//fmt.Println(a)
							fileRef = a.Val
							return
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if nodeFound {break}
			f(c)
		}
	}
	f(n)
	return fileRef
}

func main() {
	response, err := http.Get(testUrl)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()

	//body, err := ioutil.ReadAll(response.Body)
	//if err != nil {
	//	panic(err)
	//}
	//
	//res := string(body)
	//fmt.Println(res)

	doc, err := html.Parse(response.Body)
	if err != nil {
		panic(err)
	}

	fileRef := getFileRef(doc)
	fmt.Println(fileRef) // TODO: substitute ".bin" with "/file.mp4"
	res := fmt.Sprintf("https://embed-ssl.wistia.com/deliveries/%s/file.mp4", fileRef)
	fmt.Println(res)
}
