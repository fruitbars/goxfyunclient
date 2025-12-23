package ocr

import (
	"strings"
	"sync"
)

// Language 结构体用于存储语言信息
type Language struct {
	ID           int    // 编号
	Name         string // 语种名称
	Code         string // 语种代码
	NvidiaGPU    string // 英伟达 GPU
	CambriconMLU string // 寒武纪 MLU
	HuaweiAtlas  string // 华为 Atlas
	ASECode      string // ASE的语种代码
}

// LanguageMap 用于存储和查询语言数据
type LanguageMap struct {
	byID           map[int]*Language
	byName         map[string]*Language
	byCode         map[string]*Language
	byNvidiaGPU    map[string][]*Language
	byCambriconMLU map[string][]*Language
	byHuaweiAtlas  map[string][]*Language
	byASECode      map[string]*Language
}

var (
	instance *LanguageMap
	once     sync.Once
)

// GetInstance 返回LanguageMap的单例实例
func GetInstance() *LanguageMap {
	once.Do(func() {
		instance = &LanguageMap{
			byID:           make(map[int]*Language),
			byName:         make(map[string]*Language),
			byCode:         make(map[string]*Language),
			byNvidiaGPU:    make(map[string][]*Language),
			byCambriconMLU: make(map[string][]*Language),
			byHuaweiAtlas:  make(map[string][]*Language),
			byASECode:      make(map[string]*Language),
		}
		instance.initData()
	})
	return instance
}

