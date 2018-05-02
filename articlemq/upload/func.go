//Package upload 上传方法
package upload

import (
	"articlemq/common"
	config "articlemq/config"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/tidwall/gjson"
)

//PublicUploadPic 图片上传
func PublicUploadPic(httpPic string, root bool) (localPic string) {
	if runtime.GOOS == "linux" {
		localPic = AWSS3uploadPic(httpPic, root)
	} else {
		localPic = APIUploadPic(httpPic, root)
	}
	return
}

//APIUploadPic 接口上传
func APIUploadPic(httpPic string, root bool) (localPic string) {
	httpPic = url.QueryEscape(httpPic)
	resp, err := http.Get(config.UploadURL + httpPic)
	if err == nil {
		defer resp.Body.Close()
		body, err2 := ioutil.ReadAll(resp.Body)
		if err2 == nil {
			json := string(body)
			code := gjson.Get(json, "code").String()
			if code == "0" {
				if root == true {
					localPic = gjson.Get(json, "data").String()
				} else {
					localPic = config.ContentURLPrefix + gjson.Get(json, "data").String()
				}
				fmt.Println("http api upload ok!")
			}
		}
	} else {
		fmt.Println(err)
	}
	return
}

//CreateDir 生成目录
func CreateDir(src string) {
	_ = os.MkdirAll(src, 0777)
}

//AWSS3uploadPic AWSs3上传图片
func AWSS3uploadPic(httpPic string, root bool) (localPic string) {
	proxy := func(_ *http.Request) (*url.URL, error) {
		return url.Parse(config.ProxyHost)
	}
	dial := func(netw, addr string) (net.Conn, error) {
		c, err := net.DialTimeout(netw, addr, time.Second*20) //设置建立连接超时20S
		if err != nil {
			return nil, err
		}
		c.SetDeadline(time.Now().Add(30 * time.Second)) //设置发送接收数据超时30S
		return c, nil
	}
	client := &http.Client{
		Transport: &http.Transport{
			Dial: dial,
		},
	}
	if strings.Contains(httpPic, "media.orientaldaily.com.my") {
		client = &http.Client{
			Transport: &http.Transport{
				Dial:  dial,
				Proxy: proxy,
			},
		}
	}

	picDownloadType := "GET"
	if strings.Contains(httpPic, "kwongwah.com") {
		picDownloadType = "POST"
	}
	req, _ := http.NewRequest(picDownloadType, httpPic, nil)
	resp, errDownload := client.Do(req)
	if errDownload != nil {
		fmt.Println("------无法下载-------")
		localPic = "-1"
	} else {
		if resp.ContentLength < 150 {
			fmt.Println("------原图为空图或预加载图-------")
			localPic = "-1"
		} else {
			pix, _ := ioutil.ReadAll(resp.Body)
			defer resp.Body.Close()
			contentType := http.DetectContentType(pix)
			var contentExt string
			if contentType == "image/gif" {
				contentExt = ".gif"
			} else {
				contentExt = ".jpg"
			}
			filename := common.GetTimestamp() + contentExt
			localSRC := config.DownloadSRC + "/" + filename
			file, _ := os.Create(localSRC)

			_, _ = file.Write(pix)
			defer file.Close()
			s3SRC := time.Unix(time.Now().Unix(), 0).Format("2006") + "/" + time.Unix(time.Now().Unix(), 0).Format("1") + "/"

			shell := "aws s3 cp " + localSRC + " s3://dabo-pictures/" + s3SRC + " 2>&1"
			_, err := exec.Command("bash", "-c", shell, "./").Output()
			if err == nil {
				if root == true {
					localPic = s3SRC + filename
				} else {
					localPic = config.ContentURLPrefix + s3SRC + filename
				}
				fmt.Println("AWS S3 upload ok!")
			}
			os.Remove(localSRC)
		}
	}
	return
}
