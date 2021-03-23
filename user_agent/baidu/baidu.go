package baidu

type BaiduBot struct {
}

func (ua BaiduBot) Header() (value string) {
	return "Mozilla/5.0 (compatible; Baiduspider/2.0; +http://www.baidu.com/search/spider.html)"
}
