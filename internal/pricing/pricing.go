// Clean pricing.go
package pricing

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
)

var (
	pricesMu sync.RWMutex
	prices   = make(map[string]*LiteLLMPrice)
)

type LiteLLMPrice struct {
	Model         string  `json:"model_name"`
	Input         float64 `json:"input"`
	Output        float64 `json:"output"`
	CacheCreate   float64 `json:"cache_creation_input"`
	CacheRead     float64 `json:"cache_read_input"`
	ContextWindow int     `json:"context_window"`
	SupportsFast  bool    `json:"supports_speed"`
}

const url = "https://raw.githubusercontent.com/BerriAI/litellm/main/model_prices_and_context_window.json"

func LoadPrices(offline bool) error {
	if offline {
		return nil
	}
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var models []LiteLLMPrice
	if err := json.Unmarshal(body, &models); err != nil {
		return err
	}
	pricesMu.Lock()
	defer pricesMu.Unlock()
	for _, m := range models {
		if strings.HasPrefix(m.Model, "claude/") || strings.HasPrefix(m.Model, "anthropic/") {
			prices[m.Model] = &m
		}
	}
	return nil
}

func CalcCost(model string, input, output, cc, cr uint64, speed string) float64 {
	// Try exact match first
	p := prices[model]
	if p == nil {
		// Try anthropic/ prefix
		p = prices["anthropic/"+model]
	}
	if p == nil {
		// Hardcoded fallback for common models
		return calcHardcoded(model, input, output, cc, cr, speed)
	}
	rateIn, rateOut := p.Input/1e6, p.Output/1e6
	rateCC, rateCR := p.CacheCreate/1e6, p.CacheRead/1e6
	if speed == "fast" {
		rateIn *= 6
		rateOut *= 6
	}
	return float64(input)*rateIn + float64(output)*rateOut + float64(cc)*rateCC + float64(cr)*rateCR
}

func calcHardcoded(model string, input, output, cc, cr uint64, speed string) float64 {
	var rateIn, rateOut, rateCC, rateCR float64
	switch {
	case strings.Contains(model, "sonnet"):
		rateIn, rateOut, rateCC, rateCR = 3.0/1e6, 15.0/1e6, 3.75/1e6, 0.30/1e6
	case strings.Contains(model, "opus"):
		rateIn, rateOut, rateCC, rateCR = 15.0/1e6, 75.0/1e6, 18.75/1e6, 1.50/1e6
	case strings.Contains(model, "haiku"):
		rateIn, rateOut, rateCC, rateCR = 0.25/1e6, 1.25/1e6, 0.30/1e6, 0.03/1e6
	default:
		rateIn, rateOut = 3.0/1e6, 15.0/1e6
	}
	if speed == "fast" {
		rateIn *= 6
		rateOut *= 6
	}
	return float64(input)*rateIn + float64(output)*rateOut + float64(cc)*rateCC + float64(cr)*rateCR
}
