//Package html HTML相关方法
package html

import (
	config "articlemq/config"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"net/url"
	"regexp"
	"strings"
)

//Substr 截取字符串 start 起点下标 length 需要截取的长度
func Substr(str string, start int, length int) string {
	rs := []rune(str)
	rl := len(rs)
	end := 0

	if start < 0 {
		start = rl - 1 + start
	}
	end = start + length

	if start > end {
		start, end = end, start
	}

	if start < 0 {
		start = 0
	}
	if start > rl {
		start = rl
	}
	if end < 0 {
		end = 0
	}
	if end > rl {
		end = rl
	}
	return string(rs[start:end])
}

//Unicode2string unicode转字符
func Unicode2string(form string) (to string, err error) {
	bs, err := hex.DecodeString(strings.Replace(form, `\u`, ``, -1))
	if err != nil {
		return
	}
	for i, bl, br, r := 0, len(bs), bytes.NewReader(bs), uint16(0); i < bl; i += 2 {
		binary.Read(br, binary.BigEndian, &r)
		to += string(r)
	}
	return
}

//StrHTMLEntityDecode html转义
func StrHTMLEntityDecode(encode string) string {
	encode = strings.Replace(encode, "&lt;", "<", -1)
	encode = strings.Replace(encode, "&gt;", ">", -1)
	encode = strings.Replace(encode, "&#x3D;", "=", -1)
	encode = strings.Replace(encode, "&quot;", "\"", -1)
	return encode
}

//ContentPicToJSON 提取html代码中图片src，生成json
func ContentPicToJSON(content string, contentImgRex string) (picdata string) {
	picdatas := []string{}
	valid := regexp.MustCompile(`<img[\S\s]*?>`)
	contentImgs := valid.FindAllString(content, -1)
	for _, contentImg := range contentImgs {
		valid = regexp.MustCompile(contentImgRex + "=\"(.+?)\"")
		contentImgSrcs := valid.FindAllString(contentImg, -1)
		for _, contentImgSrc := range contentImgSrcs {
			pic := strings.Replace(contentImgSrc, contentImgRex+"=", "", -1)
			pic = strings.Replace(pic, "\"", "", -1)
			if strings.Contains(pic, config.ContentURLPrefix) {
				pic = strings.Replace(pic, config.ContentURLPrefix, "", -1)
				picdatas = append(picdatas, pic)
			}
		}
	}
	picdata = ""
	if len(picdatas) > 0 {
		json, err := json.Marshal(picdatas)
		if err == nil {
			picdata = string(json)
		}
	}
	return
}

//FilterStr 过滤html中js和json
func FilterStr(html *string) {
	// re, _ := regexp.Compile("\\<[\\S\\s]+?\\>")
	// *html = re.ReplaceAllStringFunc(*html, strings.ToLower)
	re := regexp.MustCompile(`<div id="feelbox-widget" class="voted small">[\S\s]*</div>`)
	*html = re.ReplaceAllString(*html, "")

	re = regexp.MustCompile(`<div class="like_post"[\S\s]*</div>`)
	*html = re.ReplaceAllString(*html, "")

	re = regexp.MustCompile(`<div class="paszone_container[\S\s]*</div>`)
	*html = re.ReplaceAllString(*html, "")

	re = regexp.MustCompile(`{[\S\s]+?}|【[\S\s]*?】|<a[\S\s]*?>|<input[\S\s]*?>|<!--[^>]+>|<h1[\S\s]*?</h1>|<h2[\S\s]*?</h2>|<h3[\S\s]*?</h3>|<h4[\S\s]*?</h4>|<ins[\S\s]+?</ins>|<style[\S\s]*?</style>|<iframe[\S\s]*?</iframe>|<iframe[\S\s]*?>|<script[\S\s]*?</script>|style="[\S\s]*?"|width="[\S\s]*?"|alt=""|rel="[\S\s]*?"|srcset="[\S\s]*?"|class="[\S\s]*?"|height="[\S\s]*?"|sizes="[\S\s]*?"|id="[\S\s]*?"|onclick="[\S\s]*?"`)
	*html = re.ReplaceAllString(*html, "")
	re = regexp.MustCompile(`\s+`)
	*html = re.ReplaceAllString(*html, " ")
	*html = strings.Replace(*html, " >", ">", -1)
	*html = strings.Replace(*html, "<span class=\"td-adspot-title\">Advertisement</span>", "", -1)
	*html = strings.Replace(*html, "<p><br></p>", "", -1)
	*html = strings.Replace(*html, "[", "", -1)
	*html = strings.Replace(*html, "]", "", -1)
	*html = strings.Replace(*html, "\n", "", -1)
	*html = strings.Replace(*html, "<div></div>", "", -1)
	*html = strings.Replace(*html, "Advertisement", "", -1)
	*html = strings.Replace(*html, "Gobiz Sponsored", "", -1)
	*html = strings.Replace(*html, "<span></span>", "", -1)
	*html = strings.Replace(*html, "</a>", "", -1)
	*html = strings.Replace(*html, "Sponsored Links", "", -1)
}

//EraseQueryString 过滤html中js和json
func EraseQueryString(urlString *string) {
	if strings.Contains(*urlString, ".php") == false {
		u, _ := url.Parse(*urlString)
		if u.RawQuery != "" {
			*urlString = strings.Replace(*urlString, "?"+u.RawQuery, "", 1)
		}
	}

}
