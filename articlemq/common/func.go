//Package common 通用方法
package common

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	randSeek = int64(1)
	l        sync.Mutex
)

//RemoveDuplicatesAndEmpty 数组去重去空
func RemoveDuplicatesAndEmpty(a []string) (ret []string) {
	aLen := len(a)
	for i := 0; i < aLen; i++ {
		if len(a[i]) == 0 {
			continue
		}
		ret = append(ret, a[i])
	}
	return
}

//GetHTTPImg 修正获取远程图片链接
func GetHTTPImg(url string, large bool) (imageURL string) {
	url = strings.Replace(url, "&amp;#038;", "&", -1)
	url = strings.Replace(url, "http:", "", -1)
	url = strings.Replace(url, "https:", "", -1)
	urlArray := strings.Split(url, "/")
	urlArray = RemoveDuplicatesAndEmpty(urlArray)
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
	imageURL = strings.Replace(imageURL, " ", "", -1)
	imageURL = strings.Replace(imageURL, "///", "//", -1)
	return
}

//GetTimestamp 获取毫秒级时间戳
func GetTimestamp() (timestamp string) {
	t := time.Now()
	timestamp = strconv.FormatInt(t.UTC().UnixNano(), 10)
	timestamp = timestamp[:13]
	return
}

//BytesToString 二进制字节转换字符串
func BytesToString(b *[]byte) *string {
	s := bytes.NewBuffer(*b)
	r := s.String()
	return &r
}

//TraceTempfile 记录临时日志
func TraceTempfile(strContent string) {
	fd, _ := os.OpenFile("./temp/"+GetTimestamp()+".html", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	fdContent := strings.Join([]string{strContent, "\r\n"}, "")
	buf := []byte(fdContent)
	fd.Write(buf)
	fd.Close()
}

//Tracefile 记录日志
func Tracefile(strContent string) {
	fd, _ := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	fdContent := strings.Join([]string{strContent, "\r\n"}, "")
	buf := []byte(fdContent)
	fd.Write(buf)
	fd.Close()
}

//UnicodeToString 记录日志
func UnicodeToString(textUnquoted string) (context string) {
	sUnicodev := strings.Split(textUnquoted, "\\u")
	for _, v := range sUnicodev {
		if len(v) < 1 {
			continue
		}
		temp, err := strconv.ParseInt(v, 16, 32)
		if err != nil {
			panic(err)
		}
		context += fmt.Sprintf("%c", temp)
	}
	return context
}

//GetRandomString 生成随机字符串
func GetRandomString(l int) string {
	str := "123456789abcdefghijklmnopqrstuvwxyz"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(getRandSeek()))
	for i := 0; i < l; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

func getRandSeek() int64 {
	l.Lock()
	if randSeek >= 100000000 {
		randSeek = 1
	}
	randSeek++
	l.Unlock()
	return time.Now().UnixNano() + randSeek
}
