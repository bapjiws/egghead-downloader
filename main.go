package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"

	"golang.org/x/net/html"
)

var (
	// TODO: use os.Args[1:] or, better yet, flags
	courseUrl             = "https://egghead.io/courses/learn-the-best-and-most-useful-scss" // https://egghead.io/courses/introduction-to-reactive-programming
	downloadCounter int32 = 0
	serviceMu       sync.Mutex
)

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
		name: getFileName(l.order, doc),
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

func getFileName(order int, doc *html.Node) string {
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

	return fmt.Sprintf("%d. %s", order, fileName)
}

type lesson struct {
	url   string
	order int
}

func getLessonUrls(courseUrl string, lessons chan lesson) {
	lessonIds := map[string]bool{}
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
					lessons <- lesson{url: a.Val, order: lessonOrder}
					lessonOrder++
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	close(lessons)
}

type file struct {
	url  string
	name string
}

// TODO: Check with https://golang.org/doc/articles/race_detector.html
func main() {
	// Pattern taken from: https://tour.golang.org/concurrency/4
	ls := make(chan lesson)
	go getLessonUrls(courseUrl, ls)

	lessonsDownloaded := []string{}

	var wg sync.WaitGroup

	for l := range ls {
		wg.Add(1)
		go func(l lesson) {
			defer wg.Done()

			f := getFile(l)
			//fmt.Println(f)

			wg.Add(1)
			// TODO: use progress bar to visualize downloading? Check this one out: https://github.com/cheggaaa/pb.
			go func(f file) {
				defer wg.Done()
				//fmt.Printf("Downloading file from %s\n", f.url)
				fileName := fmt.Sprintf("%s.mp4", f.name)
				out, _ := os.Create(fileName)
				defer out.Close()

				/*Right now using https://embedwistia-a.akamaihd.net/deliveries/<file_id>/file.mp4.
				An alternative: fmt.Sprintf("https://embed-ssl.wistia.com/deliveries/%s/file.mp4", url)*/
				resp, err := http.Get(f.url)
				if err != nil {
					fmt.Printf("Error: %s. Skipping %s\n", err.Error(), fileName)
					os.Remove(fileName)
					return
				}
				defer resp.Body.Close()

				_, err = io.Copy(out, resp.Body)
				if err != nil {
					fmt.Printf("Error: %s. Skipping %s\n", err.Error(), fileName)
					os.Remove(fileName)
					return
				}
				// TODO: track and display file sizes in MBs (maybe display total as well).
				//fmt.Printf("Byte s copied: %d\n", n)

				atomic.AddInt32(&downloadCounter, 1)
				serviceMu.Lock()
				defer serviceMu.Unlock()
				lessonsDownloaded = append(lessonsDownloaded, fileName)
			}(f)
		}(l)
	}
	wg.Wait()

	// TODO: sort lessonsDownloaded beforehand and print 'em out using a for loop.
	fmt.Printf("\nSuccessfully downloaded %d lessons: %s\n", downloadCounter, lessonsDownloaded)
}
