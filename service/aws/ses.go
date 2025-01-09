package aws

import (
	"context"
	"email/global"
	"email/models"
	"encoding/json"
	"fmt"
	ses "github.com/aws/aws-sdk-go-v2/service/sesv2"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
)

// 调用SES发送新的邮件
func SendNewEmailByAwsSes(e *models.SendNewEmailRequest, senderEmailAddress, senderName string) error {
	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(global.Config.AWS.ConfigRegion),
		config.WithSharedConfigProfile(global.Config.AWS.ConfigProfile),
	)
	if err != nil {
		log.Fatalf("无法加载AWS配置: %v", err)
	}
	ses.NewDefaultEndpointResolverV2()
	sesClient := ses.NewFromConfig(cfg)
	sender := fmt.Sprintf("%s <%s>", senderName, senderEmailAddress)
	recipient := e.To
	subject := e.Subject

	input := &ses.SendEmailInput{
		Destination: &types.Destination{
			ToAddresses: []string{
				recipient,
			},
		},
		Content: &types.EmailContent{
			Simple: &types.Message{
				Body: &types.Body{
					Html: &types.Content{
						Charset: aws.String("UTF-8"),
						Data:    aws.String(string(e.HtmlBody)),
					},
					Text: &types.Content{
						Charset: aws.String("UTF-8"),
						Data:    aws.String(e.TextBody),
					},
				},
				Subject: &types.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(subject),
				},
			},
		},
		FromEmailAddress: aws.String(sender),
	}
	result, err := sesClient.SendEmail(context.TODO(), input)
	if err != nil {
		return err
	}
	marshal, err := json.Marshal(result)
	if err != nil {
		return err
	}
	fmt.Println(string(marshal))
	return nil
}

// 调用SES回复邮件
func ReplyEmailByAwsSes(e *models.ReplyEmailRequest, senderEmailAddress, senderName string) error {
	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(global.Config.AWS.ConfigRegion),
		config.WithSharedConfigProfile(global.Config.AWS.ConfigProfile),
	)
	if err != nil {
		log.Fatalf("无法加载AWS配置: %v", err)
	}
	sesClient := ses.NewFromConfig(cfg)
	sender := fmt.Sprintf("%s <%s>", senderName, senderEmailAddress)
	recipient := e.To
	subject := e.Subject

	input := &ses.SendEmailInput{
		Destination: &types.Destination{
			ToAddresses: []string{
				recipient,
			},
		},
		Content: &types.EmailContent{
			Simple: &types.Message{
				Body: &types.Body{
					Html: &types.Content{
						Charset: aws.String("UTF-8"),
						Data:    aws.String(string(e.HtmlBody)),
					},
					Text: &types.Content{
						Charset: aws.String("UTF-8"),
						Data:    aws.String(e.TextBody),
					},
				},
				Subject: &types.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(subject),
				},
			},
		},
		FromEmailAddress: aws.String(sender),
	}
	result, err := sesClient.SendEmail(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("发送邮件失败: %v", err)
	}

	log.Printf("邮件发送成功，消息ID: %s", *result.MessageId)
	return nil
}

func SendEmailByAwsSesWithRawMessage(rawMessage []byte, from string, to, cc, bcc []string) error {
	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(global.Config.AWS.ConfigRegion),
		config.WithSharedConfigProfile(global.Config.AWS.ConfigProfile),
	)
	if err != nil {
		return fmt.Errorf("无法加载AWS配置: %v", err)
	}
	sesClient := ses.NewFromConfig(cfg)

	input := &ses.SendEmailInput{
		Content: &types.EmailContent{
			Raw: &types.RawMessage{
				Data: rawMessage,
			},
		},
		Destination: &types.Destination{
			ToAddresses:  to,
			CcAddresses:  cc,
			BccAddresses: bcc,
		},
		FromEmailAddress: aws.String(from),
	}

	result, err := sesClient.SendEmail(ctx, input)
	if err != nil {
		return fmt.Errorf("发送邮件失败: %v", err)
	}

	log.Printf("原始邮件发送成功，消息ID: %s", *result.MessageId)
	return nil
}
