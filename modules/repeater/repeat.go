package repeater

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/config"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
)

func init() {
	instance = &repeat{}
	bot.RegisterModule(instance)
}

var instance *repeat
var logger = utils.GetModuleLogger("kumiko.repeat")
var tem map[string][]string

type repeat struct {
}

func (m *repeat) MiraiGoModule() bot.ModuleInfo {
	return bot.ModuleInfo{
		ID:       "kumiko.repeat",
		Instance: instance,
	}
}

func (m *repeat) Init() {
	// 初始化过程
	// 在此处可以进行 Module 的初始化配置
	// 如配置读取
	path := config.GlobalConfig.GetString("kumiko.repeat.path")

	if path == "" {
		path = "./repeats.json"
	}

	bytes := utils.ReadFile(path)
	err := json.Unmarshal(bytes, &tem)

	if err != nil {
		logger.WithError(err).Errorf("unable to read config file in %s", path)
	}
}

func (m *repeat) PostInit() {
	// 第二次初始化
	// 再次过程中可以进行跨Module的动作
	// 如通用数据库等等
}

func (m *repeat) Serve(b *bot.Bot) {
	// 注册服务函数部分
	b.OnGroupMessage(func(c *client.QQClient, msg *message.GroupMessage) {
		if isRepeat(msg.ToString()) {
			m := message.NewSendingMessage().Append(message.NewText(msg.ToString()))
			c.SendGroupMessage(msg.GroupCode, m)
		}
	})

	b.OnPrivateMessage(func(c *client.QQClient, msg *message.PrivateMessage) {
		if isRepeat(msg.ToString()) {
			m := message.NewSendingMessage().Append(message.NewText(msg.ToString()))
			c.SendPrivateMessage(msg.Sender.Uin, m)
		}
	})
}

func (m *repeat) Start(b *bot.Bot) {
	// 此函数会新开携程进行调用
	// ```go
	// 		go exampleModule.Start()
	// ```

	// 可以利用此部分进行后台操作
	// 如http服务器等等
}

func (m *repeat) Stop(b *bot.Bot, wg *sync.WaitGroup) {
	// 别忘了解锁
	defer wg.Done()
	// 结束部分
	// 一般调用此函数时，程序接收到 os.Interrupt 信号
	// 即将退出
	// 在此处应该释放相应的资源或者对状态进行保存
}

func isRepeat(in string) bool {
	println(in)
	for _, word := range tem["words"] {
		if strings.Contains(in, word) {
			return true
		}
	}
	return false
}
