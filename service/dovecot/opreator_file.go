package dovecot

import (
	"bytes"
	"email/global"
	"email/models"
	"errors"
	"fmt"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"time"
)

type ImapConfig struct {
	Server   string
	Port     int // 使用 143 作为非加密端口
	Username string
	Password string
}

type MailMover struct {
	config ImapConfig
	client *client.Client
}

func NewMailMover(config ImapConfig) *MailMover {
	return &MailMover{
		config: config,
	}
}

// Imap连接
func (m *MailMover) Connect() error {
	fmt.Printf("Attempting to connect to %s:%d\n", m.config.Server, m.config.Port)

	c, err := client.DialTLS(fmt.Sprintf("%s:%d", m.config.Server, m.config.Port), nil)
	if err != nil {
		return fmt.Errorf("failed to connect to IMAP server: %v", err)
	}
	fmt.Println("Connection established")

	fmt.Printf("Attempting to login as %s\n", m.config.Username)
	if err := c.Login(m.config.Username, m.config.Password); err != nil {
		c.Logout()
		return fmt.Errorf("failed to login: %v", err)
	}
	fmt.Println("Login successful")

	m.client = c
	return nil
}

// 根据msgid获取邮件uid
func (m *MailMover) FindEmailByMessageID(mailbox, messageID string) ([]uint32, error) {
	_, err := m.client.Select(mailbox, true)

	if err != nil {
		return nil, err
	}

	criteria := imap.NewSearchCriteria()
	criteria.Header.Add("Message-ID", messageID)

	uids, err := m.client.Search(criteria)
	if err != nil {
		return nil, err
	}

	return uids, nil
}

// 移动邮件
func (m *MailMover) MoveEmails(sourceMailbox, targetMailbox string, uid uint32) error {
	if _, err := m.client.Select(sourceMailbox, false); err != nil {
		return fmt.Errorf("failed to select mailbox %s: %v", sourceMailbox, err)
	}
	seqSet := new(imap.SeqSet)
	seqSet.AddNum(uid)
	if err := m.client.Move(seqSet, targetMailbox); err != nil {
		return err
	}
	return nil
}

// 设置邮件已读未读状态
func (m *MailMover) SetReadStatus(mailbox string, uid uint32, read bool) error {
	_, err := m.client.Select(mailbox, false)
	if err != nil {
		return fmt.Errorf("failed to select mailbox: %v", err)
	}

	seqSet := new(imap.SeqSet)
	seqSet.AddNum(uid)

	if read {
		return m.client.Store(seqSet, imap.AddFlags, []interface{}{imap.SeenFlag}, nil)
	}
	return m.client.Store(seqSet, imap.RemoveFlags, []interface{}{imap.SeenFlag}, nil)

}

// 根据messageid查找邮件是否存在
func (m *MailMover) IsEmailExist(mailbox, messageID string) (bool, error) {
	_, err := m.client.Select(mailbox, true)
	if err != nil {
		return false, err
	}

	criteria := imap.NewSearchCriteria()
	criteria.Header.Add("Message-ID", messageID)

	uids, err := m.client.Search(criteria)
	if err != nil {
		return false, err
	}

	return len(uids) > 0, nil
}

// 添加邮件到本地
func (m *MailMover) AppendRawEmail(mailbox string, rawMessage []byte) error {
	// 选择邮箱
	if _, err := m.client.Select(mailbox, false); err != nil {
		return err
	}

	flags := []string{}

	err := m.client.Append(mailbox, flags, time.Now(), bytes.NewReader(rawMessage))
	if err != nil {
		return err
	}

	return nil
}

func (m *MailMover) Close() error {
	if m.client != nil {
		return m.client.Logout()
	}
	return nil
}

func ImapOperatorEmail(option models.ImapOperatorOptions) error {
	config := ImapConfig{
		Server:   "imap.net",
		Port:     993,
		Username: option.UserName + "*testuser",
		Password: "test",
	}

	mover := NewMailMover(config)
	if err := mover.Connect(); err != nil {
		return errors.New(fmt.Sprintf("连接 Imap 服务器失败，原因：%s", err))
	}
	defer mover.Close()

	if option.MoveType == global.AppendEmail {
		exist, err := mover.IsEmailExist(option.TargetMailBox, option.MessageID)
		if err != nil {
			return err
		}
		if !exist {
			return mover.AppendRawEmail(option.TargetMailBox, option.EmailRawMessage)
		}
		return nil
	}

	ids, err := mover.FindEmailByMessageID(option.SourceMailBox, option.MessageID)
	if err != nil || len(ids) == 0 {
		return errors.New(fmt.Sprintf("查询邮件 MessageID 失败，原因：%s ｜ 原始邮箱标记：%s ｜ Meessage-ID：%s", err, option.SourceMailBox, option.MessageID))
	}

	if option.MoveType == global.MoveEmail {
		return mover.MoveEmails(option.SourceMailBox, option.TargetMailBox, ids[0])
	}

	if option.MoveType == global.ChangeReadStatus {
		return mover.SetReadStatus(option.SourceMailBox, ids[0], option.ReadStatus)
	}

	return nil

}
