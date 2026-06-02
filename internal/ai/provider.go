package ai

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

type ChatResponse struct {
	Content string `json:"content"`
}

type ImageRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Size   string `json:"size"`
	N      int    `json:"n"`
}

type ImageResponse struct {
	URLs []string `json:"urls"`
}

type StreamChunk struct {
	Content string
	Done    bool
	Error   string
}

type Provider interface {
	Chat(req ChatRequest) (*ChatResponse, error)
	ChatStream(req ChatRequest) (<-chan StreamChunk, error)
	GenerateImage(req ImageRequest) (*ImageResponse, error)
}

func NewProvider(providerType, apiKey, baseURL, textModel, imageModel string) Provider {
	switch providerType {
	case "openai":
		return newOpenAI(apiKey, baseURL, textModel, imageModel)
	case "qwen":
		return newQwen(apiKey, baseURL, textModel, imageModel)
	case "custom":
		return newCustom(apiKey, baseURL, textModel, imageModel)
	default:
		return newOpenAI(apiKey, baseURL, textModel, imageModel)
	}
}
