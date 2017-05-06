package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/gosuri/uiprogress"
	"golang.org/x/net/html"
)

var (
	// TODO: use os.Args[1:] or, better yet, flags
	courseUrl             = "https://egghead.io/courses/introduction-to-reactive-programming"
	//courseUrl             = "https://egghead.io/courses/learn-the-best-and-most-useful-scss"
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

	nodeWithFileUrlFound, nodeWithFileNameFound := false, false
	fileUrl, fileName := "", ""

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
							nodeWithFileUrlFound = true
							return
						}
					}
				}
				if a.Key == "itemprop" && a.Val == "name" {
					//fmt.Println(a)
					//fileName = a.Val
					for _, a := range n.Attr {
						if a.Key == "content" && strings.Index(a.Val, ".mp4") == -1 {
							//fmt.Println(a)
							fileName = fmt.Sprintf("%d. %s", l.order, a.Val)
							//fmt.Println(fileName)
							nodeWithFileNameFound = true
							return
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if nodeWithFileUrlFound && nodeWithFileNameFound {
				break
			}
			f(c)
		}
	}
	f(doc)

	//fmt.Println("fileUrl: ", fileUrl)
	//fmt.Println("fileName: ", fileName)

	return file{
		url:  fileUrl,
		name: fileName,
	}

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

// ProxyReader is inspired by: https://github.com/cheggaaa/pb
// Also see: https://stackoverflow.com/questions/22421375/how-to-print-the-bytes-while-the-file-is-being-downloaded-golang
type ProxyReader struct {
	io.Reader
	bar   *uiprogress.Bar
	total int
}

func (pr *ProxyReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)

	//serviceMu.Lock()
	//defer serviceMu.Unlock()
	//pr.bar.Total += n
	pr.total += n
	pr.bar.Set(pr.total)

	//if err == nil {
	//	fmt.Printf("Read %d bytes for a total of %d; %s complete\n", n, pr.bar.Total, pr.bar.CompletedPercentString())
	//}

	return n, err
}

// TODO: Check with https://golang.org/doc/articles/race_detector.html
func main() {
	// Pattern taken from: https://tour.golang.org/concurrency/4
	ls := make(chan lesson)
	go getLessonUrls(courseUrl, ls)

	lessonsDownloaded := []string{}

	var wg sync.WaitGroup
	uiprogress.Start()

	for l := range ls {
		wg.Add(1)
		go func(l lesson) {
			defer wg.Done()

			f := getFile(l)

			//fmt.Printf("Downloading file from %s\n", f.url)
			fileName := fmt.Sprintf("%s.mp4", f.name)
			out, _ := os.Create(fileName)
			defer out.Close()

			/*Right now using https://embedwistia-a.akamaihd.net/deliveries/<file_id>/file.mp4.
			An alternative: fmt.Sprintf("https://embed-ssl.wistia.com/deliveries/%s/file.mp4", url)*/
			resp, err := http.Get(f.url)

			if err != nil {
				//fmt.Printf("Error: %s. Skipping %s\n", err.Error(), fileName)
				os.Remove(fileName)
				return
			}
			defer resp.Body.Close()

			fmt.Printf("Lesson %d, length is %d MB\n", l.order, resp.ContentLength/1000000) // 1048576)

			bar := uiprogress.AddBar(int(resp.ContentLength)).AppendCompleted().PrependElapsed()
			bar.PrependFunc(func(b *uiprogress.Bar) string {
				return fmt.Sprintf("Lesson %d: %d our of %d", l.order, b.Current(), b.Total)
			})
			proxyReader := &ProxyReader{Reader: resp.Body, bar: bar}

			_, err = io.Copy(out, proxyReader)
			if err != nil {
				//fmt.Printf("Error: %s. Skipping %s\n", err.Error(), fileName)
				os.Remove(fileName)
				return
			}
			// TODO: track and display file sizes in MBs (maybe display total as well).

			atomic.AddInt32(&downloadCounter, 1)

			serviceMu.Lock()
			defer serviceMu.Unlock()
			lessonsDownloaded = append(lessonsDownloaded, fileName)
		}(l)
	}
	wg.Wait()

	uiprogress.Stop()

	// TODO: sort lessonsDownloaded before printing.
	fmt.Printf("\nSuccessfully downloaded %d lessons: %s\n", downloadCounter, lessonsDownloaded)
}
