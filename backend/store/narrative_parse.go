package store

import (
	"encoding/json"
	"fmt"
	"unicode/utf8"
)

type LightNovelChapter struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func ParseLightNovelChapter(raw string) (*LightNovelChapter, error) {
	var ch LightNovelChapter
	if err := json.Unmarshal([]byte(raw), &ch); err != nil {
		return nil, err
	}
	if ch.Content == "" {
		return nil, fmt.Errorf("轻小说正文为空")
	}
	if ch.Title == "" {
		ch.Title = "未命名章节"
	}
	return &ch, nil
}

func ParseNarrativeContent(raw string) (string, error) {
	var resp struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return "", err
	}
	if resp.Content == "" {
		return "", fmt.Errorf("正文为空")
	}
	return resp.Content, nil
}

func TailRunes(s string, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}
	if utf8.RuneCountInString(s) <= maxRunes {
		return s
	}
	runes := []rune(s)
	return string(runes[len(runes)-maxRunes:])
}

func FormatLightNovelChapter(title, content string) string {
	return fmt.Sprintf("# %s\n\n%s", title, content)
}
