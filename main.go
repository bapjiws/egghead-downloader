package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"

	"golang.org/x/net/html"
)

// TODO: use os.Args[1:] or, better yet, flags
var courseUrl = "https://egghead.io/courses/learn-the-best-and-most-useful-scss"
var serviceMu sync.Mutex

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
		if n.Type == html.ElementNode && n.Data == "meta" {
			//fmt.Println(n)
			for _, a := range n.Attr {
				if a.Key == "itemprop" && a.Val == "contentURL" {
					//fmt.Println(a)
					//fmt.Println(n)
					nodeFound = true
				}
				if nodeFound {
					for _, a := range n.Attr {
						if a.Key == "content" {
							//fmt.Println(a)
							fileRef = a.Val
							return
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if nodeFound {
				break
			}
			f(c)
		}
	}
	f(doc)

	return strings.Replace(fileRef, ".bin", "/file.mp4", 1)
}

func getLessons(url string) []string {
	var lessons []string
	doc := getDocFromUrl(url)

	var f func(n *html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			//fmt.Println(n)
			for _, a := range n.Attr {
				if a.Key == "href" && strings.Index(a.Val, "https://egghead.io/lessons/") != -1 {
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

	// TODO: create a struct to store url and lesson name.
	fileRefs := map[string]string{}

	lessons := getLessons(courseUrl)

	var wg sync.WaitGroup

	for _, lesson := range lessons {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()

			fileRef := getFileRef(url)

			fileId := strings.TrimLeft(fileRef, "http://embed.wistia.com/deliveries/")
			fileId = strings.TrimRight(fileId, "/file.mp4")

			serviceMu.Lock()
			defer serviceMu.Unlock()
			if _, ok := fileRefs[fileId]; !ok {
				fileRefs[fileId] = fileRef

				wg.Add(1)
				go func(url string) {
					defer wg.Done()

					fmt.Printf("Downloading file from %s\n", fileRef)
					out, _ := os.Create(fmt.Sprintf("%s.mp4", fileId))
					defer out.Close()

					/*Right now using https://embedwistia-a.akamaihd.net/deliveries/<file_id>/file.mp4.
					An alternative: fmt.Sprintf("https://embed-ssl.wistia.com/deliveries/%s/file.mp4", fileRef)*/
					resp, _ := http.Get(fileRef)
					defer resp.Body.Close()

					n, _ := io.Copy(out, resp.Body)
					fmt.Printf("Bytes copied: %d\n", n)
				}(fileRef)
			}
		}(lesson)
	}
	wg.Wait()

	fmt.Printf("Total lessons downloaded: %d\n", len(fileRefs))
}
