package joinandquit

import (
	"strconv"
	"sync"

	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
)

func init() {
	instance = &interacting{}
	bot.RegisterModule(instance)
}

type interacting struct {
}

func (m *interacting) MiraiGoModule() bot.ModuleInfo {
	return bot.ModuleInfo{
		ID:       "internal.interacting",
		Instance: instance,
	}
}

func (m *interacting) Init() {
	// 初始化过程
	// 在此处可以进行 Module 的初始化配置
	// 如配置读取
}

func (m *interacting) PostInit() {
	// 第二次初始化
	// 再次过程中可以进行跨Module的动作
	// 如通用数据库等等
}

func (m *interacting) Serve(b *bot.Bot) {
	// 注册服务函数部分
	b.OnGroupMemberJoined(func(c *client.QQClient, event *client.MemberJoinGroupEvent) {
		m := message.NewSendingMessage().Append(message.NewText("欢迎 " + event.Member.Nickname + "(" + strconv.FormatInt(event.Member.Uin, 10) + ") 加入"))
		c.SendGroupMessage(event.Group.Uin, m)
	})
	b.OnGroupMemberLeaved(func(c *client.QQClient, event *client.MemberLeaveGroupEvent) {
		m := message.NewSendingMessage().Append(message.NewText(event.Member.Nickname + "(" + strconv.FormatInt(event.Member.Uin, 10) + ") 离开了，丢人"))
		c.SendGroupMessage(event.Group.Uin, m)
	})
	b.OnGroupMuted(func(c *client.QQClient, event *client.GroupMuteEvent) { //strconv.FormatInt
		groupInfo := c.FindGroup(event.GroupCode)
		operatorName := groupInfo.FindMember(event.OperatorUin).DisplayName()
		targetName := groupInfo.FindMember(event.TargetUin).DisplayName()

		time := ""
		if event.Time != 0 {
			second := event.Time / 60
			if second >= 1440 {
				day := second / 1440
				second = second % 1440
				time += strconv.FormatInt(int64(day), 10) + "天"
			}
			if second >= 60 {
				hour := second / 60
				second = second % 60
				time += strconv.FormatInt(int64(hour), 10) + "小时"
			}
			if second != 0 {
				time += strconv.FormatInt(int64(second), 10) + "分钟"
			}
			m := message.NewSendingMessage().Append(message.NewText(targetName + "喝下了" + operatorName + "的红茶，昏睡" + time))
			c.SendGroupMessage(event.GroupCode, m)
		} else {
			m := message.NewSendingMessage().Append(message.NewText(targetName + "听见了" + operatorName + "的野兽咆哮，" + "缓缓苏醒过来"))
			c.SendGroupMessage(event.GroupCode, m)
		}
	})
}

func (m *interacting) Start(b *bot.Bot) {
	// 此函数会新开携程进行调用
	// ```go
	// 		go exampleModule.Start()
	// ```

	// 可以利用此部分进行后台操作
	// 如http服务器等等
}

func (m *interacting) Stop(b *bot.Bot, wg *sync.WaitGroup) {
	// 别忘了解锁
	defer wg.Done()
	// 结束部分
	// 一般调用此函数时，程序接收到 os.Interrupt 信号
	// 即将退出
	// 在此处应该释放相应的资源或者对状态进行保存
}

var instance *interacting

var logger = utils.GetModuleLogger("internal.interacting")
