package model

import (
	"articlemq/common"
	config "articlemq/config"
	"articlemq/db"
	"articlemq/html"
	"net/url"
	"strings"
	"time"
)

//Article json结构体
type Article struct {
	Spider_source     string `json:"spider_source"`
	Publish_time      string `json:"publish_time"`
	Title             string `json:"title"`
	Title_T           string `json:"title_t"`
	Image             string `json:"image"`
	Content           string `json:"content"`
	Content_T         string `json:"content_t"`
	Spider_source_url string `json:"spider_source_url"`
	Spider_created_at string `json:"spider_created_at"`
	Link              string `json:"link"`
	Article_id        string `json:"article_id"`
}

//InsertArticle 写入article
func (article *Article) InsertArticle() (code int, msg string) {
	//查询爬虫源列表地址相关账号
	rows, err := db.SQLDB.Query(`SELECT a.uid,a.url,b.username,b.region,b.nickname,a.tags,a.copyrights FROM icms_carry_url a
		 INNER JOIN icms_member b on a.uid=b.uid WHERE a.url=? AND a.gid=2 AND a.url_status!=3`, url.QueryEscape(article.Spider_source_url))
	defer rows.Close()
	code = -2
	if err == nil {
		urls := make([]URL, 0)
		for rows.Next() {
			var url URL
			rows.Scan(&url.UID, &url.URL, &url.Username, &url.Region, &url.Nickname, &url.Tags, &url.Copyrights)
			urls = append(urls, url)
		}
		err = rows.Err()
		if err == nil && len(urls) > 0 {
			for _, v := range urls {
				//查询爬虫源详情地址是否已录入
				rows2 := db.SQLDB.QueryRow("SELECT count(id) FROM vw_article WHERE source_url=? AND userid=?", article.Link, v.UID)
				var articleCount int
				rows2.Scan(&articleCount)
				if articleCount > 0 {
					if article.Spider_source_url != "" {
						db.SQLDB.Exec("UPDATE icms_article set source=? WHERE source_url=?", article.Spider_source_url, article.Link)
						code = 1
					}
				} else {
					//台湾地区转换tag
					title := article.Title
					content := article.Content
					clink := common.GetRandomString(8)
					currentTime := time.Now().Unix()
					haspic := 0
					if article.Image != "" {
						haspic = 1
					}
					if v.Region == "10000011" {
						title = article.Title_T
						content = article.Content_T
						v.Tags = strings.Replace(v.Tags, "娱乐", "星聞", -1)
						v.Tags = strings.Replace(v.Tags, "科学", "科技", -1)
						v.Tags = strings.Replace(v.Tags, "历史", "文史", -1)
					}
					status := 3
					if v.Copyrights >= 7 && article.Image != config.DefaultCoverPic {
						status = 1
					}
					stmt, _ := db.SQLDB.Prepare(`INSERT INTO icms_article(cid,tpl,status,title,source,source_url,pic,haspic,tags,copyrights,
						clink,userid,region,username,editor,postime,updatetime,pubdate) VALUES (88796,'picture',?,?,?,?,?,?,?,?,?,?,?,
							?,?,?,?,?)`)
					defer stmt.Close()
					ret, _ := stmt.Exec(status, title, article.Spider_source_url, article.Link, article.Image, haspic, v.Tags, v.Copyrights,
						clink, v.UID, v.Region, v.Username, v.Nickname, currentTime, currentTime, currentTime)
					if LastInsertId, err := ret.LastInsertId(); nil == err {
						picdata := html.ContentPicToJSON(content, "src")
						stmt2, _ := db.SQLDB.Prepare(`INSERT INTO icms_article_data(aid,clink,description,picdata) VALUES
						 (?,?,?,?)`)
						defer stmt2.Close()
						_, err := stmt2.Exec(LastInsertId, clink, content, picdata)
						if err == nil {
							code = 0
						}
					}
				}
			}
		} else {
			code = -3
			msg = article.Spider_source_url
		}
	}
	return
}
