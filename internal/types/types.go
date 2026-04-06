package types

import "time"

type Event struct {
	CWD       string `json:"cwd,omitempty"`
	SessionID string `json:"sessionId,omitempty"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version,omitempty"`
	Message   struct {
		Usage struct {
			InputTokens              uint64 `json:"input_tokens"`
			OutputTokens             uint64 `json:"output_tokens"`
			CacheCreationInputTokens uint64 `json:"cache_creation_input_tokens,omitempty"`
			CacheReadInputTokens     uint64 `json:"cache_read_input_tokens,omitempty"`
			Speed                    string `json:"speed,omitempty"`
		} `json:"usage"`
		Model string `json:"model,omitempty"`
		ID    string `json:"id,omitempty"`
	} `json:"message"`
	CostUSD   float64 `json:"costUSD,omitempty"`
	RequestID  string    `json:"requestId,omitempty"`
	Project    string    `json:"project,omitempty"`
	Instance   string    `json:"instance,omitempty"`
	ParsedTime time.Time `json:"-"`
}

func (e *Event) Time() time.Time {
	if e.Timestamp == "" {
		return time.Time{}
	}
	if t, err := time.Parse(time.RFC3339, e.Timestamp); err == nil {
		return t
	}
	if t, err := time.Parse("2006-01-02T15:04:05Z", e.Timestamp); err == nil {
		return t
	}
	if t, err := time.Parse("2006-01-02", e.Timestamp); err == nil {
		return t
	}
	return time.Time{}
}

func (e *Event) Tokens() (input, output, cacheCreate, cacheRead uint64) {
	return e.Message.Usage.InputTokens, e.Message.Usage.OutputTokens,
		e.Message.Usage.CacheCreationInputTokens, e.Message.Usage.CacheReadInputTokens
}

type ModelUsage struct {
	Model        string  `json:"modelName"`
	InputTokens  uint64  `json:"inputTokens"`
	OutputTokens uint64  `json:"outputTokens"`
	CacheCreate  uint64  `json:"cacheCreationTokens"`
	CacheRead    uint64  `json:"cacheReadTokens"`
	Cost         float64 `json:"cost"`
	Percentage   float64 `json:"percentage"`
}

type Bucket struct {
	Date           string                 `json:"date"`
	InputTokens    uint64                 `json:"inputTokens"`
	OutputTokens   uint64                 `json:"outputTokens"`
	CacheCreate    uint64                 `json:"cacheCreationTokens"`
	CacheRead      uint64                 `json:"cacheReadTokens"`
	TotalCost      float64                `json:"totalCost"`
	ModelBreakdown map[string]*ModelUsage `json:"modelBreakdowns,omitempty"`
	Count          int                    `json:"count"`
	SortedModels   []string
}

type AggResult struct {
	ReportType string                        `json:"reportType"`
	Buckets    []*Bucket                     `json:"buckets"`
	Total      *Bucket                       `json:"total"`
	ModelsUsed []string                      `json:"modelsUsed"`
	Projects   map[string]map[string]*Bucket `json:"projects,omitempty"`
}

type Filters struct {
	Since   string
	Until   string
	Project string
}
