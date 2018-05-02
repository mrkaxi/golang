package main

import (
	common "articlemq/common"
	config "articlemq/config"
	html "articlemq/html"
	model "articlemq/model"
	upload "articlemq/upload"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/stevenyao/go-opencc"
	"github.com/streadway/amqp"
)

var conn *amqp.Connection
var channel *amqp.Channel

const (
	queueName = "articles_mq"
)

//导入数据
func importToMysql(code string, i int) (importCode int, msg string) {
	//json str 转struct
	var article model.Article
	json.Unmarshal([]byte(code), &article)

	if article.Spider_source_url == "" {
		return 0, ""
	}
	if strings.Contains(article.Image, "http") == false {
		article.Image = ""
	}
	//过滤标题
	html.FilterStr(&article.Title)
	//转换繁体标题
	convert := opencc.NewConverter(config.ConvertS2TWFile)
	article.Title_T = convert.Convert(article.Title)

	//上传封面图
	if article.Image != "" {
		article.Image = upload.PublicUploadPic(common.GetHTTPImg(article.Image, false), true)
		fmt.Println(strconv.Itoa(i) + ":" + article.Image)
	}

	//contenthtml转义并上传外链图片
	article.Content = html.StrHTMLEntityDecode(article.Content)
	var contentImgRex string
	u, _ := url.Parse(article.Link)
	//获取源HOST
	contentImgHost := u.Scheme + "://" + u.Host + "/"
	switch article.Spider_source {
	case "goody25":
		html.FilterStr(&article.Content)
		contentImgRex = "attr"
	case "orientaldaily":
		html.FilterStr(&article.Content)
		article.Content = strings.Replace(article.Content, "data-original", "src", -1)
		contentImgRex = "src"
	case "chinapress":
		html.FilterStr(&article.Content)
		contentImgRex = "src"
	default:
		contentImgRex = "src"
	}

	//转存content中外站图片-开始
	valid := regexp.MustCompile(`<img[\S\s]*?>`)
	contentImgs := valid.FindAllString(article.Content, -1)
	//fmt.Println(contentImgs)
	for _, contentImg := range contentImgs {
		valid = regexp.MustCompile(contentImgRex + "=\"(.+?)\"")
		contentImgSrcs := valid.FindAllString(contentImg, -1)
		for _, contentImgSrc := range contentImgSrcs {
			localPic := strings.Replace(contentImgSrc, contentImgRex+"=", "", -1)
			localPic = strings.Replace(localPic, "\"", "", -1)
			if strings.Contains(config.ContentIgnorePic, localPic) {
				continue
			} else {
				if html.Substr(localPic, 0, 2) == "//" {
					localPic = "http:" + localPic
				}
				if html.Substr(localPic, 0, 4) != "http" {
					localPic = contentImgHost + localPic
				}
				fmt.Println(localPic)
				html.EraseQueryString(&localPic)
				fmt.Println(common.GetHTTPImg(localPic, false))
				localPic = upload.PublicUploadPic(common.GetHTTPImg(localPic, false), false)
				//localPic = upload.PublicUploadPic(localPic, false)
				//无封面图取内容第一张图
				if article.Image == "" || article.Image == "-1" {
					article.Image = strings.Replace(localPic, config.ContentURLPrefix, "", 1)
				}
				//下载图片超时，删除此图
				if localPic == "-1" {
					article.Content = strings.Replace(article.Content, contentImg, "", 1)
				} else {
					article.Content = strings.Replace(article.Content, contentImg, "<img src=\""+localPic+"\">", 1)
					fmt.Println(strconv.Itoa(i) + ":" + localPic)
				}
			}
		}
	}
	//转存content中外站图片-结束

	//无封面图，内容无图，给予默认图
	if article.Image == "" || article.Image == "-1" {
		article.Image = config.DefaultCoverPic
	}

	//过滤script、json等非标准冗余字符
	html.FilterStr(&article.Content)
	//转换繁体内容
	article.Content_T = convert.Convert(article.Content)
	// fmt.Println(article.Content)
	// runtime.Breakpoint()
	defer convert.Close()
	//提交入库
	if config.RunMode == 0 {
		importCode = 0
		common.TraceTempfile(article.Content)
	} else {
		importCode, msg = article.InsertArticle()
	}
	return
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
	conn, err = amqp.Dial(config.Mqurl)
	failOnErr(err, "failed to connect tp rabbitmq")

	channel, err = conn.Channel()
	failOnErr(err, "failed to open a channel")

	err = channel.Qos(
		config.GoPrefetchCount, // prefetch count
		0,     // prefetch size
		false, // global
	)
	failOnErr(err, "Failed to set QoS")

	msgs, err := channel.Consume(queueName, "", false, false, false, false, nil)
	failOnErr(err, "")

	for d := range msgs {
		s := common.BytesToString(&(d.Body))
		common.Tracefile(*s)
		f, m := importToMysql(*s, i)
		fmt.Println("入库的回调状态码:" + strconv.Itoa(f))
		if f >= 0 && config.RunMode == 1 {
			d.Ack(false)
		} else {
			fmt.Println(m)
		}
	}
}

func close() {
	channel.Close()
	conn.Close()
}

func main() {
	upload.CreateDir(config.DownloadSRC)
	runtime.GOMAXPROCS(runtime.NumCPU())

	c := make(chan bool)
	for i := 0; i < config.MqConnectNUm; i++ {
		go mqConnect(i)
	}
	<-c
}
