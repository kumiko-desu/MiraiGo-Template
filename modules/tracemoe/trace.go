package tracemoe

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"

	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
)

func init() {
	instance = &trace{}
	bot.RegisterModule(instance)
}

var instance *trace
var logger = utils.GetModuleLogger("kumiko.tracemoe")
var ch = make(chan []byte, 5)
var text = make(chan string, 5)
var searching = make(map[int64]int) //0未请求，1等待图片，2搜索中

type trace struct {
}

//TraceRes res
type TraceRes struct {
	RawDocsCount      int  `json:"RawDocsCount"`
	CacheHit          bool `json:"CacheHit"`
	Trial             int  `json:"trial"`
	Limit             int  `json:"limit"`
	LimitTTL          int  `json:"limit_ttl"`
	Quota             int  `json:"quota"`
	QuotaTTL          int  `json:"quota_ttl"`
	RawDocsSearchTime int  `json:"RawDocsSearchTime"`
	ReRankSearchTime  int  `json:"ReRankSearchTime"`
	Docs              []struct {
		Filename        string        `json:"filename"`
		Episode         int           `json:"episode"`
		From            float64       `json:"from"`
		To              float64       `json:"to"`
		Similarity      float64       `json:"similarity"`
		AnilistID       int           `json:"anilist_id"`
		Anime           string        `json:"anime"`
		At              float64       `json:"at"`
		IsAdult         bool          `json:"is_adult"`
		MalID           int           `json:"mal_id"`
		Season          string        `json:"season"`
		Synonyms        []interface{} `json:"synonyms"`
		SynonymsChinese []interface{} `json:"synonyms_chinese"`
		Title           string        `json:"title"`
		TitleChinese    string        `json:"title_chinese"`
		TitleEnglish    string        `json:"title_english"`
		TitleNative     string        `json:"title_native"`
		TitleRomaji     string        `json:"title_romaji"`
		Tokenthumb      string        `json:"tokenthumb"`
	} `json:"docs"`
}

func (m *trace) MiraiGoModule() bot.ModuleInfo {
	return bot.ModuleInfo{
		ID:       "kumiko.tracemoe",
		Instance: instance,
	}
}

func (m *trace) Init() {
	// 初始化过程
	// 在此处可以进行 Module 的初始化配置
	// 如配置读取

}

func (m *trace) PostInit() {
	// 第二次初始化
	// 再次过程中可以进行跨Module的动作
	// 如通用数据库等等
}

func (m *trace) Serve(b *bot.Bot) {
	// 注册服务函数部分
	b.OnGroupMessage(func(c *client.QQClient, msg *message.GroupMessage) {
		if searching[msg.Sender.Uin] == 1 {
			switch e := msg.Elements[0].(type) {
			case *message.ImageElement:
				searching[msg.Sender.Uin] = 2
				tip := message.NewSendingMessage().Append(message.NewAt(msg.Sender.Uin, msg.Sender.DisplayName()))
				tip.Append(message.NewText("搜索中"))
				c.SendGroupMessage(msg.GroupCode, tip)
				imgURL := e.Url
				logger.Println(imgURL)
				go traceImg(imgURL, ch, text)

				m := message.NewSendingMessage().Append(message.NewAt(msg.Sender.Uin, msg.Sender.DisplayName()))
				pimage, err := c.UploadGroupImage(msg.GroupCode, bytes.NewReader(<-ch))
				if err != nil {
					logger.WithError(err).Errorf("图片上传失败")
				}
				m.Append(pimage)
				m.Append(message.NewText(<-text))
				c.SendGroupMessage(msg.GroupCode, m)
				searching[msg.Sender.Uin] = 0
			}
		}

		if msg.ToString() == "以图搜番" {
			m := message.NewSendingMessage().Append(message.NewAt(msg.Sender.Uin, msg.Sender.DisplayName()))
			if searching[msg.Sender.Uin] == 2 {
				m.Append(message.NewText("请等待上次搜索"))
			} else {
				m.Append(message.NewText("请发送一张图片，不支持gif"))
			}
			c.SendGroupMessage(msg.GroupCode, m)
			searching[msg.Sender.Uin] = 1
		}

	})

	b.OnPrivateMessage(func(c *client.QQClient, msg *message.PrivateMessage) {

	})
}

func (m *trace) Start(b *bot.Bot) {
	// 此函数会新开携程进行调用
	// ```go
	// 		go exampleModule.Start()
	// ```

	// 可以利用此部分进行后台操作
	// 如http服务器等等
}

func (m *trace) Stop(b *bot.Bot, wg *sync.WaitGroup) {
	// 别忘了解锁
	defer wg.Done()
	// 结束部分
	// 一般调用此函数时，程序接收到 os.Interrupt 信号
	// 即将退出
	// 在此处应该释放相应的资源或者对状态进行保存
}

func (animeList TraceRes) parseMedia() []byte {
	if animeList.Docs[0].IsAdult {
		return nil
	}
	data := animeList.Docs[0]
	anilistID := strconv.Itoa(data.AnilistID)
	filename := data.Filename
	at := strconv.FormatFloat(data.At, 'f', -1, 64)
	tokenthumb := data.Tokenthumb

	url := "https://media.trace.moe/image/" + anilistID + "/" + filename + "?t=" + at + "&token=" + tokenthumb
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		logger.Errorln(err)
		return nil
	}

	res, err := client.Do(req)
	if err != nil {
		logger.Errorln(err)
		return nil
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logger.Errorln(err)
		return nil
	}
	//fmt.Println(string(body))
	return body
}
func (animeList TraceRes) getTime() string {
	anime := animeList.Docs[0]
	time := int(anime.At)
	episode := anime.Episode
	return anime.Anime + "\n第" + strconv.Itoa(episode) + "集 " + strconv.Itoa(time/60) + "分" + strconv.Itoa(time%60) + "秒"
}

func traceImg(imgURL string, ch chan []byte, text chan string) {
	var animeList TraceRes

	url := "https://trace.moe/api/search?url=" + imgURL
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		logger.Errorln(err)
		return
	}
	res, err := client.Do(req)
	if err != nil {
		logger.Errorln(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logger.Errorln(err)
		return
	}
	//fmt.Println(string(body))

	err = json.Unmarshal(body, &animeList)
	if err != nil {
		logger.Errorln(err)
	}

	ch <- animeList.parseMedia()
	text <- animeList.getTime()

	return
}
