package mail

import (
	"bytes"
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
	encodedMsg := tools.Base64encode(msg)

	// 使用 bytes.Buffer 减少内存分配
	var buf bytes.Buffer
	buf.Grow(len(sendTemplate) + len(from) + len(to) + len(subject) + len(sendTime) + len(sendVersion) + len(encodedMsg) + len(boundaryRand))

	// 手动替换模板变量，减少字符串分配
	var last int
	for i := 0; i < len(sendTemplate); {
		if i+3 < len(sendTemplate) && sendTemplate[i] == '{' {
			end := strings.Index(sendTemplate[i:], "}")
			if end != -1 {
				end += i
				buf.WriteString(sendTemplate[last:i])

				switch sendTemplate[i+1 : end] {
				case "MAIL_FROM":
					buf.WriteString(from)
				case "RCPT_TO":
					buf.WriteString(to)
				case "SUBJECT":
					buf.WriteString(subject)
				case "TIME":
					buf.WriteString(sendTime)
				case "VERSION":
					buf.WriteString(sendVersion)
				case "CONTENT":
					buf.WriteString(encodedMsg)
				case "BOUNDARY_LINE":
					buf.WriteString(boundaryRand)
				default:
					buf.WriteString(sendTemplate[i : end+1])
				}

				last = end + 1
				i = end + 1
				continue
			}
		}
		i++
	}
	buf.WriteString(sendTemplate[last:])

	return buf.String(), nil
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

	// 优化 HTML 模板替换
	var htmlBuf bytes.Buffer
	htmlBuf.Grow(len(returnHtmlTemplate) + len(msg) + len(sendSubject) + len(err_to_mail) + len(sendTime))

	var last int
	for i := 0; i < len(returnHtmlTemplate); {
		if i+3 < len(returnHtmlTemplate) && returnHtmlTemplate[i] == '{' {
			end := strings.Index(returnHtmlTemplate[i:], "}")
			if end != -1 {
				end += i
				htmlBuf.WriteString(returnHtmlTemplate[last:i])

				switch returnHtmlTemplate[i+1 : end] {
				case "TILTE":
					htmlBuf.WriteString("sc")
				case "ERR_MSG":
					htmlBuf.WriteString(msg)
				case "SEND_SUBJECT":
					htmlBuf.WriteString(sendSubject)
				case "ERR_TO_MAIL":
					htmlBuf.WriteString(err_to_mail)
				case "TIME":
					htmlBuf.WriteString(sendTime)
				default:
					htmlBuf.WriteString(returnHtmlTemplate[i : end+1])
				}

				last = end + 1
				i = end + 1
				continue
			}
		}
		i++
	}
	htmlBuf.WriteString(returnHtmlTemplate[last:])
	contentHtml := htmlBuf.String()
	encodedHtml := tools.Base64encode(contentHtml)

	// 优化主模板替换
	var contentBuf bytes.Buffer
	contentBuf.Grow(len(returnTemplate) + len(mailFrom) + len(rcptTo) + len(sendTime) + len(sendVersion) + len(encodedHtml) + len(boundaryRand))

	last = 0
	for i := 0; i < len(returnTemplate); {
		if i+3 < len(returnTemplate) && returnTemplate[i] == '{' {
			end := strings.Index(returnTemplate[i:], "}")
			if end != -1 {
				end += i
				contentBuf.WriteString(returnTemplate[last:i])

				switch returnTemplate[i+1 : end] {
				case "MAIL_FROM":
					contentBuf.WriteString(mailFrom)
				case "RCPT_TO":
					contentBuf.WriteString(rcptTo)
				case "SUBJECT":
					contentBuf.WriteString("系统退信")
				case "TIME":
					contentBuf.WriteString(sendTime)
				case "VERSION":
					contentBuf.WriteString(sendVersion)
				case "CONTENT":
					contentBuf.WriteString(encodedHtml)
				case "BOUNDARY_LINE":
					contentBuf.WriteString(boundaryRand)
				default:
					contentBuf.WriteString(returnTemplate[i : end+1])
				}

				last = end + 1
				i = end + 1
				continue
			}
		}
		i++
	}
	contentBuf.WriteString(returnTemplate[last:])

	return contentBuf.String(), nil
}
