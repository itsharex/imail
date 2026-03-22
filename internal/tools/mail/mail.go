package mail

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"time"

	"github.com/midoks/imail/internal/conf"
	"github.com/midoks/imail/internal/tools"
)

// 预编译正则表达式
var (
	subjectRegex = regexp.MustCompile(`Subject: (.*)`)
	fromRegex    = regexp.MustCompile(`From: (.*)`)
)

func GetMailSubject(content string) string {
	var err error
	match := subjectRegex.FindAllStringSubmatch(content, -1)
	if len(match) == 0 {
		return ""
	}

	val := match[0][0]
	tmp := strings.SplitN(val, ":", 2)
	if len(tmp) < 2 {
		return ""
	}
	val = strings.TrimSpace(tmp[1])

	if strings.Contains(val, "=?utf-8?B?") || strings.Contains(val, "=?UTF-8?B?") {
		val = strings.Replace(val, "=?utf-8?B?", "", -1)
		val = strings.Replace(val, "=?UTF-8?B?", "", -1)
		val = strings.Replace(val, "?=", "", -1)
		val = strings.TrimSpace(val)
		val, err = tools.Base64decode(val)
		if err == nil {
			return val
		}
	}

	if strings.Contains(val, "=?gbk?B?") || strings.Contains(val, "=?GBK?B?") {
		val = strings.Replace(val, "=?gbk?B?", "", -1)
		val = strings.Replace(val, "=?GBK?B?", "", -1)
		val = strings.Replace(val, "?=", "", -1)
		val = strings.TrimSpace(val)
		val, err = tools.Base64decode(val)
		if err == nil {
			val = tools.ConvertToString(val, "gbk", "utf-8")
			return val
		}
	}
	return val
}

func GetMailFromInContent(content string) string {
	var err error
	match := fromRegex.FindAllStringSubmatch(content, -1)
	if len(match) == 0 {
		return ""
	}

	val := match[0][0]
	tmp := strings.SplitN(val, ":", 2)
	if len(tmp) < 2 {
		return ""
	}
	val = strings.TrimSpace(tmp[1])

	tmp = strings.SplitN(val, "<", 2)
	if len(tmp) < 2 {
		return ""
	}
	val = strings.TrimSpace(tmp[0])
	val = strings.Trim(val, "\"")

	if strings.EqualFold(val, "") {
		val = tmp[1]
		val = strings.Trim(val, ">")
		tmp = strings.SplitN(val, "@", 2)
		if len(tmp) > 0 {
			return tmp[0]
		}
		return ""
	}

	if strings.Contains(val, "=?utf-8?B?") || strings.Contains(val, "=?UTF-8?B?") {
		val = strings.Replace(val, "=?utf-8?B?", "", -1)
		val = strings.Replace(val, "=?UTF-8?B?", "", -1)
		val = strings.Replace(val, "?=", "", -1)
		val = strings.TrimSpace(val)
		val, err = tools.Base64decode(val)
		if err == nil {
			return val
		}
	}
	return val
}

// 模板缓存
var (
	sendTemplate        string
	returnTemplate      string
	returnHtmlTemplate  string
	templateCacheLoaded bool
)

// loadTemplates 加载模板文件到缓存
func loadTemplates() error {
	if templateCacheLoaded {
		return nil
	}

	workDir := conf.WorkDir()

	// 加载发送模板
	data, err := ioutil.ReadFile(workDir + "/conf/tpl/send.tpl")
	if err != nil {
		return err
	}
	sendTemplate = string(data)

	// 加载退信模板
	data, err = ioutil.ReadFile(workDir + "/conf/tpl/return_to_sender.tpl")
	if err != nil {
		return err
	}
	returnTemplate = string(data)

	// 加载退信HTML模板
	data, err = ioutil.ReadFile(workDir + "/conf/tpl/return_to_sender_html.tpl")
	if err != nil {
		return err
	}
	returnHtmlTemplate = string(data)

	templateCacheLoaded = true
	return nil
}

func GetMailSend(from string, to string, subject string, msg string) (string, error) {
	if err := loadTemplates(); err != nil {
		return "", err
	}

	sendTime := time.Now().Format("Mon, 02 Jan 2006 15:04:05 -0700 (MST)")
	sendVersion := fmt.Sprintf("imail/%s", conf.App.Version)
	boundaryRand := tools.RandString(20)

	content := strings.Replace(sendTemplate, "{MAIL_FROM}", from, -1)
	content = strings.Replace(content, "{RCPT_TO}", to, -1)
	content = strings.Replace(content, "{SUBJECT}", subject, -1)
	content = strings.Replace(content, "{TIME}", sendTime, -1)
	content = strings.Replace(content, "{VERSION}", sendVersion, -1)
	content = strings.Replace(content, "{CONTENT}", tools.Base64encode(msg), -1)
	content = strings.Replace(content, "{BOUNDARY_LINE}", boundaryRand, -1)

	return content, nil
}

// 邮件退信模板
func GetMailReturnToSender(mailFrom, rcptTo string, err_to_mail string, err_content string, msg string) (string, error) {
	if err := loadTemplates(); err != nil {
		return "", err
	}

	sendSubject := GetMailSubject(err_content)

	sendTime := time.Now().Format("Mon, 02 Jan 2006 15:04:05 -0700 (MST)")
	sendVersion := fmt.Sprintf("imail/%s", conf.App.Version)
	boundaryRand := tools.RandString(20)

	contentHtml := strings.Replace(returnHtmlTemplate, "{TILTE}", "sc", -1)
	contentHtml = strings.Replace(contentHtml, "{ERR_MSG}", msg, -1)
	contentHtml = strings.Replace(contentHtml, "{SEND_SUBJECT}", sendSubject, -1)
	contentHtml = strings.Replace(contentHtml, "{ERR_TO_MAIL}", err_to_mail, -1)
	contentHtml = strings.Replace(contentHtml, "{TIME}", sendTime, -1)

	content := strings.Replace(returnTemplate, "{MAIL_FROM}", mailFrom, -1)
	content = strings.Replace(content, "{RCPT_TO}", rcptTo, -1)
	content = strings.Replace(content, "{SUBJECT}", "系统退信", -1)
	content = strings.Replace(content, "{TIME}", sendTime, -1)
	content = strings.Replace(content, "{VERSION}", sendVersion, -1)
	content = strings.Replace(content, "{CONTENT}", tools.Base64encode(contentHtml), -1)
	content = strings.Replace(content, "{BOUNDARY_LINE}", boundaryRand, -1)
	return content, nil
}
