package ist

import (
	"log"
	"net/url"
	"strconv"
	"strings"
)

// UploadOptions 存储上传参数
type UploadOptions struct {
	Language      string
	CallbackURL   string
	HotWord       string
	Candidate     int
	RoleType      int
	RoleNum       int
	Pd            string
	AudioMode     string
	AudioURL      string
	StandardWav   int
	LanguageType  int
	TrackMode     int
	TransLanguage string
	TransMode     int
	EngSegMax     int
	EngSegMin     int
	EngSegWeight  float64
	EngSmoothProc bool
	EngColloqProc bool
	EngVadMdn     int
	EngVadMargin  int
	EngRLang      int
	Predict       bool
	Duration      string
}

// UploadOption 定义选项函数
type UploadOption func(*UploadOptions)

func WithDuration(duration string) UploadOption {
	return func(o *UploadOptions) {
		o.Duration = duration
	}
}

// WithLanguage 设置语种
func WithLanguage(language string) UploadOption {
	return func(o *UploadOptions) {
		o.Language = language
	}
}

// WithCallbackURL 设置回调地址
func WithCallbackURL(url string) UploadOption {
	return func(o *UploadOptions) {
		o.CallbackURL = url
	}
}

// WithHotWord 设置热词
func WithHotWord(hotWord string) UploadOption {
	return func(o *UploadOptions) {
		o.HotWord = hotWord
	}
}

// WithCandidate 设置多候选开关
func WithCandidate(enabled bool) UploadOption {
	return func(o *UploadOptions) {
		o.Candidate = boolToInt(enabled)
	}
}

// WithRoleType 设置角色分离
func WithRoleType(enabled bool) UploadOption {
	return func(o *UploadOptions) {
		o.RoleType = boolToInt(enabled)
	}
}

// WithRoleNum 设置说话人数
func WithRoleNum(num int) UploadOption {
	return func(o *UploadOptions) {
		if num >= 0 && num <= 10 {
			o.RoleNum = num
		}
	}
}

// WithPd 设置领域个性化参数
func WithPd(pd string) UploadOption {
	return func(o *UploadOptions) {
		o.Pd = pd
	}
}

// WithAudioMode 设置音频上传方式
func WithAudioMode(mode string) UploadOption {
	return func(o *UploadOptions) {
		o.AudioMode = mode
	}
}

// WithAudioURL 设置音频外链地址
func WithAudioURL(url string) UploadOption {
	return func(o *UploadOptions) {
		o.AudioURL = url
	}
}

// WithStandardWav 设置标准 PCM/WAV
func WithStandardWav(enabled bool) UploadOption {
	return func(o *UploadOptions) {
		o.StandardWav = boolToInt(enabled)
	}
}

// 其他参数的选项函数
func WithLanguageType(languageType int) UploadOption {
	return func(o *UploadOptions) {
		o.LanguageType = languageType
	}
}

func WithTrackMode(trackMode int) UploadOption {
	return func(o *UploadOptions) {
		o.TrackMode = trackMode
	}
}

func WithTransLanguage(transLanguage string) UploadOption {
	return func(o *UploadOptions) {
		o.TransLanguage = transLanguage
	}
}

func WithTransMode(transMode int) UploadOption {
	return func(o *UploadOptions) {
		o.TransMode = transMode
	}
}

func WithEngSegMax(max int) UploadOption {
	return func(o *UploadOptions) {
		o.EngSegMax = max
	}
}

func WithEngSegMin(min int) UploadOption {
	return func(o *UploadOptions) {
		o.EngSegMin = min
	}
}

func WithEngSegWeight(weight float64) UploadOption {
	return func(o *UploadOptions) {
		o.EngSegWeight = weight
	}
}

func WithEngSmoothProc(enabled bool) UploadOption {
	return func(o *UploadOptions) {
		o.EngSmoothProc = enabled
	}
}

func WithEngColloqProc(enabled bool) UploadOption {
	return func(o *UploadOptions) {
		o.EngColloqProc = enabled
	}
}

func WithEngVadMdn(mode int) UploadOption {
	return func(o *UploadOptions) {
		o.EngVadMdn = mode
	}
}

func WithEngVadMargin(margin int) UploadOption {
	return func(o *UploadOptions) {
		o.EngVadMargin = margin
	}
}

func WithEngRLang(rlang int) UploadOption {
	return func(o *UploadOptions) {
		o.EngRLang = rlang
	}
}

// WithPredict 设置是否启用质检功能
func WithPredict(enabled bool) UploadOption {
	return func(o *UploadOptions) {
		o.Predict = enabled
	}
}

// boolToInt 辅助函数
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// ToURLValues 将 UploadOptions 转换为 url.Values
func (o UploadOptions) ToURLValues() url.Values {
	params := url.Values{}
	if o.Language != "" {
		params.Set("language", o.Language)
	}
	if o.CallbackURL != "" {
		params.Set("callbackUrl", o.CallbackURL)
	}
	if o.HotWord != "" {
		params.Set("hotWord", o.HotWord)
	}
	if o.Candidate != 0 {
		params.Set("candidate", strconv.Itoa(o.Candidate))
	}
	if o.RoleType != 0 {
		params.Set("roleType", strconv.Itoa(o.RoleType))
	}
	if o.RoleNum != 0 {
		params.Set("roleNum", strconv.Itoa(o.RoleNum))
	}
	if o.Pd != "" {
		params.Set("pd", o.Pd)
	}
	if o.AudioMode != "" {
		params.Set("audioMode", o.AudioMode)
	}
	if o.AudioURL != "" {
		params.Set("audioUrl", o.AudioURL)
	}
	if o.StandardWav != 0 {
		params.Set("standardWav", strconv.Itoa(o.StandardWav))
	}
	if o.LanguageType != 0 {
		params.Set("languageType", strconv.Itoa(o.LanguageType))
	}
	if o.TrackMode != 0 {
		params.Set("trackMode", strconv.Itoa(o.TrackMode))
	}
	if o.TransLanguage != "" {
		params.Set("transLanguage", o.TransLanguage)
	}
	if o.TransMode != 0 {
		params.Set("transMode", strconv.Itoa(o.TransMode))
	}
	if o.EngSegMax != 0 {
		params.Set("eng_seg_max", strconv.Itoa(o.EngSegMax))
	}
	if o.EngSegMin != 0 {
		params.Set("eng_seg_min", strconv.Itoa(o.EngSegMin))
	}
	if o.EngSegWeight != 0 {
		params.Set("eng_seg_weight", strconv.FormatFloat(o.EngSegWeight, 'f', -1, 64))
	}
	if o.EngSmoothProc {
		params.Set("eng_smoothproc", "true")
	}
	if o.EngColloqProc {
		params.Set("eng_colloqproc", "true")
	}
	if o.EngVadMdn != 0 {
		params.Set("eng_vad_mdn", strconv.Itoa(o.EngVadMdn))
	}
	if o.EngVadMargin != 0 {
		params.Set("eng_vad_margin", strconv.Itoa(o.EngVadMargin))
	}
	if o.EngRLang != 0 {
		params.Set("eng_rlang", strconv.Itoa(o.EngRLang))
	}
	return params
}

// GetResultType 根据选项生成 resultType
func (o UploadOptions) GetResultType() string {
	var types []string
	if o.TransLanguage != "" {
		if o.Predict {
			log.Println("Warning: predict is ignored when TransLanguage is set")
		}
		return "translate"
	}
	types = append(types, "transfer")
	if o.Predict {
		types = append(types, "predict")
	}
	return strings.Join(types, ",")
}
