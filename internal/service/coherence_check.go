package service

import (
	"encoding/json"
	"novelforge/internal/ai"
)

type CheckResult struct {
	Pass       bool     `json:"pass"`
	Score      int      `json:"score"`
	Issues     []string `json:"issues"`
	Suggestion string   `json:"suggestion"`
}

type ComicCheckResult struct {
	Pass        bool     `json:"pass"`
	Issues      []string `json:"issues"`
	RetryPrompt string   `json:"retry_prompt"`
}

func (s *OutlineService) CheckCoherence(
	previousSummaries string,
	characterSheets string,
	worldSetting string,
	chapterContent string,
) (*CheckResult, error) {
	systemPrompt := `你是一个专业的小说编辑，审查章节剧情连贯性。
检查维度：
1. 人物一致性：性格、关系、已知经历是否前后矛盾
2. 情节逻辑：事件因果链是否断裂
3. 世界观冲突：设定是否冲突

输出纯 JSON（不要 markdown 包裹）：
{"pass":bool,"score":0-10,"issues":["具体问题描述"],"suggestion":"修改建议"}`

	prompt := "前文摘要：" + previousSummaries +
		"\n人物设定：" + characterSheets +
		"\n世界观：" + worldSetting +
		"\n当前章节：" + chapterContent

	resp, err := s.provider.Chat(ai.ChatRequest{
		Messages: []ai.ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		return nil, err
	}

	var result CheckResult
	json.Unmarshal([]byte(resp.Content), &result)
	return &result, nil
}
