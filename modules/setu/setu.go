package setu

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
	instance = &st{}
	bot.RegisterModule(instance)
}

var instance *st
var logger = utils.GetModuleLogger("kumiko.st")
var apikey string
var ch = make(chan []byte, 5)
var getting = make(map[int64]bool) //false未请求，true获取中

type st struct {
}

type result struct {
	Code        int    `json:"code"`
	Msg         string `json:"msg"`
	Quota       int    `json:"quota"`
	QuotaMinTTL int    `json:"quota_min_ttl"`
	Count       int    `json:"count"`
	Data        []struct {
		Pid    int      `json:"pid"`
		P      int      `json:"p"`
		UID    int      `json:"uid"`
		Title  string   `json:"title"`
		Author string   `json:"author"`
		URL    string   `json:"url"`
		R18    bool     `json:"r18"`
		Width  int      `json:"width"`
		Height int      `json:"height"`
		Tags   []string `json:"tags"`
	} `json:"data"`
}

func (m *st) MiraiGoModule() bot.ModuleInfo {
	return bot.ModuleInfo{
		ID:       "kumiko.st",
		Instance: instance,
	}
}

func (m *st) Init() {
	// 初始化过程
	// 在此处可以进行 Module 的初始化配置
	// 如配置读取
	apikey = config.GlobalConfig.GetString("loliconApikey")
}

func (m *st) PostInit() {
	// 第二次初始化
	// 再次过程中可以进行跨Module的动作
	// 如通用数据库等等
}

func (m *st) Serve(b *bot.Bot) {
	// 注册服务函数部分
	b.OnGroupMessage(func(c *client.QQClient, msg *message.GroupMessage) {
		if msg.ToString() == "色图" {
			if getting[msg.Sender.Uin] { //正在获取中
				m := message.NewSendingMessage().Append(message.NewAt(msg.Sender.Uin, msg.Sender.DisplayName()))
				m.Append(message.NewText("请等待"))
				c.SendGroupMessage(msg.GroupCode, m)
			} else {
				getting[msg.Sender.Uin] = true
				logger.Println("色图命令开始")
				tip := message.NewSendingMessage().Append(message.NewText("开始色图"))
				tip.Append(message.NewAt(msg.Sender.Uin))
				c.SendGroupMessage(msg.GroupCode, tip)

				go getSetu(ch)
				pimage, err := c.UploadGroupImage(msg.GroupCode, bytes.NewReader(<-ch))
				if err != nil {
					logger.WithError(err).Errorf("图片上传失败")
				}
				m := message.NewSendingMessage().Append(pimage)
				m.Append(message.NewAt(msg.Sender.Uin))
				c.SendGroupMessage(msg.GroupCode, m)
				getting[msg.Sender.Uin] = false
			}
		}
	})

	b.OnPrivateMessage(func(c *client.QQClient, msg *message.PrivateMessage) {
		if msg.ToString() == "色图" {
			logger.Println("色图命令开始")
			go getSetu(ch)
			pimage, err := c.UploadPrivateImage(msg.Sender.Uin, bytes.NewReader(<-ch))
			if err != nil {
				logger.WithError(err).Errorf("图片上传失败")
			}
			m := message.NewSendingMessage().Append(pimage)
			c.SendPrivateMessage(msg.Sender.Uin, m)
		}
	})
}

func (m *st) Start(b *bot.Bot) {
	// 此函数会新开携程进行调用
	// ```go
	// 		go exampleModule.Start()
	// ```

	// 可以利用此部分进行后台操作
	// 如http服务器等等
}

func (m *st) Stop(b *bot.Bot, wg *sync.WaitGroup) {
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

func getSetu(ch chan []byte) {
	var res result
	//发送get请求
	resp, err := http.Get("https://api.lolicon.app/setu?size1200=1&apikey="+apikey)
	if err != nil {
		logger.Errorln("get failed, err", err)
		return
	}
	//关闭Body
	defer resp.Body.Close()
	//读取body内容
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Errorln("read from resp.Body failed, err", err)
		return
	}
	//输出字符串内容
	logger.Println("下载链接成功")

	json.Unmarshal(body, &res)

	url := res.Data[0].URL

	logger.Println("获取链接:", url)

	ch <- downloadImg(url)
}
