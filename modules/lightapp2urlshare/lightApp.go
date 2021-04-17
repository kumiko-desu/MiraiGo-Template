package lightapp2urlshare

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
)

func init() {
	instance = &lightapp2urlshare{}
	bot.RegisterModule(instance)
}

type lightapp2urlshare struct {
}
type lightApp struct {
	App    string `json:"app"`
	Desc   string `json:"desc"`
	View   string `json:"view"`
	Ver    string `json:"ver"`
	Prompt string `json:"prompt"`
	Meta   struct {
		Detail1 struct {
			Appid   string `json:"appid"`
			Title   string `json:"title"`
			Desc    string `json:"desc"`
			Icon    string `json:"icon"`
			Preview string `json:"preview"`
			URL     string `json:"url"`
			Scene   int    `json:"scene"`
			Host    struct {
				Uin  int    `json:"uin"`
				Nick string `json:"nick"`
			} `json:"host"`
			ShareTemplateID   string `json:"shareTemplateId"`
			ShareTemplateData struct {
			} `json:"shareTemplateData"`
			Qqdocurl       string `json:"qqdocurl"`
			ShowLittleTail string `json:"showLittleTail"`
			GamePoints     string `json:"gamePoints"`
			GamePointsURL  string `json:"gamePointsUrl"`
		} `json:"detail_1"`
	} `json:"meta"`
	Config struct {
		Type     string `json:"type"`
		Width    int    `json:"width"`
		Height   int    `json:"height"`
		Forward  int    `json:"forward"`
		AutoSize int    `json:"autoSize"`
		Ctime    int    `json:"ctime"`
		Token    string `json:"token"`
	} `json:"config"`
}

var data lightApp

func (m *lightapp2urlshare) MiraiGoModule() bot.ModuleInfo {
	return bot.ModuleInfo{
		ID:       "internal.lightApp",
		Instance: instance,
	}
}

func (m *lightapp2urlshare) Init() {
	// 初始化过程
	// 在此处可以进行 Module 的初始化配置
	// 如配置读取
}

func (m *lightapp2urlshare) PostInit() {
	// 第二次初始化
	// 再次过程中可以进行跨Module的动作
	// 如通用数据库等等
}

func (m *lightapp2urlshare) Serve(b *bot.Bot) {
	// 注册服务函数部分
	b.OnGroupMessage(func(c *client.QQClient, msg *message.GroupMessage) {
		for _, elem := range msg.Elements {
			//println(elem.Type())
			switch e := elem.(type) {
			case *message.LightAppElement:
				logger.Println(e.Content)
				err := json.Unmarshal([]byte(e.Content), &data)
				if err != nil {
					logger.Error(err)
				} else {
					// "com.tencent.qq.checkin", "com.tencent.groupphoto"
					if data.App == "com.tencent.qq.checkin" || data.App == "com.tencent.groupphoto" {
						break
					}
					m := message.NewSendingMessage()
					url := strings.Split(data.Meta.Detail1.Qqdocurl, "?")[0] //链接不能含有参数
					title := data.Meta.Detail1.Title + "[小程序转换]"
					content := data.Meta.Detail1.Desc
					image := "http://" + data.Meta.Detail1.Preview
					image = strings.Split(image, "?")[0] //链接不能含有参数
					icon := data.Meta.Detail1.Icon
					template := fmt.Sprintf(`<?xml version='1.0' encoding='UTF-8' standalone='yes'?><msg templateID="123" url="%s" serviceID="1" action="web" actionData="" a_actionData="" i_actionData="" brief="[小程序转换]%s" flag="0"><item layout="2"><picture cover="%s"/><title>%s</title><summary>%s</summary></item><source url="%s" icon="%s" name="关爱TIM用户" appid="0" action="web" actionData="" a_actionData="tencent0://" i_actionData=""/></msg>`,
						url, title, image, title, content, url, icon,
					)
					m.Append(message.NewRichXml(template, 0))
					c.SendGroupMessage(msg.GroupCode, m)
				}
			}
		}
	})
}

func (m *lightapp2urlshare) Start(b *bot.Bot) {
	// 此函数会新开携程进行调用
	// ```go
	// 		go exampleModule.Start()
	// ```

	// 可以利用此部分进行后台操作
	// 如http服务器等等
}

func (m *lightapp2urlshare) Stop(b *bot.Bot, wg *sync.WaitGroup) {
	// 别忘了解锁
	defer wg.Done()
	// 结束部分
	// 一般调用此函数时，程序接收到 os.Interrupt 信号
	// 即将退出
	// 在此处应该释放相应的资源或者对状态进行保存
}

var instance *lightapp2urlshare

var logger = utils.GetModuleLogger("internal.lightApp")
