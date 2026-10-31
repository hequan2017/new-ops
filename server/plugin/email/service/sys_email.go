package service

import (
	"fmt"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/email/utils"
)

type EmailService struct{}

//@author: [maplepie](https://github.com/maplepie)
//@function: EmailTest
//@description: 发送邮件测试
//@return: err error

func (e *EmailService) EmailTest() (err error) {
	subject := "test"
	body := "test"
	err = utils.EmailTest(subject, body)
	return err
}

//@author: [maplepie](https://github.com/maplepie)
//@function: EmailTest
//@description: 发送邮件测试
//@return: err error
//@params to string 	 收件人
//@params subject string   标题（主题）
//@params body  string 	 邮件内容

// SendEmail 发送邮件（失败自动重试一次并留痕：SMTP 瞬时抖动常见，重试前短暂退避）
func (e *EmailService) SendEmail(to, subject, body string) (err error) {
	err = utils.Email(to, subject, body)
	if err == nil {
		return nil
	}
	global.GVA_LOG.Warn(fmt.Sprintf("邮件发送失败，将重试一次: to=%s subject=%s err=%v", to, subject, err))
	time.Sleep(500 * time.Millisecond)
	if err = utils.Email(to, subject, body); err != nil {
		global.GVA_LOG.Error(fmt.Sprintf("邮件重试仍失败: to=%s subject=%s err=%v", to, subject, err))
	}
	return err
}
