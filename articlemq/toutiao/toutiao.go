package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/streadway/amqp"
	"github.com/tidwall/gjson"
)

var conn *amqp.Connection
var channel *amqp.Channel

const (
	//协程数
	mqConnectNUm = 5
	//单协程消息处理数
	goPrefetchCount = 10
	//下载图片临时目录
	downloadSRC      = "./temp"
	queueName        = "toutiao_channel"
	mqurl            = "amqp://admin:U2FsdGVkX18Xzrgc@13.229.126.123:5672"
	uploadURL        = "http://13.229.126.123/api/pic-upload?url="
	contentURLPrefix = "http://prsize.allviki.com/resize_500x284/"
	//articleURL       = "http://127.0.0.1:8082/api/member/toutiaoMq"
	articleURL = "http://epg-380576188.ap-southeast-1.elb.amazonaws.com:8090/api/member/toutiaoMq"
)

//记录日志
func tracefile(strContent string) {
	fd, _ := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	fdContent := strings.Join([]string{strContent, "\r\n"}, "")
	buf := []byte(fdContent)
	fd.Write(buf)
	fd.Close()
}

//图片上传
func uploadPic(httpPic string, root bool) (localPic string) {
	if runtime.GOOS == "linux" {
		localPic = s3uploadPic(httpPic, root)
	} else {
		localPic = apiUploadPic(httpPic, root)
	}
	return
}

//接口上传
func apiUploadPic(httpPic string, root bool) (localPic string) {
	for true {
		httpPic = url.QueryEscape(httpPic)
		resp, err := http.Get(uploadURL + httpPic)
		if err == nil {
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err == nil {
				json := string(body)
				code := gjson.Get(json, "code").String()
				if code == "0" {
					if root == true {
						localPic = gjson.Get(json, "data").String()
					} else {
						localPic = contentURLPrefix + gjson.Get(json, "data").String()
					}
					fmt.Println("http upload ok!")
					break
				}
			}
		}
	}
	return
}

func createDir() {
	_ = os.MkdirAll(downloadSRC, 0777)
}

//s3上传图片
func s3uploadPic(httpPic string, root bool) (localPic string) {
	for true {
		resp, _ := http.Get(httpPic)
		pix, _ := ioutil.ReadAll(resp.Body)
		contentType := http.DetectContentType(pix)
		var contentExt string
		if contentType == "image/gif" {
			contentExt = ".gif"
		} else {
			contentExt = ".jpg"
		}
		filename := getTimestamp() + contentExt
		localSRC := downloadSRC + "/" + filename
		file, _ := os.Create(localSRC)
		_, _ = file.Write(pix)
		file.Close()
		s3SRC := time.Unix(time.Now().Unix(), 0).Format("2006") + "/" + time.Unix(time.Now().Unix(), 0).Format("1") + "/"

		shell := "aws s3 cp " + localSRC + " s3://dabo-pictures/" + s3SRC + " 2>&1"
		_, err := exec.Command("bash", "-c", shell, "./").Output()
		if err == nil {
			if root == true {
				localPic = s3SRC + filename
			} else {
				localPic = contentURLPrefix + s3SRC + filename
			}
			fmt.Println("AWS S3 upload ok!")
			os.Remove(localSRC)
			break
		}
		os.Remove(localSRC)
	}
	return
}

//html转义
func htmlEntityDecode(encode string) string {
	encode = strings.Replace(encode, "&lt;", "<", -1)
	encode = strings.Replace(encode, "&gt;", ">", -1)
	encode = strings.Replace(encode, "&#x3D;", "=", -1)
	encode = strings.Replace(encode, "&quot;", "\"", -1)
	return encode
}

//过滤html中js和json
func filterStr(html *string) {
	re, _ := regexp.Compile("\\<[\\S\\s]+?\\>")
	*html = re.ReplaceAllStringFunc(*html, strings.ToLower)
	re, _ = regexp.Compile("\\<script[\\S\\s]+?\\</script\\>")
	*html = re.ReplaceAllString(*html, "")
	re, _ = regexp.Compile("\\{[\\S\\s]+?\\}")
	*html = re.ReplaceAllString(*html, "\n")
}

/**
 * 数组去重 去空
 */
func removeDuplicatesAndEmpty(a []string) (ret []string) {
	aLen := len(a)
	for i := 0; i < aLen; i++ {
		if len(a[i]) == 0 {
			continue
		}
		ret = append(ret, a[i])
	}
	return
}

/**
 * 修正获取头条图片链接
 */
