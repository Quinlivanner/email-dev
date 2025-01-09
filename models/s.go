package models

import "time"

type AwsSqsEvent struct {
	Records []struct {
		EventVersion string    `json:"eventVersion"`
		EventSource  string    `json:"eventSource"`
		AwsRegion    string    `json:"awsRegion"`
		EventTime    time.Time `json:"eventTime"`
		EventName    string    `json:"eventName"`
		UserIdentity struct {
			PrincipalId string `json:"principalId"`
		} `json:"userIdentity"`
		RequestParameters struct {
			SourceIPAddress string `json:"sourceIPAddress"`
		} `json:"requestParameters"`
		ResponseElements struct {
			XAmzRequestId string `json:"x-amz-request-id"`
			XAmzId2       string `json:"x-amz-id-2"`
		} `json:"responseElements"`
		S3 struct {
			S3SchemaVersion string `json:"s3SchemaVersion"`
			ConfigurationId string `json:"configurationId"`
			Bucket          struct {
				Name          string `json:"name"`
				OwnerIdentity struct {
					PrincipalId string `json:"principalId"`
				} `json:"ownerIdentity"`
				Arn string `json:"arn"`
			} `json:"bucket"`
			Object struct {
				Key       string `json:"key"`
				Size      int    `json:"size"`
				ETag      string `json:"eTag"`
				Sequencer string `json:"sequencer"`
			} `json:"object"`
		} `json:"s3"`
	} `json:"Records"`
}

// EmailContent 表示邮件内容
type EmailContent struct {
	From     string
	To       string
	Subject  string
	TextBody string
	HtmlBody string
	MailDir  string
}

type AwsSnsEvent struct {
	NotificationType string `json:"notificationType"`
	Mail             struct {
		Timestamp        time.Time `json:"timestamp"`
		Source           string    `json:"source"`
		MessageId        string    `json:"messageId"`
		Destination      []string  `json:"destination"`
		HeadersTruncated bool      `json:"headersTruncated"`
		Headers          []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"headers"`
		CommonHeaders struct {
			ReturnPath string   `json:"returnPath"`
			From       []string `json:"from"`
			Date       string   `json:"date"`
			To         []string `json:"to"`
			Cc         []string `json:"cc"`
			MessageId  string   `json:"messageId"`
			Subject    string   `json:"subject"`
		} `json:"commonHeaders"`
	} `json:"mail"`
	Receipt struct {
		Timestamp            time.Time `json:"timestamp"`
		ProcessingTimeMillis int       `json:"processingTimeMillis"`
		Recipients           []string  `json:"recipients"`
		SpamVerdict          struct {
			Status string `json:"status"`
		} `json:"spamVerdict"`
		VirusVerdict struct {
			Status string `json:"status"`
		} `json:"virusVerdict"`
		SpfVerdict struct {
			Status string `json:"status"`
		} `json:"spfVerdict"`
		DkimVerdict struct {
			Status string `json:"status"`
		} `json:"dkimVerdict"`
		DmarcVerdict struct {
			Status string `json:"status"`
		} `json:"dmarcVerdict"`
		Action struct {
			Type            string `json:"type"`
			TopicArn        string `json:"topicArn"`
			BucketName      string `json:"bucketName"`
			ObjectKeyPrefix string `json:"objectKeyPrefix"`
			ObjectKey       string `json:"objectKey"`
		} `json:"action"`
	} `json:"receipt"`
}

type Recipients struct {
	To         []string
	Cc         []string
	Bcc        []string
	Recipients []string
}

type EmailAddress struct {
	DisplayName string
	Address     string
}
