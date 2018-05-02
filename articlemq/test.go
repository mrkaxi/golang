package main

import (
	html "articlemq/html"
	"fmt"
	"reflect"
	"time"
)

func main() {
	urlString := "http://www.mstar.com.my/~/media/590ee2bdf7a24189b006fa9f2ff5a21f.ashx?la=en"
	html.EraseQueryString(&urlString)
	current := time.Now().Unix()
	fmt.Println(urlString)
	fmt.Println(reflect.TypeOf(current))
	// newCode, _ := json.Marshal(article)
	// code = string(newCode)
	// baseCode := base64.StdEncoding.EncodeToString([]byte(code))
	// for true {
	// 	resp, err := http.Post(config.ArticleURL(),
	// 		"application/x-www-form-urlencoded",
	// 		strings.NewReader("code="+baseCode))
	// 	if err == nil {
	// 		defer resp.Body.Close()
	// 		body, err := ioutil.ReadAll(resp.Body)
	// 		if err == nil {
	// 			json := string(body)
	// 			status := gjson.Get(json, "result.status")
	// 			importCode = status.String()
	// 			if importCode == "0" || importCode == "-3" {
	// 				fmt.Println(json)
	// 				break
	// 			}
	// 		}
	// 	}
	// }
}
