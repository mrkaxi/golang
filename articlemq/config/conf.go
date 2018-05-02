package conf

const (
	//RunMode 0 测试 1生产
	RunMode = 1
	//MqConnectNUm 协程数
	MqConnectNUm = 10
	//GoPrefetchCount 单协程消息处理数
	GoPrefetchCount = 5
	//DownloadSRC 下载图片临时目录
	DownloadSRC = "./temp"
	//Mqurl 消息队列连接字符串
	Mqurl = "amqp://admin:U2FsdGVkX18Xzrgc@13.229.250.204:5672"
	//UploadURL API接口上传地址
	UploadURL = "http://13.229.126.123/api/pic-upload?url="
	//ContentURLPrefix 内容图片缩放前缀地址
	ContentURLPrefix = "http://prsize.allviki.com/resize_500x284/"
	//DefaultCoverPic 默认文章封面图
	DefaultCoverPic = "2017/12-08/7166925a0a81483800fd390afd9508f7.jpg"
	//ContentIgnorePic 内容忽略图片
	ContentIgnorePic = "/pagespeed_static/1.JiBnMqyl6S.gif|CB170629_1.png|/image/gobiz.png"
	//ConvertS2TWFile 简转繁文件地址
	ConvertS2TWFile = "/usr/share/opencc/s2tw.json"
	//ProxyHost 代理地址
	ProxyHost = "http://172.31.19.30:5910"
)

//ArticleURL 获取远程文章接口
func ArticleURL() (url string) {
	if RunMode == 1 {
		url = "http://epg-380576188.ap-southeast-1.elb.amazonaws.com:8090/api/member/toutiaoMq"
	} else {
		url = "http://127.0.0.1:8082/api/member/toutiaoMq"
	}
	return url
}
