package saucenao

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/config"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
)

func init() {
	instance = &sauce{}
	bot.RegisterModule(instance)
}

var instance *sauce
var logger = utils.GetModuleLogger("kumiko.saucenao")
var apikey string
var ch = make(chan []byte, 5)
var text = make(chan string, 5)
var searching = make(map[int64]int) //0未请求，1等待图片，2搜索中

type sauce struct {
}

type saucenaoRes struct {
	Header  interface{} `json:"header"`
	Results []struct {
		Header struct {
			Similarity string `json:"similarity"`
			Thumbnail  string `json:"thumbnail"`
			IndexID    int    `json:"index_id"`
			IndexName  string `json:"index_name"`
			Dupes      int    `json:"dupes"`
		} `json:"header"`
		Data struct {
			ExtUrls    []string `json:"ext_urls"`
			DanbooruID int      `json:"danbooru_id"`
			GelbooruID int      `json:"gelbooru_id"`
			Creator    string   `json:"creator"`
			Material   string   `json:"material"`
			Characters string   `json:"characters"`
			Source     string   `json:"source"`
		} `json:"data"`
	} `json:"results"`
}

func (m *sauce) MiraiGoModule() bot.ModuleInfo {
	return bot.ModuleInfo{
		ID:       "kumiko.saucenao",
		Instance: instance,
	}
}

func (m *sauce) Init() {
	// 初始化过程
	// 在此处可以进行 Module 的初始化配置
	// 如配置读取
	apikey = config.GlobalConfig.GetString("saucenaoApikey")
}

func (m *sauce) PostInit() {
	// 第二次初始化
	// 再次过程中可以进行跨Module的动作
	// 如通用数据库等等
}

func (m *sauce) Serve(b *bot.Bot) {
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
				go sauceImg(imgURL, ch, text)

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

		if msg.ToString() == "以图搜图" {
			m := message.NewSendingMessage().Append(message.NewAt(msg.Sender.Uin, msg.Sender.DisplayName()))
			if searching[msg.Sender.Uin] == 2 {
				m.Append(message.NewText("请等待上次搜索"))
			} else {
				m.Append(message.NewText("请发送一张图片"))
			}
			c.SendGroupMessage(msg.GroupCode, m)
			searching[msg.Sender.Uin] = 1
		}

	})

	b.OnPrivateMessage(func(c *client.QQClient, msg *message.PrivateMessage) {

	})
}

func (m *sauce) Start(b *bot.Bot) {
	// 此函数会新开携程进行调用
	// ```go
	// 		go exampleModule.Start()
	// ```

	// 可以利用此部分进行后台操作
	// 如http服务器等等
}

func (m *sauce) Stop(b *bot.Bot, wg *sync.WaitGroup) {
	// 别忘了解锁
	defer wg.Done()
	// 结束部分
	// 一般调用此函数时，程序接收到 os.Interrupt 信号
	// 即将退出
	// 在此处应该释放相应的资源或者对状态进行保存
}

func downloadImg(url string) []byte {
	resp, err := http.Get(url)

	if err != nil {
		logger.Errorln("下载图片失败", err)
		return nil
	}
	defer resp.Body.Close()
	dat, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		logger.Errorln("读取图片失败", err)
		return nil
	}

	logger.Println("下载图片成功")
	return dat
}

func (data saucenaoRes) getText() string {
	similarity := data.Results[0].Header.Similarity
	source := data.Results[0].Data.Source
	if source == "" {
		source = data.Results[0].Data.ExtUrls[0]
	}
	return "相似度：" + similarity + "\n" + "源地址：" + source
}

func sauceImg(imgURL string, ch chan []byte, text chan string) {
	var data saucenaoRes

	url := "https://saucenao.com/search.php?db=999&output_type=2&testmode=1&numres=1&api_key=" + apikey + "&url=" + imgURL
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

	err = json.Unmarshal(body, &data)
	if err != nil {
		logger.Errorln(err)
	}

	ch <- downloadImg(data.Results[0].Header.Thumbnail)
	text <- data.getText()

	return
}
