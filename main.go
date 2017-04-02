package main

import (
	"net/http"
	//"io/ioutil"
	"fmt"
	"golang.org/x/net/html"
	"strings"
	"sync"
)

// TODO: use os.Args[1:] or, better yet, flags
var courseUrl = "https://egghead.io/courses/wrangle-your-terminal-with-tmux"

func getDocFromUrl(url string) *html.Node {
	response, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()

	doc, err := html.Parse(response.Body)
	if err != nil {
		panic(err)
	}

	return doc
}

func getFileRef(url string) string {
	var nodeFound bool
	var fileRef string
	doc := getDocFromUrl(url)

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
	f(doc)

	// Substitute ".bin" with "/file.mp4"
	return strings.Replace(fileRef,".bin", "/file.mp4", 1)
}

func getLessons(url string) []string {
	var lessons []string
	doc := getDocFromUrl(url)

	var f func(n *html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode  && n.Data == "a"{
			//fmt.Println(n)
			for _, a := range n.Attr {
				if a.Key == "href" && strings.Index(a.Val, "https://egghead.io/lessons/") != -1{
					//fmt.Println(a)
					lessons = append(lessons, a.Val)
				}

			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return lessons
}

// TODO: Check with https://golang.org/doc/articles/race_detector.html
func main() {
	// This piece is useful to display the entire html.
	//body, err := ioutil.ReadAll(response.Body)
	//if err != nil {
	//	panic(err)
	//}
	//
	//res := string(body)
	//fmt.Println(res)

	fileRefs := []string{}

	lessons := getLessons(courseUrl)
	//fmt.Println(lessons[0])

	var wg sync.WaitGroup

	for _, lesson := range lessons {
		wg.Add(1)
		go func(url string) {
			// Decrement the counter when the goroutine completes.
			defer wg.Done()

			fileRef := getFileRef(url)
			//fmt.Println(fileRef)
			fileRefs = append(fileRefs, fileRef)
		}(lesson)
	}
	wg.Wait()

	fmt.Println(fileRefs)
	fmt.Println(len(fileRefs))

	//res := fmt.Sprintf("https://embed-ssl.wistia.com/deliveries/%s/file.mp4", fileRef)
	//fmt.Println(res)
}
