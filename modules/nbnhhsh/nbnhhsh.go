package nbnhhsh

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
)

func init() {
	instance = &guess{}
	bot.RegisterModule(instance)
}

var instance *guess
var logger = utils.GetModuleLogger("kumiko.guess")
var ch = make(chan string, 5)

type guess struct {
}

type guessRes []struct {
	Name  string   `json:"name"`
	Trans []string `json:"trans"`
}

func (m *guess) MiraiGoModule() bot.ModuleInfo {
	return bot.ModuleInfo{
		ID:       "kumiko.guess",
		Instance: instance,
	}
}

func (m *guess) Init() {
	// 初始化过程
	// 在此处可以进行 Module 的初始化配置
	// 如配置读取

}

func (m *guess) PostInit() {
	// 第二次初始化
	// 再次过程中可以进行跨Module的动作
	// 如通用数据库等等
}

func (m *guess) Serve(b *bot.Bot) {
	// 注册服务函数部分
	b.OnGroupMessage(func(c *client.QQClient, msg *message.GroupMessage) {
		msgText := msg.ToString()
		if strings.HasPrefix(msgText, "缩写翻译") {
			text := strings.TrimPrefix(msgText, "缩写翻译")
			go guessAbbr(text, ch)

			m := message.NewSendingMessage().Append(message.NewAt(msg.Sender.Uin, msg.Sender.DisplayName()))
			m.Append(message.NewText("\n"))
			m.Append(message.NewText(<-ch))
			c.SendGroupMessage(msg.GroupCode, m)
		}
	})

	b.OnPrivateMessage(func(c *client.QQClient, msg *message.PrivateMessage) {

	})
}

func (m *guess) Start(b *bot.Bot) {
	// 此函数会新开携程进行调用
	// ```go
	// 		go exampleModule.Start()
	// ```

	// 可以利用此部分进行后台操作
	// 如http服务器等等
}

func (m *guess) Stop(b *bot.Bot, wg *sync.WaitGroup) {
	// 别忘了解锁
	defer wg.Done()
	// 结束部分
	// 一般调用此函数时，程序接收到 os.Interrupt 信号
	// 即将退出
	// 在此处应该释放相应的资源或者对状态进行保存
}

func guessAbbr(text string, ch chan string) {
	var data guessRes
	var str string
	url := "https://lab.magiconch.com/api/nbnhhsh/guess"
	method := "POST"

	payload := strings.NewReader(`{"text": "` + text + `"}`)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		logger.Errorln(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")

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
	logger.Println(string(body))

	err = json.Unmarshal(body, &data)
	if err != nil {
		logger.Errorln(err)
		return
	}

	for _, e := range data {
		r := e.Name + ": "
		for _, tran := range e.Trans {
			r += tran + ", "
		}
		str += strings.TrimSuffix(r, ", ") + "\n"
	}
	ch <- strings.TrimSuffix(str, "\n")
}
