package service

import (
	"bytes"
	"crypto/tls"
	"email/dao"
	"email/global"
	"email/service/aws"
	"email/utils"
	"errors"
	"fmt"
	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
	"github.com/jhillyerd/enmime"
	"io"
	"log"
	"net/mail"
	"os"
	"strings"
	"time"
)

// Backend 结构体
type Backend struct{}

type Session struct {
	authenticated bool
	recipients    []string
	from          string
}

type EmailError struct {
	Code    string
	Message string
	Err     error
}

func (e *EmailError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// 前置檢查
func (s *Session) preCheck() error {
	if !s.authenticated {
		return smtp.ErrAuthRequired
	}
	return nil
}

// 解析郵件
func (s *Session) parseEmail(r io.Reader) ([]byte, *enmime.Envelope, error) {
	// 讀取數據
	rawMessage, err := io.ReadAll(r)
	if err != nil {
		return nil, nil, fmt.Errorf("讀取郵件數據失敗: %w", err)
	}

	if len(rawMessage) == 0 {
		return nil, nil, errors.New("郵件內容為空")
	}

	// 解析郵件信封
	env, err := enmime.ReadEnvelope(bytes.NewReader(rawMessage))
	if err != nil {
		return nil, nil, fmt.Errorf("解析郵件信封失敗: %w", err)
	}

	return rawMessage, env, nil
}

// 驗證附件
func (s *Session) validateAttachments(env *enmime.Envelope) error {
	var totalSize int64
	for _, att := range env.Attachments {
		totalSize += int64(len(att.Content))
		if totalSize > 30*1024*1024 {
			return &EmailError{
				Code:    "AttachmentTooLarge",
				Message: fmt.Sprintf("附件總大小超過限制 %dMB", 30),
			}
		}
	}
	return nil
}

// 驗證郵件
func (s *Session) validateEmail(env *enmime.Envelope) error {
	// 檢查附件大小
	if err := s.validateAttachments(env); err != nil {
		return err
	}

	// 可以添加其他驗證
	return nil
}

// 解析邮件地址列表
func parseAddressList(addrList string) []string {
	if addrList == "" {
		return nil
	}

	addresses, err := mail.ParseAddressList(addrList)
	if err != nil {
		global.Log.Errorf("解析邮件地址失败: %v", err)
		return nil
	}

	result := make([]string, len(addresses))
	for i, addr := range addresses {
		result[i] = addr.Address // 只获取邮件地址部分
	}
	return result
}

// 分类收件人
func (s *Session) classifyRecipients(env *enmime.Envelope) (to, cc, bcc []string) {
	// 1. 解析邮件头中的收件人
	toList := parseAddressList(env.GetHeader("To"))
	ccList := parseAddressList(env.GetHeader("Cc"))

	// 2. 创建明确收件人的映射
	explicitRecipients := make(map[string]bool)
	for _, addr := range toList {
		explicitRecipients[addr] = true
	}
	for _, addr := range ccList {
		explicitRecipients[addr] = true
	}

	// 3. 找出 BCC 收件人（在 RCPT TO 中但不在邮件头中的收件人）
	bccList := make([]string, 0)
	for _, rcpt := range s.recipients {
		if !explicitRecipients[rcpt] {
			bccList = append(bccList, rcpt)
		}
	}

	return toList, ccList, bccList
}

// NewSession 实现 smtp.Backend 接口
func (bkd *Backend) NewSession(_ *smtp.Conn) (smtp.Session, error) {
	return &Session{
		authenticated: false,
	}, nil
}

// AnonymousLogin 处理匿名登录（这里我们禁用它）
func (bkd *Backend) AnonymousLogin(_ *smtp.Conn) (smtp.Session, error) {
	return nil, smtp.ErrAuthRequired
}

// Session 结构体，实现 smtp.Session 和 smtp.AuthSession 接口

func (s *Session) Mail(from string, opts *smtp.MailOptions) error {
	if !s.authenticated {
		return smtp.ErrAuthRequired
	}
	log.Printf("新邮件来自: %s", from)
	return nil
}

func (s *Session) Rcpt(to string, opts *smtp.RcptOptions) error {
	if !s.authenticated {
		return smtp.ErrAuthRequired
	}
	s.recipients = append(s.recipients, to)
	global.Log.Infof("收到收件人: %s", to) // 可以看到所有收件人，包括密送
	return nil
}

func (s *Session) Data(r io.Reader) error {
	if err := s.preCheck(); err != nil {
		return err
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("讀取郵件數據失敗: %w", err)
	}

	env, err := enmime.ReadEnvelope(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("解析邮件失败: %w", err)
	}
	to, cc, bcc := s.classifyRecipients(env)

	// 5. 验证邮件
	if err = s.validateEmail(env); err != nil {
		return err
	}
	global.Log.Infof("邮件验证通过，发件人 =》 ", s.from)
	// 6. 发送邮件
	err = aws.SendEmailByAwsSesWithRawMessage(data, s.from, to, cc, bcc)
	if err != nil {
		return fmt.Errorf("send email failed: %w", err)
	}

	// 7. 保存到数据库
	SaveThirtyPartySendEmailProcess(env, s.from, to, cc, bcc)

	return nil
}

func (s *Session) Reset() {
	s.authenticated = false
}

func (s *Session) Logout() error {
	s.authenticated = false
	return nil
}

// AuthMechanisms 实现 AuthSession 接口
func (s *Session) AuthMechanisms() []string {
	return []string{"PLAIN"}
}

// Auth 实现 AuthSession 接口
func (s *Session) Auth(mechanism string) (sasl.Server, error) {
	log.Printf("客户端请求的认证机制: %s", mechanism)
	switch mechanism {
	case "PLAIN":
		return sasl.NewPlainServer(func(identity, username, password string) error {
			return s.authenticate(username, password)
		}), nil
	default:
		return nil, smtp.ErrAuthUnsupported
	}
}

// authenticate 辅助函数，用于验证用户名和密码
func (s *Session) authenticate(username, password string) error {
	// 判断邮箱地址是否有效
	if !utils.IsValidEmail(username) {
		return errors.New("invalid email address")
	}
	ad, err := dao.IsAccountExist(username, strings.Split(username, "@")[1])
	if err != nil {
		return err
	}
	// 验证密码
	j := utils.CheckPasswordHash(password, ad.PasswordHash)
	if j {
		s.authenticated = true
		s.from = ad.EmailAddress
		return nil
	}
	return errors.New("invalid password")

}

func SmtpServerInit() {
	// 加载共享的 TLS 证书
	cert, err := tls.LoadX509KeyPair(
		"/etc/letsencrypt/live/smtp.xxx.net/fullchain.pem",
		"/etc/letsencrypt/live/smtp.xxx.net/privkey.pem")
	if err != nil {
		global.Log.Fatalf("加载 TLS 证书失败: %v", err)
		return
	}

	// 共享的服务器配置
	serverConfig := func(addr string) *smtp.Server {
		be := &Backend{}
		s := smtp.NewServer(be)
		s.TLSConfig = &tls.Config{Certificates: []tls.Certificate{cert}}
		s.Addr = addr
		s.Domain = "smtp.xxx.net"
		s.ReadTimeout = 30 * time.Second
		s.WriteTimeout = 30 * time.Second
		s.MaxMessageBytes = 35 * 1024 * 1024
		s.MaxRecipients = 50
		s.AllowInsecureAuth = false
		s.Debug = os.Stdout
		return s
	}

	// 启动 17896 端口
	go func() {
		s := serverConfig("0.0.0.0:17896")
		global.Log.Infof("SMTP 服务器启动在端口 17896")
		if err := s.ListenAndServe(); err != nil {
			global.Log.Errorf("端口 17896 服务器错误: %v", err)
		}
	}()

	// 启动 587 端口
	go func() {
		s := serverConfig("0.0.0.0:587")
		global.Log.Infof("SMTP 服务器启动在端口 587")
		if err := s.ListenAndServe(); err != nil {
			global.Log.Errorf("端口 587 服务器错误: %v", err)
		}
	}()
}