// initData 初始化语言数据
func (lm *LanguageMap) initData() {
	// ASE代码映射
	aseCodeMap := map[string]string{
		"af":    "荷兰",
		"az":    "阿塞拜疆",
		"bg":    "保加利亚",
		"bn":    "孟加拉",
		"ch_en": "中英",
		"cs":    "捷克",
		"da":    "丹麦",
		"de":    "德语",
		"el":    "希腊",
		"es":    "西班牙语",
		"fa":    "波斯",
		"fi":    "芬兰",
		"fr":    "法语",
		"ha":    "豪撒",
		"he":    "希伯来",
		"hr":    "克罗地亚",
		"hu":    "匈牙利",
		"hy":    "亚美尼亚",
		"id":    "印尼语",
		"it":    "意大利语",
		"ja":    "日语",
		"ka":    "格鲁吉亚",
		"kka":   "哈萨克语",
		"ko":    "韩语",
		"lo":    "老挝",
		"lt":    "立陶宛",
		"lv":    "拉脱维亚",
		"mn":    "内蒙语",
		"ms":    "马来语",
		"nb":    "挪威",
		"pl":    "波兰",
		"ps":    "普什图",
		"pt":    "葡萄牙语",
		"ro":    "罗马尼亚",
		"ru":    "俄语",
		"sk":    "斯洛伐克",
		"sl":    "斯洛文尼亚",
		"sr":    "塞尔维亚",
		"sv":    "瑞典",
		"sw":    "斯瓦西里",
		"ta":    "泰米尔",
		"te":    "泰卢固",
		"tg":    "塔吉克",
		"tk":    "土库曼",
		"tl":    "菲律宾",
		"tr":    "土耳其",
		"uk":    "乌克兰",
		"ur":    "乌尔都",
		"uz":    "乌兹别克",
		"vi":    "越南语",
		"hi":    "印地语",
		"th":    "泰语",
		"ar":    "阿拉伯语",
	}

	// 定义语言数据
	languages := []Language{
		{1, "中英", "ch_en", "ch_en", "cam.ch_en", "atlas.ch_en", "ch_en"},
		{2, "印地语", "hindi", "hindi", "cam.hindi", "atlas.hindi", "hi"},
		{3, "阿拉伯语", "ar", "arabic", "cam.arabic", "atlas.arabic", "ar"},
		{4, "泰语", "thai", "thai", "cam.thai", "atlas.thai", "th"},
		{5, "越南语", "viet", "viet", "cam.viet", "atlas.viet", "vi"},
		{6, "匈牙利语", "hu", "hu", "cam.hu", "atlas.hu", "hu"},
		{7, "法语", "fr", "mix0", "cam.mix0", "atlas.mix0", "fr"},
		{8, "西班牙语", "es", "mix0", "cam.mix0", "atlas.mix0", "es"},
		{9, "德语", "de", "mix0", "cam.mix0", "atlas.mix0", "de"},
		{10, "意大利语", "it", "mix0", "cam.mix0", "atlas.mix0", "it"},
		{11, "葡萄牙语", "pt", "mix0", "cam.mix0", "atlas.mix0", "pt"},
		{12, "马来语", "ms", "mix0", "cam.mix0", "atlas.mix0", "ms"},
		{13, "印尼语", "id", "mix0", "cam.mix0", "atlas.mix0", "id"},
		{14, "日语", "ja", "mix1", "cam.mix1", "atlas.mix1", "ja"},
		{15, "韩语", "ko", "mix1", "cam.mix1", "atlas.mix1", "ko"},
		{16, "俄语", "ru", "mix1", "cam.mix1", "atlas.mix1", "ru"},
		{17, "哈萨克语", "kka", "mix1", "cam.mix1", "atlas.mix1", "kka"},
		{18, "希腊语", "el", "mix3", "cam.mix3", "atlas.mix3", "el"},
		{19, "老挝语", "lo", "mix3", "cam.mix3", "atlas.mix3", "lo"},
		{20, "泰米尔语", "ta", "mix3", "cam.mix3", "atlas.mix3", "ta"},
		{21, "泰卢固语", "te", "mix3", "cam.mix3", "atlas.mix3", "te"},
		{22, "亚美尼亚语", "hy", "mix3", "cam.mix3", "atlas.mix3", "hy"},
		{23, "格鲁吉亚语", "ka", "mix4", "cam.mix4", "atlas.mix4", "ka"},
		{24, "拉脱维亚语", "lv", "mix4", "cam.mix4", "atlas.mix4", "lv"},
		{25, "阿塞拜疆语", "az", "mix4", "cam.mix4", "atlas.mix4", "az"},
		{26, "丹麦语", "da", "mix4", "cam.mix4", "atlas.mix4", "da"},
		{27, "芬兰语", "fi", "mix4", "cam.mix4", "atlas.mix4", "fi"},
		{28, "斯瓦西里语", "sw", "mix5", "cam.mix5", "atlas.mix5", "sw"},
		{29, "罗马尼亚语", "ro", "mix5", "cam.mix5", "atlas.mix5", "ro"},
		{30, "豪撒语", "ha", "mix5", "cam.mix5", "atlas.mix5", "ha"},
		{31, "瑞典语", "sv", "mix5", "cam.mix5", "atlas.mix5", "sv"},
		{32, "土耳其语", "tr", "mix5", "cam.mix5", "atlas.mix5", "tr"},
		{33, "乌兹别克语", "uz", "mix5", "cam.mix5", "atlas.mix5", "uz"},
		{34, "克罗地亚语", "hr", "mix6", "cam.mix6", "atlas.mix6", "hr"},
		{35, "孟加拉语", "bn", "mix6", "cam.mix6", "atlas.mix6", "bn"},
		{36, "波兰语", "pl", "mix6", "cam.mix6", "atlas.mix6", "pl"},
		{37, "捷克语", "cs", "mix6", "cam.mix6", "atlas.mix6", "cs"},
		{38, "菲律宾语", "tl", "mix6", "cam.mix6", "atlas.mix6", "tl"},
		{39, "荷兰语", "af", "mix6", "cam.mix6", "atlas.mix6", "af"},
		{40, "斯洛伐克语", "sk", "mix7", "cam.mix7", "atlas.mix7", "sk"},
		{41, "立陶宛语", "lt", "mix7", "cam.mix7", "atlas.mix7", "lt"},
		{42, "斯洛文尼亚语", "sl", "mix7", "cam.mix7", "atlas.mix7", "sl"},
		{43, "挪威语", "nb", "mix7", "cam.mix7", "atlas.mix7", "nb"},
		{44, "塔吉克语", "tg", "mix7", "cam.mix7", "atlas.mix7", "tg"},
		{45, "土库曼语", "tk", "mix7", "cam.mix7", "atlas.mix7", "tk"},
		{46, "波斯语", "fa", "mix8", "cam.mix8", "atlas.mix8", "fa"},
		{47, "乌尔都语", "ur", "mix8", "cam.mix8", "atlas.mix8", "ur"},
		{48, "希伯来语", "he", "mix8", "cam.mix8", "atlas.mix8", "he"},
		{49, "普什图语", "ps", "mix8", "cam.mix8", "atlas.mix8", "ps"},
		{50, "保加利亚语", "bg", "mix9", "cam.mix9", "atlas.mix9", "bg"},
		{51, "乌克兰语", "uk", "mix9", "cam.mix9", "atlas.mix9", "uk"},
		{52, "塞尔维亚语", "sr", "mix9", "cam.mix9", "atlas.mix9", "sr"},
		{53, "蒙语", "mn", "mn", "cam.mn", "atlas.mn", "mn"},
		{54, "维语", "uyg", "wei", "cam.uyg", "atlas.uyg", "uyg"},
	}

	// 添加所有语言到映射中
	for i := range languages {
		lang := &languages[i]

		// 检查ASE代码是否存在，如果不存在，尝试通过名称查找
		if lang.ASECode == "" {
			for code, name := range aseCodeMap {
				if strings.Contains(lang.Name, name) || strings.Contains(name, strings.TrimSuffix(lang.Name, "语")) {
					lang.ASECode = code
					break
				}
			}
		}

		lm.byID[lang.ID] = lang
		lm.byName[lang.Name] = lang
		lm.byCode[lang.Code] = lang

		// 添加到GPU映射
		lm.byNvidiaGPU[lang.NvidiaGPU] = append(lm.byNvidiaGPU[lang.NvidiaGPU], lang)
		lm.byCambriconMLU[lang.CambriconMLU] = append(lm.byCambriconMLU[lang.CambriconMLU], lang)
		lm.byHuaweiAtlas[lang.HuaweiAtlas] = append(lm.byHuaweiAtlas[lang.HuaweiAtlas], lang)

		// 如果有ASE代码，添加到ASE映射
		if lang.ASECode != "" {
			lm.byASECode[lang.ASECode] = lang
		}
	}
}

