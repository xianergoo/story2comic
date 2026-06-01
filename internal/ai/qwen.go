package ai

import "net/http"

type QwenProvider struct {
	OpenAIProvider
}

func newQwen(apiKey, baseURL, textModel, imageModel string) *QwenProvider {
	if baseURL == "" {
		baseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	}
	return &QwenProvider{
		OpenAIProvider: OpenAIProvider{
			apiKey:     apiKey,
			baseURL:    baseURL,
			textModel:  textModel,
			imageModel: imageModel,
			client:     &http.Client{},
		},
	}
}
