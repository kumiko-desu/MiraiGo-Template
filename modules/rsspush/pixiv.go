package rsspush

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"sync"

	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/config"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/robfig/cron"
)

func init() {
	instance = &rsspushing{}
	bot.RegisterModule(instance)
}

type rsspushing struct {
}

// Rss rssXML
type Rss struct {
	XMLName xml.Name `xml:"rss"`
	Text    string   `xml:",chardata"`
	Atom    string   `xml:"atom,attr"`
	Version string   `xml:"version,attr"`
	Channel struct {
		Text  string `xml:",chardata"`
		Title string `xml:"title"`
		Link  struct {
			Text string `xml:",chardata"`
			Href string `xml:"href,attr"`
			Rel  string `xml:"rel,attr"`
			Type string `xml:"type,attr"`
		} `xml:"link"`
		Description   string `xml:"description"`
		Generator     string `xml:"generator"`
		WebMaster     string `xml:"webMaster"`
		Language      string `xml:"language"`
		LastBuildDate string `xml:"lastBuildDate"`
		Ttl           string `xml:"ttl"`
		Item          []struct {
			Text        string `xml:",chardata"`
			Title       string `xml:"title"`
			Description string `xml:"description"`
			PubDate     string `xml:"pubDate"`
			Guid        struct {
				Text        string `xml:",chardata"`
				IsPermaLink string `xml:"isPermaLink,attr"`
			} `xml:"guid"`
			Link   string `xml:"link"`
			Author string `xml:"author"`
		} `xml:"item"`
	} `xml:"channel"`
}

//RssInfo Rss + Lastestdate
type RssInfo struct {
	Rss         Rss
	LastestDate string
}

var on bool = false
var isFirst bool = true
var urls []string
var rsses = make(map[string]RssInfo)
var task = cron.New()
var managerQQ, sendGroupCode int64

func (m *rsspushing) MiraiGoModule() bot.ModuleInfo {
	return bot.ModuleInfo{
		ID:       "internal.rsspushing",
		Instance: instance,
	}
}

func (m *rsspushing) Init() {
	// 初始化过程
	// 在此处可以进行 Module 的初始化配置
	// 如配置读取
	managerQQ = config.GlobalConfig.GetInt64("managerQQ")
	urls = config.GlobalConfig.GetStringSlice("rss.pixiv")
	sendGroupCode = config.GlobalConfig.GetInt64("rss.sendGroupCode")

}

func (m *rsspushing) PostInit() {
	// 第二次初始化
	// 再次过程中可以进行跨Module的动作
	// 如通用数据库等等
}

func (m *rsspushing) Serve(b *bot.Bot) {
	// 注册服务函数部
	b.OnGroupMessage(func(c *client.QQClient, msg *message.GroupMessage) {
		if msg.Sender.Uin == managerQQ && msg.ToString() == "开启RSS" && on == false {
			on = true
			m := message.NewSendingMessage().Append(message.NewText("开始RSS推送"))
			c.SendGroupMessage(msg.GroupCode, m)
			if isFirst {
				for _, url := range urls {
					var url = url
					// 添加定时任务
					task.AddFunc("* */20 * * * * ", func() {
						update(url, c)
					})
				}
				isFirst = false
			}
			task.Start()
		}
		if msg.Sender.Uin == managerQQ && msg.ToString() == "关闭RSS" && on == true {
			on = false
			m := message.NewSendingMessage().Append(message.NewText("关闭RSS推送"))
			c.SendGroupMessage(msg.GroupCode, m)
			task.Stop()
		}
	})
}

func (m *rsspushing) Start(b *bot.Bot) {
	// 此函数会新开携程进行调用
	// ```go
	// 		go exampleModule.Start()
	// ```

	// 可以利用此部分进行后台操作
	// 如http服务器等等
	// 初次获取RSS
	for _, url := range urls {
		rsses[url] = getRssInfo(url)
	}
}

