package translate

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
)

func init() {
	instance = &translation{}
	bot.RegisterModule(instance)
}

var instance *translation

var logger = utils.GetModuleLogger("kumiko.translate")
var ch = make(chan string, 3) //限制同时翻译数为3

type translation struct {
}

func (m *translation) MiraiGoModule() bot.ModuleInfo {
	return bot.ModuleInfo{
		ID:       "kumiko.translate",
		Instance: instance,
	}
}

func (m *translation) Init() {
	// 初始化过程
	// 在此处可以进行 Module 的初始化配置
	// 如配置读取
}

func (m *translation) PostInit() {
	// 第二次初始化
	// 再次过程中可以进行跨Module的动作
	// 如通用数据库等等
}

func (m *translation) Serve(b *bot.Bot) {
	// 注册服务函数部分
	b.OnGroupMessage(func(c *client.QQClient, msg *message.GroupMessage) {
		msgText := msg.ToString()
		if strings.HasPrefix(msgText, "翻译") {
			logger.Println("翻译命令开始")
			tip := message.NewSendingMessage().Append(message.NewAt(msg.Sender.Uin, msg.Sender.DisplayName()))
			tip.Append(message.NewText("开始翻译"))
			c.SendGroupMessage(msg.GroupCode, tip)

			rmt := strings.ReplaceAll(strings.TrimPrefix(msgText, "翻译"), "\r", "")
			go translate(rmt, ch)

			m := message.NewSendingMessage().Append(message.NewText(<-ch))
			m.Append(message.NewReply(msg))
			c.SendGroupMessage(msg.GroupCode, m)
		}
	})
	b.OnPrivateMessage(func(c *client.QQClient, msg *message.PrivateMessage) {
		msgText := msg.ToString()
		if strings.HasPrefix(msgText, "翻译") {
			logger.Println("翻译命令开始")

			rmt := strings.ReplaceAll(strings.TrimPrefix(msgText, "翻译"), "\r", "")
			go translate(rmt, ch)

			m := message.NewSendingMessage().Append(message.NewText(<-ch))
			c.SendPrivateMessage(msg.Sender.Uin, m)
		}
	})
}

func (m *translation) Start(b *bot.Bot) {
	// 此函数会新开携程进行调用
	// ```go
	// 		go exampleModule.Start()
	// ```

	// 可以利用此部分进行后台操作
	// 如http服务器等等
}

func (m *translation) Stop(b *bot.Bot, wg *sync.WaitGroup) {
	// 别忘了解锁
	defer wg.Done()
	// 结束部分
	// 一般调用此函数时，程序接收到 os.Interrupt 信号
	// 即将退出
	// 在此处应该释放相应的资源或者对状态进行保存
}

//TransRes DEEPL翻译返回结果
type TransRes struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  struct {
		Translations []struct {
			Beams []struct {
				PostprocessedSentence string `json:"postprocessed_sentence"`
				NumSymbols            int    `json:"num_symbols"`
			} `json:"beams"`
			Quality string `json:"quality"`
		} `json:"translations"`
		TargetLang            string `json:"target_lang"`
		SourceLang            string `json:"source_lang"`
		SourceLangIsConfident bool   `json:"source_lang_is_confident"`
		DetectedLanguages     struct {
		} `json:"detectedLanguages"`
		Timestamp int    `json:"timestamp"`
		Date      string `json:"date"`
	} `json:"result"`
}

//TransResYoudao 有道翻译返回结果
type TransResYoudao struct {
	Type            string `json:"type"`
	ErrorCode       int    `json:"errorCode"`
	ElapsedTime     int    `json:"elapsedTime"`
	TranslateResult [][]struct {
		Src string `json:"src"`
		Tgt string `json:"tgt"`
	} `json:"translateResult"`
}

func setClient() {

	url := "https://www.deepl.com/PHP/backend/clientState.php?request_type=jsonrpc&il=ZH"
	method := "POST"

	payload := strings.NewReader(`{
		"jsonrpc": "2.0",
		"method": "getClientState",
		"params": {
			"v": "20180814",
			"clientVars": {
				"userCountry": "ZH",
				"showAppOnboarding": true,
				"uid": "d97353e4-89be-4f78-9d08-bc4a936aa178"
			}
		},
		"id": 2560001
	}`)

	timeout := time.Duration(1 * time.Second)
	client := &http.Client{
		Timeout: timeout,
	}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		logger.Errorln(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Cookie", "LMTBID=v2|213ca75e-bf20-4d28-a3a4-5ef87367258f|f87d00d1ecb16e5778c0b750095adac2")

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
}

func translateByDeepl(text string) (string, error) {
	var result TransRes

	url := "https://www2.deepl.com/jsonrpc"
	method := "POST"

	payload := strings.NewReader(`{
		"jsonrpc": "2.0",
		"method": "LMT_handle_jobs",
		"params": {
			"jobs": [
				{
					"kind": "default",
					"raw_en_sentence": "` + text + `",
					"raw_en_context_before": [],
					"raw_en_context_after": [],
					"preferred_num_beams": 1,
					"quality": "fast"
				}
			],
			"lang": {
				"source_lang_user_selected": "auto",
				"target_lang": "ZH"
			},
			"priority": -1,
			"commonJobParams": {
				"formality": null
			},
			"timestamp": 1613966943856
		},
		"id": 2560002
	}`)

	timeout := time.Duration(3 * time.Second)
	client := &http.Client{
		Timeout: timeout,
	}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		logger.Errorln(err)
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Cookie", "LMTBID=v2|213ca75e-bf20-4d28-a3a4-5ef87367258f|f87d00d1ecb16e5778c0b750095adac2")

	res, err := client.Do(req)
	if err != nil {
		logger.Errorln(err)
		return "", err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logger.Errorln(err)
		return "", err
	}
	logger.Println(string(body))

	if strings.Contains(string(body), "error") {
		logger.Errorln("DEEPL翻译接口限制！")
		return "", errors.New("DEEPL翻译接口限制！")
	}
	//解析为json
	err = json.Unmarshal(body, &result)
	if err != nil {
		logger.Errorln(err)
		return "", err
	}

	return result.Result.Translations[0].Beams[0].PostprocessedSentence, nil
}

func translateByYoudao(text string) string {
	var result TransResYoudao
	var fanyi string
	url := "http://fanyi.youdao.com/translate?doctype=json&type=AUTO&i=" + strings.ReplaceAll(text, " ", "%20")
	method := "GET"

	payload := strings.NewReader(``)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		logger.Errorln(err)
		return ""
	}

	res, err := client.Do(req)
	if err != nil {
		logger.Errorln(err)
		return ""
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logger.Errorln(err)
		return ""
	}
	logger.Println(string(body))

	//解析为json
	err = json.Unmarshal(body, &result)
	if err != nil {
		logger.Errorln(err)
		return ""
	}
	for _, part := range result.TranslateResult[0] {
		fanyi += part.Tgt + " "
	}
	logger.Println(fanyi)
	return fanyi
}

func translate(text string, ch chan string) {
	setClient()
	var fanyi string
	for _, line := range strings.Split(text, "\n") {
		l, err := translateByDeepl(line)
		if err != nil {
			logger.Warnln("更换有道翻译")
			text = strings.ReplaceAll(text, "\n", " ")
			ch <- translateByYoudao(text) + "\n——Youdao翻译"
			return
		}
		fanyi += l
		time.Sleep(200000000)
	}
	ch <- fanyi + "\n——DEEPL翻译"
}