// FindByID 通过ID查找语言
func (lm *LanguageMap) FindByID(id int) *Language {
	return lm.byID[id]
}

// FindByName 通过名称查找语言
func (lm *LanguageMap) FindByName(name string) *Language {
	return lm.byName[name]
}

// FindByCode 通过代码查找语言
func (lm *LanguageMap) FindByCode(code string) *Language {
	return lm.byCode[code]
}

// FindByNvidiaGPU 通过英伟达GPU代码查找语言
func (lm *LanguageMap) FindByNvidiaGPU(code string) []*Language {
	if langs, ok := lm.byNvidiaGPU[code]; ok {
		return langs
	}
	return nil
}

// FindByCambriconMLU 通过寒武纪MLU代码查找语言
func (lm *LanguageMap) FindByCambriconMLU(code string) []*Language {

	if langs, ok := lm.byCambriconMLU[code]; ok {
		return langs
	}
	return nil
}

// FindByHuaweiAtlas 通过华为Atlas代码查找语言
func (lm *LanguageMap) FindByHuaweiAtlas(code string) []*Language {
	if langs, ok := lm.byHuaweiAtlas[code]; ok {
		return langs
	}
	return nil
}

// FindByASECode 通过ASE代码查找语言
func (lm *LanguageMap) FindByASECode(code string) *Language {
	return lm.byASECode[code]
}

