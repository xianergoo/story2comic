package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"novelforge/internal/ai"
	"novelforge/internal/model"

	"gorm.io/gorm"
)

type PanelScript struct {
	Panel    int    `json:"panel"`
	Scene    string `json:"scene"`
	Action   string `json:"action"`
	Dialogue string `json:"dialogue"`
	Camera   string `json:"camera"`
}

type ComicScript struct {
	Panels []PanelScript `json:"panels"`
}

type ComicService struct {
	db       *gorm.DB
	provider ai.Provider
	imageDir string
}

func NewComicService(db *gorm.DB, imageDir string) *ComicService {
	return &ComicService{db: db, imageDir: imageDir}
}

func (s *ComicService) SetProvider(p ai.Provider) { s.provider = p }

func (s *ComicService) Generate(chapter *model.Chapter, novel *model.Novel, outline *model.Outline) error {
	if novel.ImageMode == "single" {
		return s.generateSingle(chapter, novel)
	}
	return s.generateMulti(chapter, novel, outline)
}

func (s *ComicService) generateSingle(chapter *model.Chapter, novel *model.Novel) error {
	const maxPromptLen = 500
	content := chapter.Content
	if len(content) > maxPromptLen {
		content = content[:maxPromptLen]
	}
	prompt := fmt.Sprintf("小说《%s》第%d章插画。根据内容生成一张漫画风格插画：%s",
		novel.Title, chapter.ChapterNo, content)

	resp, err := s.provider.GenerateImage(ai.ImageRequest{Prompt: prompt, Size: "1024x1024", N: 1})
	if err != nil {
		return err
	}
	imagePath, err := s.downloadImage(resp.URLs[0], novel.ID, chapter.ChapterNo, 1)
	if err != nil {
		return err
	}
	urlsJSON, _ := json.Marshal([]string{imagePath})

	page := &model.ComicPage{
		ChapterID:  chapter.ID,
		NovelID:    novel.ID,
		PageNo:     1,
		PanelCount: 1,
		ImageURLs:  string(urlsJSON),
		Status:     "done",
	}
	return s.db.Create(page).Error
}

func (s *ComicService) generateMulti(chapter *model.Chapter, novel *model.Novel, outline *model.Outline) error {
	script, err := s.generateScript(chapter)
	if err != nil {
		return err
	}

	var imagePaths []string
	for _, panel := range script.Panels {
		imgPrompt := fmt.Sprintf("漫画格 %d/%d：%s。动作：%s。镜头：%s。风格：日系黑白漫画。",
			panel.Panel, len(script.Panels), panel.Scene, panel.Action, panel.Camera)
		resp, err := s.provider.GenerateImage(ai.ImageRequest{Prompt: imgPrompt, Size: "1024x1024", N: 1})
		if err != nil {
			return err
		}
		path, err := s.downloadImage(resp.URLs[0], novel.ID, chapter.ChapterNo, panel.Panel)
		if err != nil {
			return err
		}
		imagePaths = append(imagePaths, path)
	}

	checkResult, err := s.comicCheck(chapter.Content, script, outline)
	if err == nil && !checkResult.Pass && checkResult.RetryPrompt != "" {
		resp, err := s.provider.GenerateImage(ai.ImageRequest{Prompt: checkResult.RetryPrompt, Size: "1024x1024", N: 1})
		if err == nil {
			path, _ := s.downloadImage(resp.URLs[0], novel.ID, chapter.ChapterNo, 0)
			imagePaths = append(imagePaths, path)
		}
	}

	scriptJSON, _ := json.Marshal(script)
	urlsJSON, _ := json.Marshal(imagePaths)
	page := &model.ComicPage{
		ChapterID:  chapter.ID,
		NovelID:    novel.ID,
		PageNo:     1,
		PanelCount: len(script.Panels),
		Script:     string(scriptJSON),
		ImageURLs:  string(urlsJSON),
		Status:     "done",
	}
	return s.db.Create(page).Error
}

func (s *ComicService) generateScript(chapter *model.Chapter) (*ComicScript, error) {
	systemPrompt := `你是专业漫画分镜师。将小说段落转化为 4 格漫画分镜脚本。
输出纯 JSON（不要 markdown 包裹）：
{"panels":[{"panel":1,"scene":"场景描述","action":"动作","dialogue":"对话","camera":"特写/中景/远景"}]}`

	const maxContent = 800
	content := chapter.Content
	if len(content) > maxContent {
		content = content[:maxContent]
	}

	resp, err := s.provider.Chat(ai.ChatRequest{
		Messages: []ai.ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: "章节内容：" + content},
		},
	})
	if err != nil {
		return nil, err
	}
	var script ComicScript
	json.Unmarshal([]byte(resp.Content), &script)
	if len(script.Panels) == 0 {
		script.Panels = []PanelScript{{Panel: 1, Scene: content, Camera: "中景"}}
	}
	return &script, nil
}

func (s *ComicService) comicCheck(sourceText string, script *ComicScript, outline *model.Outline) (*ComicCheckResult, error) {
	scriptJSON, _ := json.Marshal(script)
	systemPrompt := `你是专业漫画编辑。检查漫画分镜是否准确还原小说场景。
维度：场景还原、人物外观、分镜连贯、画风统一。
输出纯 JSON：{"pass":bool,"issues":["具体问题"],"retry_prompt":"修正后的生图 prompt（如果不通过）"}`

	prompt := fmt.Sprintf("原文段落：%s\n\n人物设定：%s\n\n当前分镜脚本：%s",
		sourceText, outline.CharacterSheets, string(scriptJSON))

	resp, err := s.provider.Chat(ai.ChatRequest{
		Messages: []ai.ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		return nil, err
	}
	var result ComicCheckResult
	json.Unmarshal([]byte(resp.Content), &result)
	return &result, nil
}

func (s *ComicService) downloadImage(url string, novelID uint, chapterNo, pageNo int) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	dir := filepath.Join(s.imageDir, fmt.Sprintf("%d", novelID))
	os.MkdirAll(dir, 0755)
	filename := fmt.Sprintf("%d_%d.png", chapterNo, pageNo)
	fullPath := filepath.Join(dir, filename)

	f, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	io.Copy(f, resp.Body)

	return fmt.Sprintf("%d/%s", novelID, filename), nil
}
