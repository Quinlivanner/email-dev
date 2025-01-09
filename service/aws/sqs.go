package aws

import (
	"context"
	"email/global"
	"email/utils"
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	"email/models"
)

// 处理 SQS 邮件消息
func ProcessSQSEmailMessages() {
	global.Log.Info("开始处理 SQS 邮件消息...")
	// 创建上下文
	ctx := context.TODO()

	// 加载 AWS 配置（从默认位置）
	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(global.Config.AWS.ConfigRegion),
		config.WithSharedConfigProfile(global.Config.AWS.ConfigProfile),
	)
	if err != nil {
		log.Fatalf("无法加载 AWS 配置: %v", err)
	}

	// 创建 SQS 客户端
	client := sqs.NewFromConfig(cfg)

	// 指定队列的 URL
	queueURL := global.Config.AWS.SQSUrl

	// 设置接收消息的参数
	receiveParams := &sqs.ReceiveMessageInput{
		QueueUrl:            &queueURL,
		MaxNumberOfMessages: 10,
		WaitTimeSeconds:     10,
	}

	// 接收消息
	for {
		msgs, err := client.ReceiveMessage(ctx, receiveParams)
		if err != nil {
			log.Printf("接收消息失败: %v", err)
			continue
		}
		if len(msgs.Messages) == 0 {
			continue
		}
		for _, message := range msgs.Messages {
			var s = models.AwsSnsEvent{}
			fmt.Printf("收到消息: %s\n", *message.Body)
			//把消息解析到结构体
			err = json.Unmarshal([]byte(*message.Body), &s)
			if err != nil {
				global.Log.Errorf("解析消息失败: %v", err)
				continue
			}
			if len(s.NotificationType) == 0 {
				global.Log.Error("解析消息失败: 缺少必要字段")
				DeleteQueueMessage(queueURL, message.ReceiptHandle, client, ctx)
				continue
			}
			if !VerifyEvent(&s) {
				global.Log.Error("事件验证失败")
				DeleteQueueMessage(queueURL, message.ReceiptHandle, client, ctx)
				continue
			}

			recipients := ClassifyRecipients(&s)
			b, _ := json.Marshal(recipients)
			global.Log.Error(string(b))
			//fmt.Println(string(b))

			//处理邮件
			err = S3EmailDataProcessor(s.Receipt.Action.ObjectKey, &recipients)
			if err != nil {
				fmt.Println(err)
				if err == gorm.ErrRecordNotFound {
					DeleteQueueMessage(queueURL, message.ReceiptHandle, client, ctx)
				}
				global.Log.Errorf("处理邮件失败: %v", err)
			} else {
				DeleteQueueMessage(queueURL, message.ReceiptHandle, client, ctx)
			}
		}
	}
}

func DeleteQueueMessage(queueURL string, ReceiptHandle *string, client *sqs.Client, ctx context.Context) {
	//删除已处理的消息
	deleteParams := &sqs.DeleteMessageInput{
		QueueUrl:      &queueURL,
		ReceiptHandle: ReceiptHandle,
	}

	_, err := client.DeleteMessage(ctx, deleteParams)
	if err != nil {
		log.Printf("删除消息失败: %v", err)
	} else {
		fmt.Println("消息已删除")
	}
}

func ClassifyRecipients(event *models.AwsSnsEvent) models.Recipients {
	// 1. 解析 To 和 Cc 地址
	explicitTo := utils.ParseAddressList(event.Mail.CommonHeaders.To)
	explicitCc := utils.ParseAddressList(event.Mail.CommonHeaders.Cc)

	// 2. 创建查找映射用于识别 BCC
	explicitRecipients := make(map[string]bool)
	for _, addr := range explicitTo {
		explicitRecipients[addr] = true
	}
	for _, addr := range explicitCc {
		explicitRecipients[addr] = true
	}

	// 3. 找出 Bcc（Receipt.Recipients 中的地址已经是纯邮件地址）
	var bccList []string
	for _, addr := range event.Receipt.Recipients {
		if !explicitRecipients[addr] {
			bccList = append(bccList, addr)
		}
	}

	// 4. 安全地合并所有收件人
	allRecipients := make([]string, 0)
	allRecipients = append(allRecipients, explicitTo...)
	allRecipients = append(allRecipients, explicitCc...)
	allRecipients = append(allRecipients, bccList...)

	return models.Recipients{
		To:         explicitTo,
		Cc:         explicitCc,
		Bcc:        bccList,
		Recipients: allRecipients,
	}
}

// 验证事件
func VerifyEvent(event *models.AwsSnsEvent) bool {
	if event.NotificationType != "Received" {
		return false
	}
	if event.Receipt.Action.Type != "S3" ||
		event.Receipt.Action.ObjectKeyPrefix != "email/" ||
		event.Receipt.Action.BucketName != global.Config.AWS.S3Bucket {
		return false
	}
	return true
}
