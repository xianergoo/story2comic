package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"novelforge/internal/ai"
	"novelforge/internal/model"

	"gorm.io/gorm"
)

type OutlineService struct {
	db       *gorm.DB
	provider ai.Provider
}

func NewOutlineService(db *gorm.DB) *OutlineService { return &OutlineService{db: db} }

func (s *OutlineService) SetProvider(p ai.Provider) { s.provider = p }

func (s *OutlineService) GetByNovel(novelID uint) (*model.Outline, error) {
	var o model.Outline
	err := s.db.Where("novel_id = ?", novelID).Order("version DESC").First(&o).Error
	return &o, err
}

type OutlineResult struct {
	Content         string           `json:"content"`
	CharacterSheets CharacterSheets  `json:"character_sheets"`
	WorldSetting    string           `json:"world_setting"`
	ChapterPlan     []ChapterPlanItem `json:"chapter_plan"`
}

type CharacterSheets map[string]CharacterSheet

type CharacterSheet struct {
	Name        string `json:"name"`
	Role        string `json:"role"`
	Personality string `json:"personality"`
	Appearance  string `json:"appearance"`
}

type ChapterPlanItem struct {
	ChapterNo int    `json:"chapter_no"`
	Title     string `json:"title"`
	Summary   string `json:"summary"`
}

func (s *OutlineService) Generate(novel *model.Novel, userInput string) (*model.Outline, error) {
	return s.GenerateStream(novel, userInput, func(chunk string, done bool) {})
}

func (s *OutlineService) GenerateStream(novel *model.Novel, userInput string, onChunk func(chunk string, done bool)) (*model.Outline, error) {
	systemPrompt := `你是一个专业的小说策划，请根据用户输入生成完整的故事大纲。

输出纯 JSON 格式（不要 markdown 包裹）：
{
  "content": "故事大纲全文",
  "character_sheets": {"角色名": {"name":"...","role":"主角/配角/反派","personality":"性格描述","appearance":"外貌描述"}},
  "world_setting": "世界观描述",
  "chapter_plan": [{"chapter_no": 1, "title": "第一章标题", "summary": "本章概要"}]
}
请生成 8-12 章的章节规划。`

	userPrompt := fmt.Sprintf("用户输入：%s", userInput)

	ch, err := s.provider.ChatStream(ai.ChatRequest{
		Messages: []ai.ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	})
	if err != nil {
		return nil, err
	}

	var full strings.Builder
	for chunk := range ch {
		if chunk.Error != "" {
			return nil, fmt.Errorf("stream error: %s", chunk.Error)
		}
		full.WriteString(chunk.Content)
		onChunk(chunk.Content, chunk.Done)
	}
	return s.parseAndSave(novel, full.String())
}

func (s *OutlineService) parseAndSave(novel *model.Novel, raw string) (*model.Outline, error) {
	var result OutlineResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, fmt.Errorf("failed to parse outline JSON: %w", err)
	}
	csJSON, _ := json.Marshal(result.CharacterSheets)
	cpJSON, _ := json.Marshal(result.ChapterPlan)
	outline := &model.Outline{
		NovelID:         novel.ID,
		Content:         result.Content,
		CharacterSheets: string(csJSON),
		WorldSetting:    result.WorldSetting,
		ChapterPlan:     string(cpJSON),
	}
	s.db.Create(outline)
	return outline, nil
}
