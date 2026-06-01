package ai

func newCustom(apiKey, baseURL, textModel, imageModel string) *OpenAIProvider {
	return newOpenAI(apiKey, baseURL, textModel, imageModel)
}
