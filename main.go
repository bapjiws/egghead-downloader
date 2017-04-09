package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"golang.org/x/net/html"
	"os"
	"io"
)

// TODO: use os.Args[1:] or, better yet, flags
var courseUrl = "https://egghead.io/courses/learn-the-best-and-most-useful-scss"

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

func getFile(url string) file {
	doc := getDocFromUrl(url)

	// TODO: Do this inside one func to avoid traversing the doc twice
	return file{
		url:  getFileUrl(doc),
		name: getFileName(doc),
	}

}

func getFileUrl(doc *html.Node) string {
	nodeFound := false
	fileUrl := ""

	var f func(n *html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "meta" {
			for _, a := range n.Attr {
				if a.Key == "itemprop" && a.Val == "contentURL" {
					//fmt.Println(a)
					//fmt.Println(n)
					for _, a := range n.Attr {
						if a.Key == "content" {
							//fmt.Println(a)
							fileUrl = strings.Replace(a.Val, ".bin", "/file.mp4", 1)
							nodeFound = true
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

	return fileUrl
}

func getFileName(doc *html.Node) string {
	fileName := ""
	nodeFound := false

	var f func(n *html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "meta" {
			//fmt.Println(n)
			for _, a := range n.Attr {
				if a.Key == "itemprop" && a.Val == "name" {
					//fmt.Println(a)
					//fileName = a.Val
					for _, a := range n.Attr {
						if a.Key == "content" {
							fileName = a.Val
							//fmt.Println(fileName)
							nodeFound = true
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

	return fileName
}

func getLessonUrls(courseUrl string) []string {
	lessonIds := map[string]bool{}
	lessonUrls := []string{}
	doc := getDocFromUrl(courseUrl)

	var f func(n *html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			//fmt.Println(n)
			for _, a := range n.Attr {
				if a.Key == "href" && strings.Index(a.Val, "https://egghead.io/lessons/") != -1 && !lessonIds[a.Val] {
					//fmt.Println(a)
					lessonIds[a.Val] = true
					lessonUrls = append(lessonUrls, a.Val)
				}

			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return lessonUrls
}

type file struct {
	url  string
	name string
}

// TODO: Check with https://golang.org/doc/articles/race_detector.html
func main() {

	lessonUrls := getLessonUrls(courseUrl)
	fmt.Println(len(lessonUrls))

	var wg sync.WaitGroup

	// TODO: Use fain-out pattern, i.e., make lessonUrls a channel
	for _, lessonUrl := range lessonUrls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()

			f := getFile(url)
			//fmt.Println(file)

			wg.Add(1)

			go func(f file) {
				defer wg.Done()
				// TODO: handle all the errors below.
				// TODO: save files in order.
				fmt.Printf("Downloading file from %s\n", f.url)
				out, _ := os.Create(fmt.Sprintf("%s.mp4", f.name))
				defer out.Close()

				/*Right now using https://embedwistia-a.akamaihd.net/deliveries/<file_id>/file.mp4.
				An alternative: fmt.Sprintf("https://embed-ssl.wistia.com/deliveries/%s/file.mp4", url)*/
				resp, _ := http.Get(f.url)
				defer resp.Body.Close()

				n, _ := io.Copy(out, resp.Body)
				fmt.Printf("Bytes copied: %d\n", n)
			}(f)
		}(lessonUrl)
	}
	wg.Wait()

	// TODO: use atomic counter to track the number of downloaded files.
	//fmt.Printf("Total lessonUrls downloaded: %d\n", len(files))
}
