package models

// 顶层
type Root struct {
	Lattice  []LatticeItem  `json:"lattice"`
	Lattice2 []Lattice2Item `json:"lattice2"`
}

// lattice: json_1best 是一个被转义的 JSON 字符串
type LatticeItem struct {
	JSON1Best string `json:"json_1best"`
}

// lattice2: json_1best 是真正的对象
type Lattice2Item struct {
	LID       string       `json:"lid"`
	End       string       `json:"end"`
	Begin     string       `json:"begin"`
	JSON1Best JSON1BestObj `json:"json_1best"`
	Spk       string       `json:"spk"`
}

// 通用：json_1best 对象
type JSON1BestObj struct {
	St St `json:"st"`
	// 这些键在不同片段里可选出现
	Pt string `json:"pt,omitempty"`
	Bg string `json:"bg,omitempty"`
	Si string `json:"si,omitempty"`
	Rl string `json:"rl,omitempty"`
	Ed string `json:"ed,omitempty"`
}

// st 节点
type St struct {
	Sc string `json:"sc"`
	Pa string `json:"pa"`
	Rt []Rt   `json:"rt"`
	// 有些版本把这些也放在 st 里（你的示例里两种都出现了），因此做可选
	Bg string `json:"bg,omitempty"`
	Rl string `json:"rl,omitempty"`
	Ed string `json:"ed,omitempty"`
}

// rt 项（lattice2 里多出 nb/nc）
type Rt struct {
	Nb string `json:"nb,omitempty"`
	Nc string `json:"nc,omitempty"`
	Ws []Ws   `json:"ws"`
}

// 分词窗口
type Ws struct {
	Cw []Cw `json:"cw"`
	Wb int  `json:"wb"`
	We int  `json:"we"`
}

// 候选词
type Cw struct {
	W  string `json:"w"`
	Wp string `json:"wp"`
	Wc string `json:"wc"`
}
