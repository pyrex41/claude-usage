package aggregator

import (
	"sort"
	"sync"
	"time"

	"github.com/pyrex41/claude-usage/internal/pricing"
	"github.com/pyrex41/claude-usage/internal/types"
)

type ReportType string

const (
	Daily ReportType = "daily"
)

type Aggregator struct {
	Type       ReportType
	Timezone   *time.Location
	Buckets    map[string]*types.Bucket
	Total      *types.Bucket
	Projects   map[string]map[string]*types.Bucket
	mu         sync.Mutex
	DoProjects bool
}

func NewAgg(rt ReportType, tz *time.Location) *Aggregator {
	return &Aggregator{
		Type:     rt,
		Timezone: tz,
		Buckets:  make(map[string]*types.Bucket),
		Total: &types.Bucket{
			ModelBreakdown: make(map[string]*types.ModelUsage),
		},
		Projects: make(map[string]map[string]*types.Bucket),
	}
}

func (a *Aggregator) EnableProjects() {
	a.DoProjects = true
}

func (a *Aggregator) Aggregate(ch <-chan types.Event) {
	for event := range ch {
		a.addEvent(event)
	}
}

func (a *Aggregator) addEvent(e types.Event) {
	t := e.Time()
	if t.IsZero() || t.Year() < 2020 {
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	key := a.bucketKey(t.In(a.Timezone), e)
	b := a.getOrCreateBucket(key)

	in, out, cc, cr := e.Tokens()
	cost := e.CostUSD
	if cost == 0 {
		cost = pricing.CalcCost(e.Message.Model, in, out, cc, cr, e.Message.Usage.Speed)
	}

	model := e.Message.Model

	mu := b.ModelBreakdown[model]
	if mu == nil {
		mu = &types.ModelUsage{Model: model}
		b.ModelBreakdown[model] = mu
	}
	mu.InputTokens += in
	mu.OutputTokens += out
	mu.CacheCreate += cc
	mu.CacheRead += cr
	mu.Cost += cost

	b.Count++

	totalMu := a.Total.ModelBreakdown[model]
	if totalMu == nil {
		totalMu = &types.ModelUsage{Model: model}
		a.Total.ModelBreakdown[model] = totalMu
	}
	totalMu.InputTokens += in
	totalMu.OutputTokens += out
	totalMu.CacheCreate += cc
	totalMu.CacheRead += cr
	totalMu.Cost += cost
	a.Total.Count++

	if a.DoProjects {
		projBuckets := a.Projects[e.Project]
		if projBuckets == nil {
			projBuckets = make(map[string]*types.Bucket)
			a.Projects[e.Project] = projBuckets
		}
		dateKey := a.bucketKey(t.In(a.Timezone), e)
		pb := projBuckets[dateKey]
		if pb == nil {
			pb = &types.Bucket{Date: dateKey, ModelBreakdown: make(map[string]*types.ModelUsage)}
			projBuckets[dateKey] = pb
		}
		pmu := pb.ModelBreakdown[model]
		if pmu == nil {
			pmu = &types.ModelUsage{Model: model}
			pb.ModelBreakdown[model] = pmu
		}
		pmu.InputTokens += in
		pmu.OutputTokens += out
		pmu.CacheCreate += cc
		pmu.CacheRead += cr
		pmu.Cost += cost
		pb.Count++
	}
}

func (a *Aggregator) bucketKey(t time.Time, e types.Event) string {
	switch a.Type {
	case Daily:
		return t.Format("2006-01-02")
	default:
		return t.Format("2006-01-02")
	}
}

func (a *Aggregator) getOrCreateBucket(key string) *types.Bucket {
	b, ok := a.Buckets[key]
	if !ok {
		b = &types.Bucket{Date: key, ModelBreakdown: make(map[string]*types.ModelUsage)}
		a.Buckets[key] = b
	}
	return b
}

func (a *Aggregator) calculateTotals() {
	a.Total.InputTokens = 0
	a.Total.OutputTokens = 0
	a.Total.CacheCreate = 0
	a.Total.CacheRead = 0
	a.Total.TotalCost = 0
	for _, mu := range a.Total.ModelBreakdown {
		a.Total.InputTokens += mu.InputTokens
		a.Total.OutputTokens += mu.OutputTokens
		a.Total.CacheCreate += mu.CacheCreate
		a.Total.CacheRead += mu.CacheRead
		a.Total.TotalCost += mu.Cost
	}

	for _, projBuckets := range a.Projects {
		for _, pb := range projBuckets {
			pb.InputTokens = 0
			pb.OutputTokens = 0
			pb.CacheCreate = 0
			pb.CacheRead = 0
			pb.TotalCost = 0
			for _, mu := range pb.ModelBreakdown {
				pb.InputTokens += mu.InputTokens
				pb.OutputTokens += mu.OutputTokens
				pb.CacheCreate += mu.CacheCreate
				pb.CacheRead += mu.CacheRead
				pb.TotalCost += mu.Cost
			}
		}
	}

	for _, b := range a.Buckets {
		b.InputTokens = 0
		b.OutputTokens = 0
		b.CacheCreate = 0
		b.CacheRead = 0
		b.TotalCost = 0
		for _, mu := range b.ModelBreakdown {
			b.InputTokens += mu.InputTokens
			b.OutputTokens += mu.OutputTokens
			b.CacheCreate += mu.CacheCreate
			b.CacheRead += mu.CacheRead
			b.TotalCost += mu.Cost
		}
	}
}

func (a *Aggregator) Result() *types.AggResult {
	a.calculateTotals()

	res := &types.AggResult{ReportType: string(a.Type)}
	keys := make([]string, 0, len(a.Buckets))
	for k := range a.Buckets {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		b := a.Buckets[k]
		modelKeys := make([]string, 0, len(b.ModelBreakdown))
		for m := range b.ModelBreakdown {
			modelKeys = append(modelKeys, m)
		}
		sort.Strings(modelKeys)
		b.SortedModels = modelKeys
		for _, mu := range b.ModelBreakdown {
			if b.TotalCost > 0 {
				mu.Percentage = mu.Cost / b.TotalCost * 100
			}
		}
		res.Buckets = append(res.Buckets, b)
	}
	res.Total = a.Total

	if a.DoProjects {
		res.Projects = a.Projects
	}
	return res
}