func (m *rsspushing) Stop(b *bot.Bot, wg *sync.WaitGroup) {
	// 别忘了解锁
	defer wg.Done()
	// 结束部分
	// 一般调用此函数时，程序接收到 os.Interrupt 信号
	// 即将退出
	// 在此处应该释放相应的资源或者对状态进行保存
}

var instance *rsspushing

var logger = utils.GetModuleLogger("internal.rsspushing")

func sendUpdate(c *client.QQClient, imageURLs []string, title string, author string, link string) {
	m := message.NewSendingMessage().Append(message.NewText("检测到RSS更新"))
	for _, iurl := range imageURLs {
		logger.Println("图片：", iurl)
		pimage, err := c.UploadGroupImage(sendGroupCode, bytes.NewReader(downloadImg(iurl)))
		if err != nil {
			logger.WithError(err).Errorf("图片上传失败")
		}
		fmt.Println(pimage)
		m.Append(pimage)
	}
	m.Append(message.NewText("标题：" + title + "\n作者：" + author + "\n链接：" + link))
	c.SendGroupMessage(sendGroupCode, m)
}

func update(url string, c *client.QQClient) {
	rssinfo := getRssInfo(url)
	// rssinfo.LastestDate = "Tue, 25 Aug 2020 03:31:34 GMT"
	if rssinfo.LastestDate != rsses[url].LastestDate {
		//检测到rss更新
		//具体操作
		for _, it := range rssinfo.Rss.Channel.Item {
			fmt.Println(it.PubDate, rsses[url].LastestDate)
			if it.PubDate == rsses[url].LastestDate {
				break
			}
			re := regexp.MustCompile(`https://[^"]*`)
			imageURLs := re.FindAllString(it.Description, -1)

			go sendUpdate(c, imageURLs, it.Title, it.Author, it.Link)
			// logger.Println("图片：", imageURLs[0])
			// pimage, err := c.UploadGroupImage(sendGroupCode, bytes.NewReader(downloadImg(imageURLs[0])))
			// if err != nil {
			// 	logger.WithError(err).Errorf("图片上传失败")
			// }
			// m.Append(pimage)

			// if len(imageURLs) > 1 {
			// 	m.Append(message.NewText("...等" + strconv.Itoa(len(imageURLs)-1) + "张图片\n"))
			// }

			logger.Println("标题：", it.Title)
			logger.Println("作者：", it.Author)
			logger.Println("链接：", it.Link)

		}
		logger.Println(url + " update!")
		//更新rssinfo
		rsses[url] = rssinfo
	} //else {
	// 	fmt.Println(url + " 未更新")
	// }
}

func getRss(url string) (Rss, error) {
	var data Rss
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		//fmt.Println(err)
		return Rss{}, err
	}

	res, err := client.Do(req)
	if err != nil {
		//fmt.Println(err)
		return Rss{}, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		//fmt.Println(err)
		return Rss{}, err
	}
	//fmt.Println(string(body))

	err = xml.Unmarshal(body, &data)
	if err != nil {
		//fmt.Println(err)
		return Rss{}, err
	}
	// fmt.Println(data.Channel.Item[0].Title)
	// fmt.Println(data.Channel.Item[0].Description)
	// fmt.Println(data.Channel.Item[0].Author)
	// fmt.Println(data.getLatestDate())
	return data, nil
}

func getRssInfo(url string) RssInfo {
	var data RssInfo
	rss, err := getRss(url)
	if err != nil {
		logger.Errorln(err)
	}
	data.Rss = rss
	data.LastestDate = ""
	if rss.Channel.Item != nil {
		data.LastestDate = rss.Channel.Item[0].PubDate
	} else {
		return RssInfo{}
	}
	return data
}

func downloadImg(url string) []byte {
	resp, err := http.Get(url)

	if err != nil {
		logger.Errorln("下载图片失败：", err)
		return nil
	}
	defer resp.Body.Close()
	dat, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		logger.Errorln("读取图片失败：", err)
		return nil
	}

	logger.Println("下载图片成功")
	return dat
}
