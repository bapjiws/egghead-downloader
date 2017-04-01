package main

import (
	"net/http"
	//"io/ioutil"
	"fmt"
	"golang.org/x/net/html"
)

var testUrl = "https://egghead.io/lessons/tools-share-a-tmux-session-for-pair-programming-with-ssh"

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

	//z := html.NewTokenizer(response.Body)
	//for {
	//	tt := z.Next()
	//
	//	switch {
	//	case tt == html.ErrorToken:
	//		// End of the document, we're done
	//		return
	//	case tt == html.StartTagToken:
	//		//t := z.Token()
	//		//if t.Data == "meta" {
	//		//	for _, a := range t.Attr {
	//		//		if a.Key == "content" {
	//		//			fmt.Println(a)
	//		//		}
	//		//	}
	//		//}
	//
	//		t := z.Token()
	//		if t.Data == "div" {
	//			for _, a := range t.Attr {
	//				if a.Key == "ng-app" {
	//					fmt.Println(a)
	//				}
	//			}
	//		}
	//	}
	//}

	doc, err := html.Parse(response.Body)
	if err != nil {
		panic(err)
	}
	//fmt.Printf("%T\n", doc)
	//fmt.Println(doc.FirstChild.PrevSibling.Data)

	// <meta itemprop="thumbnailUrl" content="https://embed-ssl.wistia.com/deliveries/9e2143e3cfe64ef7c9c3dd404a337cee3ea5cdd0.bin" />
	// { ng-app egghead-lesson}
	var fileId string
	var f func(n *html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode  && n.Data == "meta"{ //&& n.Data == "meta" {
			//fmt.Println(n)
			for _, a := range n.Attr {
				rightNode := false
				if a.Key == "itemprop" && a.Val == "contentURL" {
					//fmt.Println(a)
					//fmt.Println(n)
					rightNode = true
				}
				if rightNode {
					for _, a := range n.Attr {
						if a.Key == "content"{
							//fmt.Println(a)
							fileId = a.Val
							return
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	fmt.Println(fileId) // TODO: substitute ".bin" with "/file.mp4"
	//res := fmt.Sprintf("https://embed-ssl.wistia.com/deliveries/%s/file.mp4", fileId)
	//fmt.Println(res)
}
