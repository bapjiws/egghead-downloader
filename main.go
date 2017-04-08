package main

import (
	"fmt"
	//"io"
	"net/http"
	//"os"
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

func getFile(url string) file {
	doc := getDocFromUrl(url)

	// TODO: Do this inside one func to avoid traversing the doc twice
	return file{
		url:  getFileUrl(doc),
		name: getFileName(doc),
	}

}

func getFileUrl(doc *html.Node) string { // url string
	nodeFound := false
	fileUrl := ""
	//file := file{}
	//doc := getDocFromUrl(url)

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
							//fileUrl = strings.Replace(a.Val, ".bin", "/file.mp4", 1)
							fileUrl = a.Val
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

	return strings.Replace(fileUrl, ".bin", "/file.mp4", 1)
}

func getFileName(doc *html.Node) string { // url string
	fileName := ""
	nodeFound := false
	//file := file{}
	//doc := getDocFromUrl(url)

	var f func(n *html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "meta" {
			//fmt.Println(n)
			for _, a := range n.Attr {
				if a.Key == "itemprop" && a.Val == "name" {
					//fmt.Println(a.Val)
					//fileName = a.Val
					nodeFound = true
				}
				if nodeFound {
					for _, a := range n.Attr {
						if a.Key == "content" {
							fileName = a.Val
							//fmt.Println(fileName)
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

func getLessonUrls(url string) []string {
	var lessonUrls []string
	doc := getDocFromUrl(url)

	var f func(n *html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			//fmt.Println(n)
			for _, a := range n.Attr {
				if a.Key == "href" && strings.Index(a.Val, "https://egghead.io/lessons/") != -1 {
					//fmt.Println(a)
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
	// This piece is useful to display the entire html.
	//body, err := ioutil.ReadAll(response.Body)
	//if err != nil {
	//	panic(err)
	//}
	//
	//res := string(body)
	//fmt.Println(res)

	// TODO: create a struct to store url and lessonUrl name (h1 class="title").
	files := map[string]file{}

	lessonUrls := getLessonUrls(courseUrl)
	fmt.Println(lessonUrls)

	var wg sync.WaitGroup

	// TODO: Use fain-out pattern, i.e., make lessonUrls a channel
	for _, lessonUrl := range lessonUrls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()

			//fmt.Println(url)
			file := getFile(url)

			fileId := strings.TrimLeft(file.url, "http://embed.wistia.com/deliveries/")
			fileId = strings.TrimRight(fileId, "/file.mp4")

			serviceMu.Lock()
			defer serviceMu.Unlock()
			if _, ok := files[fileId]; !ok {
				files[fileId] = file

				//wg.Add(1)
				//go func(url string) {
				//	defer wg.Done()
				//
				//	fmt.Printf("Downloading file from %s\n", url)
				//	out, _ := os.Create(fmt.Sprintf("%s.mp4", fileId))
				//	defer out.Close()
				//
				//	/*Right now using https://embedwistia-a.akamaihd.net/deliveries/<file_id>/file.mp4.
				//	An alternative: fmt.Sprintf("https://embed-ssl.wistia.com/deliveries/%s/file.mp4", url)*/
				//	resp, _ := http.Get(url)
				//	defer resp.Body.Close()
				//
				//	n, _ := io.Copy(out, resp.Body)
				//	fmt.Printf("Bytes copied: %d\n", n)
				//}(url)
			}
		}(lessonUrl)
	}
	wg.Wait()

	fmt.Printf("Total lessonUrls downloaded: %d\n", len(files))
}
