package models

// 文档根结构
type EngineResult struct {
	Document      []DocumentItem `json:"document"`
	EngineVersion string         `json:"engine_version"`
	EvalID        string         `json:"eval_id"`
	Image         []Image        `json:"image"`
	Version       string         `json:"version"`
}

// 文档项
type DocumentItem struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// 图像信息
type Image struct {
	Content [][]ContentNode `json:"content"`
	Height  int             `json:"height"`
	ID      string          `json:"id"`
	Width   int             `json:"width"`
}

// 通用内容节点（替换原来的ImageContent和ContentItem）
type ContentNode struct {
	Angle     float64         `json:"angle"`
	Attribute []Attribute     `json:"attribute"`
	Category  string          `json:"category"`
	Content   [][]ContentNode `json:"content,omitempty"`
	Contour   []Point         `json:"contour"`
	Coord     []Point         `json:"coord"`
	Direction string          `json:"direction,omitempty"`
	ID        string          `json:"id"`
	Score     float64         `json:"score,omitempty"`
	Text      []string        `json:"text,omitempty"`
	Type      string          `json:"type"`
}

// 文本单元
type TextUnit struct {
	Attribute []Attribute     `json:"attribute"`
	Category  string          `json:"category"`
	Content   [][][]PrintText `json:"content"`
	Contour   []Point         `json:"contour"`
	Coord     []Point         `json:"coord"`
	ID        string          `json:"id"`
	Text      []string        `json:"text"`
	Type      string          `json:"type"`
}

// 打印文本
type PrintText struct {
	Attribute []TextAttribute `json:"attribute"`
	Category  string          `json:"category"`
	ID        string          `json:"id"`
	Text      string          `json:"text"`
	Type      string          `json:"type"`
}

// 文本属性
type TextAttribute struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

// 通用属性
type Attribute struct {
	Name  string      `json:"name,omitempty"`
	Value interface{} `json:"value,omitempty"`
}

// 坐标点
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}
