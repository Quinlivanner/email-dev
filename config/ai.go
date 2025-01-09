package config

type AIRequestsApi struct {
	AIChatAPI   string `yaml:"ai_chat_api"`
	AIChatKey   string `yaml:"ai_chat_key"`
	AIModelName string `yaml:"ai_model_name"`
	AIPrompt    string `yaml:"ai_prompt"`
}
