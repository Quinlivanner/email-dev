package service

import (
	"email/global"
	"encoding/json"
	"github.com/levigross/grequests"
)

func GetAiEmailReply(subject, textContent string) (string, error) {
	reqOption := &grequests.RequestOptions{
		JSON: req{
			Message: []struct {
				Content string `json:"content"`
				Role    string `json:"role"`
			}{
				{
					Content: global.Config.RequestsApi.AIPrompt,
					Role:    "system",
				},
				{
					Content: "Subject: " + subject + "\\n" + "textContent:" + textContent,
					Role:    "user",
				},
			},
			Model:  global.Config.RequestsApi.AIModelName,
			Stream: false,
		},
		Headers: map[string]string{
			"Authorization": "Bearer " + global.Config.RequestsApi.AIChatKey,
			"Content-Type":  "application/json",
			"Accept":        "application/json",
		},
	}
	get, err := grequests.Get(global.Config.RequestsApi.AIChatAPI, reqOption)
	if err != nil {
		global.Log.Error("AI请求失败", err)
		return "", err
	}
	var aw aiAnswer
	err = json.Unmarshal(get.Bytes(), &aw)
	if err != nil {
		global.Log.Error("序列化失败", err)
		return "", err
	}
	return aw.Choices[0].Message.Content, nil
}

type req struct {
	Message []struct {
		Content string `json:"content"`
		Role    string `json:"role"`
	} `json:"messages"`
	Model  string `json:"model"`
	Stream bool   `json:"stream"`
}

type aiAnswer struct {
	Id      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	SystemFingerprint string `json:"system_fingerprint"`
}