// CodeToNvidia 将语种代码转换为英伟达GPU代码
func CodeToNvidia(code string) string {
	lm := GetInstance()
	lang := lm.FindByCode(code)
	if lang == nil {
		return ""
	}
	return lang.NvidiaGPU
}

// CodeToCambricon 将语种代码转换为寒武纪MLU代码
func CodeToCambricon(code string) string {
	lm := GetInstance()
	lang := lm.FindByCode(code)
	if lang == nil {
		return ""
	}
	return lang.CambriconMLU
}

// CodeToAtlas 将语种代码转换为华为Atlas代码
func CodeToAtlas(code string) string {
	lm := GetInstance()
	lang := lm.FindByCode(code)
	if lang == nil {
		return ""
	}
	return lang.HuaweiAtlas
}

// ASEToCode 将ASE代码转换为语种代码
func ASEToCode(aseCode string) string {
	lm := GetInstance()
	lang := lm.FindByASECode(aseCode)
	if lang == nil {
		return ""
	}
	return lang.Code
}

// ASEToNvidia 将ASE代码转换为英伟达GPU代码
func ASEToNvidia(aseCode string) string {
	lm := GetInstance()
	lang := lm.FindByASECode(aseCode)
	if lang == nil {
		return ""
	}
	return lang.NvidiaGPU
}

// ASEToCambricon 将ASE代码转换为寒武纪MLU代码
func ASEToCambricon(aseCode string) string {
	lm := GetInstance()
	lang := lm.FindByASECode(aseCode)
	if lang == nil {
		return ""
	}
	return lang.CambriconMLU
}

// ASEToAtlas 将ASE代码转换为华为Atlas代码
func ASEToAtlas(aseCode string) string {
	lm := GetInstance()
	lang := lm.FindByASECode(aseCode)
	if lang == nil {
		return ""
	}
	return lang.HuaweiAtlas
}

// NvidiaToLanguages 将英伟达GPU代码转换为支持的语言列表
func NvidiaToLanguages(nvidiaCode string) []*Language {
	lm := GetInstance()
	return lm.FindByNvidiaGPU(nvidiaCode)
}

// CambriconToLanguages 将寒武纪MLU代码转换为支持的语言列表
func CambriconToLanguages(cambriconCode string) []*Language {
	lm := GetInstance()
	return lm.FindByCambriconMLU(cambriconCode)
}

// AtlasToLanguages 将华为Atlas代码转换为支持的语言列表
func AtlasToLanguages(atlasCode string) []*Language {
	lm := GetInstance()
	return lm.FindByHuaweiAtlas(atlasCode)
}

// GetLanguageByID 通过ID查找语言（便捷函数）
func GetLanguageByID(id int) *Language {
	return GetInstance().FindByID(id)
}

// GetLanguageByName 通过名称查找语言（便捷函数）
func GetLanguageByName(name string) *Language {
	return GetInstance().FindByName(name)
}

// GetLanguageByCode 通过代码查找语言（便捷函数）
func GetLanguageByCode(code string) *Language {
	return GetInstance().FindByCode(code)
}

// GetLanguageByASECode 通过ASE代码查找语言（便捷函数）
func GetLanguageByASECode(aseCode string) *Language {
	return GetInstance().FindByASECode(aseCode)
}

// GPUToLanguages 根据任何GPU代码(NVIDIA、寒武纪或华为Atlas)查找支持的语言列表
func GPUCategoryToLanguages(code string) []*Language {
	lm := GetInstance()

	// 首先检查是否符合各平台代码格式，然后在相应的映射中查找
	if strings.HasPrefix(code, "cam.") {
		// 寒武纪MLU代码格式
		return lm.FindByCambriconMLU(code)
	} else if strings.HasPrefix(code, "atlas.") {
		// 华为Atlas代码格式
		return lm.FindByHuaweiAtlas(code)
	} else {
		// 默认尝试作为NVIDIA GPU代码查找
		// 注意：这里无需特殊前缀判断，因为NVIDIA代码没有统一前缀
		return lm.FindByNvidiaGPU(code)
	}
}
