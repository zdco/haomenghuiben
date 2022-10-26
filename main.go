package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"

	"golang.org/x/net/html/charset"
	"golang.org/x/text/transform"
)

var header = map[string]string{
	"Host":                      "mp.weixin.qq.com",
	"Connection":                "keep-alive",
	"Cache-Control":             "max-age=0",
	"Upgrade-Insecure-Requests": "1",
	"User-Agent":                "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/53.0.2785.143 Safari/537.36",
	"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
}

var downloadPrev = "https://res.wx.qq.com/voice/getvoice?mediaid="
var downloadDir = "media"

var urlMap = make(map[string]bool) // New empty set

func determineEncodings(r io.Reader) []byte {
	OldReader := bufio.NewReader(r)
	bytes, err := OldReader.Peek(1024)
	if err != nil {
		panic(err)
	}
	e, _, _ := charset.DetermineEncoding(bytes, "")
	reader := transform.NewReader(OldReader, e.NewDecoder())
	all, err := ioutil.ReadAll(reader)
	if err != nil {
		panic(err)
	}
	return all
}

type StructEvent struct {
	Key string `json:"key"`
}

// 定义新的数据类型
type Spider struct {
	url    string
	header map[string]string
}

// 定义 Spider get的方法
func (keyword Spider) get_html_header() string {
	client := &http.Client{}
	req, err := http.NewRequest("GET", keyword.url, nil)
	if err != nil {
		panic(err)
	}
	for key, value := range keyword.header {
		req.Header.Add(key, value)
	}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	return string(determineEncodings(resp.Body))
}

func getUrls() []string {
	url := "https://mp.weixin.qq.com/s/NOAXaRaMysYV4MwUlulxsQ"
	spider := &Spider{url, header}
	html := spider.get_html_header()

	//链接
	pattern := `href="(.*?)"`
	rp := regexp.MustCompile(pattern)
	find_txt := rp.FindAllStringSubmatch(html, -1)

	urlIndex := 0
	var urls []string
	for i := 0; i < len(find_txt); i++ {
		url := find_txt[i][1]
		if strings.Contains(url, "mp.weixin.qq.com") {
			urlIndex++
			if _, ok := urlMap[url]; !ok {
				urls = append(urls, url)
				urlMap[url] = true
			} else {
				fmt.Printf("===== Duplicate data:%d =====\n", urlIndex)
			}
		}
	}

	return urls
}

func download(url string, to string) error {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	ioutil.WriteFile(downloadDir+"/"+to, data, 0777)
	return nil
}

func downloadAudio(url string) {
	spider := &Spider{url, header}
	html := spider.get_html_header()

	//ID
	pattern0 := `第(.*?)晚|第(.*?)：`
	rp0 := regexp.MustCompile(pattern0)
	find_txt0 := rp0.FindAllStringSubmatch(html, -1)

	//ID
	pattern1 := `voice_encode_fileid="(.*?)"`
	rp1 := regexp.MustCompile(pattern1)
	find_txt1 := rp1.FindAllStringSubmatch(html, -1)

	//名称
	pattern2 := `<mpvoice.* name="(.*?)"`
	rp2 := regexp.MustCompile(pattern2)
	find_txt2 := rp2.FindAllStringSubmatch(html, -1)

	if len(find_txt0) > 0 && len(find_txt1) > 0 && len(find_txt2) > 0 {
		fullUrl := downloadPrev + find_txt1[0][1]
		fileName := find_txt0[0][0] + "." + find_txt2[0][1] + ".mp3"
		fmt.Println(url, fullUrl, fileName)
		download(fullUrl, fileName)
	}
}

func main() {
	os.Mkdir(downloadDir, os.ModePerm)
	urls := getUrls()
	for i := 0; i < len(urls); i++ {
		downloadAudio(urls[i])
	}
}
