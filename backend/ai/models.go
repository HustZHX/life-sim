package ai

import "fmt"

// ModelOption 可选 AI 模型（对外展示）
type ModelOption struct {
	ID          string `json:"id"`
	APIModel    string `json:"api_model"`
	Label       string `json:"label"`
	Description string `json:"description"`
	EstTime     string `json:"est_time"`
	Speed       string `json:"speed"`
	Quality     string `json:"quality"`
}

var modelCatalog = []ModelOption{
	{
		ID: "flash", APIModel: "deepseek-v4-flash",
		Label:       "V4 Flash（推荐）",
		Description: "轻量高速模型，响应最快，适合大多数人生时间轴生成与日常编辑重算。",
		EstTime:     "全时间轴约 20 秒～1 分钟",
		Speed:       "最快",
		Quality:     "良好",
	},
	{
		ID: "pro", APIModel: "deepseek-v4-pro",
		Label:       "V4 Pro",
		Description: "旗舰模型，逻辑与文笔最佳，长人物时间轴更细腻，但耗时最长。",
		EstTime:     "全时间轴约 1～3 分钟",
		Speed:       "较慢",
		Quality:     "最佳",
	},
}

func ListModels() []ModelOption {
	out := make([]ModelOption, len(modelCatalog))
	copy(out, modelCatalog)
	return out
}

// DefaultModelID 用户未指定模型时的默认选项。
const DefaultModelID = "flash"

// NormalizeModelID 将请求或历史任务中的模型 ID 规范为 flash / pro。
func NormalizeModelID(modelID string) string {
	if modelID == "pro" || modelID == "deepseek-v4-pro" {
		return "pro"
	}
	return DefaultModelID
}

func ResolveAPIModel(modelID string) (string, error) {
	if modelID == "" {
		modelID = DefaultModelID
	}
	for _, m := range modelCatalog {
		if m.ID == modelID || m.APIModel == modelID {
			return m.APIModel, nil
		}
	}
	return "", fmt.Errorf("不支持的模型: %s（仅支持 flash、pro）", modelID)
}