func getToutiaoImg(url string, large bool) (imageURL string) {
	url = strings.Replace(url, "http:", "", -1)
	url = strings.Replace(url, "https:", "", -1)
	urlArray := strings.Split(url, "/")
	urlArray = removeDuplicatesAndEmpty(urlArray)
	if large == true {
		for k, v := range urlArray {
			if k == 0 {
				imageURL += "http://" + v + "/large/"
			}
			if k == len(urlArray)-1 {
				imageURL += v
			}
		}
	} else {
		for k, v := range urlArray {
			if k == 0 {
				imageURL += "http://"
			}
			imageURL += v
			if k != len(urlArray)-1 {
				imageURL += "/"
			}
		}
	}
	return
}

func getTimestamp() (timestamp string) {
	t := time.Now()
	timestamp = strconv.FormatInt(t.UTC().UnixNano(), 10)
	timestamp = timestamp[:13]
	return
}

//Article json结构体
type Article struct {
	ID                string `json:"id"`
	Title             string `json:"title"`
	Spider_source_url string `json:"spider_source_url"`
	Abstract          string `json:"abstract"`
	Tag               string `json:"tag"`
	Chinese_tag       string `json:"chinese_tag"`
	Comments_count    int    `json:"comments_count"`
	Go_detail_count   int    `json:"go_detail_count"`
	Image_url         string `json:"image_url"`
	Image             string `json:"image"`
	Article_url       string `json:"article_url"`
	Content           string `json:"content"`
}

//导入数据
func importToMysql(code string, i int) bool {
	//json str 转struct
	var article Article
	json.Unmarshal([]byte(code), &article)
	if article.Image_url != "" {
		article.Image = article.Image_url
	}
	if article.Image == "" || article.Content == "" {
		return true
	}
	//上传封面图
	article.Image = uploadPic(getToutiaoImg(article.Image, true), true)
	fmt.Println(strconv.Itoa(i) + ":" + article.Image)
	article.Content = strings.Replace(article.Content, " ", "+", -1)
	article.Content = strings.Replace(article.Content, "\n", "", -1)
	u, _ := base64.StdEncoding.DecodeString(article.Content)

	article.Content = htmlEntityDecode(string(u))
	//去除script及json字符
	filterStr(&article.Content)
	var valid = regexp.MustCompile("src=\"(.+?)\"")
	toutiaoPics := valid.FindAllString(article.Content, -1)
	for _, toutiaoPic := range toutiaoPics {
		localPic := strings.Replace(toutiaoPic, "src=", "", -1)
		localPic = strings.Replace(localPic, "\"", "", -1)
		localPic = uploadPic(getToutiaoImg(localPic, false), false)
		article.Content = strings.Replace(article.Content, toutiaoPic, "src=\""+localPic+"\"", 1)
		fmt.Println(strconv.Itoa(i) + ":" + localPic)
	}
	newCode, _ := json.Marshal(article)
	code = string(newCode)
	baseCode := base64.StdEncoding.EncodeToString([]byte(code))
	for true {
		resp, err := http.Post(articleURL,
			"application/x-www-form-urlencoded",
			strings.NewReader("code="+baseCode))
		if err == nil {
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err == nil {
				json := string(body)
				fmt.Println(json)
				status := gjson.Get(json, "result.status")
				if status.String() == "0" {
					break
				}
			}
		}
	}
	return true
}

//错误处理
func failOnErr(err error, msg string) {
	if err != nil {
		log.Fatalf("%s:%s", msg, err)
		panic(fmt.Sprintf("%s:%s", msg, err))
	}
}

//队列处理
func mqConnect(i int) {
	var err error
	conn, err = amqp.Dial(mqurl)
	failOnErr(err, "failed to connect tp rabbitmq")

	channel, err = conn.Channel()
	failOnErr(err, "failed to open a channel")

	err = channel.Qos(
		goPrefetchCount, // prefetch count
		0,               // prefetch size
		false,           // global
	)
	failOnErr(err, "Failed to set QoS")

	msgs, err := channel.Consume(queueName, "", false, false, false, false, nil)
	failOnErr(err, "")

	for d := range msgs {
		s := bytesToString(&(d.Body))
		tracefile(*s)
		f := importToMysql(*s, i)
		fmt.Println(f)
		if f == true {
			d.Ack(false)
		}
	}
}

func close() {
	channel.Close()
	conn.Close()
}

//二进制字节转换字符串
func bytesToString(b *[]byte) *string {
	s := bytes.NewBuffer(*b)
	r := s.String()
	return &r
}

func main() {
	createDir()
	runtime.GOMAXPROCS(runtime.NumCPU())

	c := make(chan bool)
	for i := 0; i < mqConnectNUm; i++ {
		go mqConnect(i)
	}
	<-c
}
