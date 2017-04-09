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

func getFile(l lesson) file {
	doc := getDocFromUrl(l.url)

	// TODO: Do this inside one func to avoid traversing the doc twice
	return file{
		url:  getFileUrl(doc),
		name: fmt.Sprintf("%d. %s", l.order, getFileName(doc)), // TODO: improve
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

type lesson struct {
	url   string
	order int
}

func getLessonUrls(courseUrl string) []lesson {
	lessonIds := map[string]bool{}
	lessons := []lesson{}
	doc := getDocFromUrl(courseUrl)
	lessonOrder := 1

	var f func(n *html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			//fmt.Println(n)
			for _, a := range n.Attr {
				if a.Key == "href" && strings.Index(a.Val, "https://egghead.io/lessons/") != -1 && !lessonIds[a.Val] {
					//fmt.Println(a)
					lessonIds[a.Val] = true
					lessons = append(lessons, lesson{url: a.Val, order: lessonOrder})
					lessonOrder++
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

type file struct {
	url  string
	name string
}

// TODO: Check with https://golang.org/doc/articles/race_detector.html
func main() {

	ls := getLessonUrls(courseUrl)
	fmt.Println(len(ls))

	var wg sync.WaitGroup

	// TODO: Use fain-out pattern, i.e., make ls a channel
	for _, l := range ls {
		wg.Add(1)
		go func(l lesson) {
			defer wg.Done()

			f := getFile(l)
			fmt.Println(f)

			wg.Add(1)

			go func(f file) {
				defer wg.Done()
				// TODO: handle all the errors below.
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
		}(l)
	}
	wg.Wait()

	// TODO: use atomic counter to track the number of downloaded files.
	//fmt.Printf("Total lessonUrls downloaded: %d\n", len(files))
}
