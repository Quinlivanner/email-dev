package dovecot

import (
	"bufio"
	"email/dao"
	"email/global"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type ProcessState struct {
	LastProcessedTime time.Time `json:"last_processed_time"`
}

type EmailFlagChange struct {
	MessageID    string
	FromBox      string
	Box          string
	UID          string
	Flags        []string
	Action       string
	EmailAddress string // 添加邮箱地址字段
}

func getStateFilePath() string {
	dir, err := os.Getwd()
	if err != nil {
		global.Log.Errorf("获取当前目录失败: %v\n", err)
		return "email_monitor_state.json"
	}
	return filepath.Join(dir, "email_monitor_state.json")
}

func loadState() ProcessState {
	stateFile := getStateFilePath()
	data, err := os.ReadFile(stateFile)
	if err != nil {
		global.Log.Errorf("读取状态文件失败: %v\n", err)
		return ProcessState{
			LastProcessedTime: time.Now(),
		}
	}

	var state ProcessState
	if err := json.Unmarshal(data, &state); err != nil {
		global.Log.Errorf("解析状态文件失败: %v\n", err)
		return ProcessState{
			LastProcessedTime: time.Now(),
		}
	}
	return state
}

func saveState(state ProcessState) error {
	stateFile := getStateFilePath()
	data, err := json.MarshalIndent(state, "", "    ")
	if err != nil {
		return fmt.Errorf("序列化状态失败: %v", err)
	}

	err = os.WriteFile(stateFile, data, 0644)
	if err != nil {
		return fmt.Errorf("写入状态文件失败: %v", err)
	}

	return nil
}

func verifyState(expectedTime time.Time) {
	state := loadState()
	if !state.LastProcessedTime.Equal(expectedTime) {
		global.Log.Errorf("警告: 状态保存验证失败\n期望时间: %v\n实际时间: %v\n",
			expectedTime, state.LastProcessedTime)
	}
}

func parseFlagChange(line string, regex *regexp.Regexp) *EmailFlagChange {
	matches := regex.FindStringSubmatch(line)
	if matches == nil {
		return nil
	}

	//fmt.Printf("Debug - 正则匹配结果: %v\n", matches)

	// 提取邮箱地址
	emailRegex := regexp.MustCompile(`imap\(([^@]+@[^)]+)\)`)
	emailMatches := emailRegex.FindStringSubmatch(line)
	emailAddress := ""
	if len(emailMatches) > 1 {
		emailAddress = emailMatches[1]
	}

	// 确定操作类型
	action := ""
	if strings.Contains(line, "flag_change") {
		action = "flag_change"
	} else if strings.Contains(line, "copy from") {
		action = "move"
	} else if strings.Contains(line, "delete") {
		action = "delete"
	} else if strings.Contains(line, "expunge") {
		if strings.Contains(line, "\\Deleted") {
			action = "permanent_delete" // 彻底删除
		} else {
			action = "move_complete" // 移动完成
		}
	}

	var flags []string
	if len(matches) > 5 && matches[5] != "" {
		flagStr := strings.TrimSpace(matches[5])
		if flagStr != "" {
			flags = strings.Split(flagStr, " ")
			for i, flag := range flags {
				flags[i] = strings.TrimSpace(flag)
			}
		}
	}

	return &EmailFlagChange{
		FromBox:      matches[1],
		Box:          matches[2],
		UID:          matches[3],
		MessageID:    matches[4],
		Flags:        flags,
		Action:       action,
		EmailAddress: emailAddress,
	}
}

func getBoxName(box string) string {
	switch box {
	case "INBOX":
		return "inbox"
	case "Trash":
		return "deleted"
	case "Drafts":
		return "draft"
	case "Sent":
		return "sent"
	case "Junk":
		return "trash"
	default:
		return "inbox"
	}
}

func printFlagChange(change *EmailFlagChange) {
	fmt.Printf("\n时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("邮箱: %s\n", change.EmailAddress)
	fmt.Printf("Message-ID: %s\n", change.MessageID)

	fmt.Printf("Debug - 原始标记: %v\n", change.Flags)
	for i, flag := range change.Flags {
		fmt.Printf("Debug - 标记 %d: [%s]\n", i, flag)
	}

	// 打印操作类型
	switch change.Action {
	case "move":
		err := UpdateEmailType(change.MessageID, change.EmailAddress, getBoxName(change.Box))
		if err != nil {
			global.Log.Errorf("操作: 从 %s 移动到 %s，操作状态：%s\n", getBoxName(change.FromBox), getBoxName(change.Box), err)
		}
		return
	//case "move_complete":
	//	fmt.Printf("操作: 移动完成\n")
	//case "delete":
	//	fmt.Printf("操作: 标记删除 - 从 %s\n", getBoxName(change.Box))
	case "permanent_delete":
		err := DeleteEmail(change.MessageID, change.EmailAddress)
		if err != nil {
			global.Log.Errorf("操作: 彻底删除 - 从 %s，操作状态：%s\n", getBoxName(change.Box), err)
		}
		return
		//case "flag_change":
		//	fmt.Printf("操作: 标记变更\n")
	}

	// 打印目标位置
	fmt.Printf("位置: %s\n", getBoxName(change.Box))

	// 打印标记
	if len(change.Flags) > 0 {
		hasSeenFlag := false
		//hasDeletedFlag := false
		//hasDraftFlag := false

		for _, flag := range change.Flags {
			switch flag {
			case "\\Seen":
				hasSeenFlag = true
				//case "\\Deleted":
				//	hasDeletedFlag = true
				//case "\\Draft":
				//	hasDraftFlag = true
			}
		}

		// 输出状态
		if hasSeenFlag {
			err := UpdateEmailReadStatus(change.MessageID, change.EmailAddress, true)
			if err != nil {
				global.Log.Errorf("操作: 标记已读，操作状态：%s\n", err)
				return
			}
		} else {
			err := UpdateEmailReadStatus(change.MessageID, change.EmailAddress, false)
			if err != nil {
				global.Log.Errorf("操作: 标记未读，操作状态：%s\n", err)
				return
			}
		}
		//if hasDeletedFlag {
		//	fmt.Println("状态: 已删除")
		//}
		//if hasDraftFlag {
		//	fmt.Println("状态: 草稿")
		//}
	} else {
		err := UpdateEmailReadStatus(change.MessageID, change.EmailAddress, false)
		if err != nil {
			global.Log.Errorf("操作: 标记未读，操作状态：%s\n", err)
			return
		}
		fmt.Println("状态: 未读")
	}
	fmt.Println(strings.Repeat("-", 50))
}

func DovecotStatusMonitorInit() {
	logPath := "/var/log/dovecot/info.log"

	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		fmt.Printf("日志文件不存在: %s\n", logPath)
		return
	}

	state := loadState()
	fmt.Printf("加载状态完成，上次处理时间: %v\n", state.LastProcessedTime)

	cmd := exec.Command("tail", "-F", logPath)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("创建管道错误: %v\n", err)
		return
	}

	if err := cmd.Start(); err != nil {
		fmt.Printf("启动tail命令错误: %v\n", err)
		return
	}

	flagChangeRegex := regexp.MustCompile(`(?:flag_change|copy from ([^:]+)|expunge): box=([^,]+), uid=(\d+), msgid=(<[^>]+>).*?flags=\((.*?)\)`)

	fmt.Println("开始监控邮件状态变化...")

	reader := bufio.NewReader(stdout)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			continue
		}

		// 解析日志时间
		currentYear := time.Now().Year()
		logTime, err := time.Parse("Jan 2 15:04:05 2006", line[:15]+fmt.Sprintf(" %d", currentYear))
		if err != nil {
			continue
		}

		if logTime.After(state.LastProcessedTime) {
			if strings.Contains(line, "flag_change") ||
				strings.Contains(line, "move") ||
				strings.Contains(line, "copy") ||
				strings.Contains(line, "expunge") {
				if change := parseFlagChange(line, flagChangeRegex); change != nil {
					printFlagChange(change)

					// 更新并保存状态
					state.LastProcessedTime = logTime
					if err := saveState(state); err != nil {
						fmt.Printf("保存状态失败: %v\n", err)
					} else {
						verifyState(state.LastProcessedTime)
					}
				}
			}
		}
	}

	if err := cmd.Wait(); err != nil {
		fmt.Printf("tail命令结束错误: %v\n", err)
	}
}

// 将邮件移动到指定分类
func UpdateEmailType(emailMessageID, emailAddress string, targetType string) error {
	accountId, err := dao.GetAccountIDByEmailAddress(emailAddress)
	if err != nil {
		return err
	}
	//根据messageid和emailaddress更新邮件分类
	err = dao.MoveEmailDirect(emailMessageID, accountId, targetType)
	if err != nil {
		return err
	}
	return nil

}

// 删除邮件
func DeleteEmail(emailMessageID, emailAddress string) error {
	accountId, err := dao.GetAccountIDByEmailAddress(emailAddress)
	if err != nil {
		return err
	}
	err = dao.DeleteEmailByMessageID(accountId, emailMessageID)
	if err != nil {
		return err
	}
	return nil

}

// 更新邮件阅读状态
func UpdateEmailReadStatus(emailMessageID, emailAddress string, readStatus bool) error {
	accountId, err := dao.GetAccountIDByEmailAddress(emailAddress)
	if err != nil {
		return err
	}
	err = dao.UpdateEmailReadStatusByMessageID(accountId, emailMessageID, readStatus)
	if err != nil {
		return err
	}
	return nil
}
