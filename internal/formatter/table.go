package formatter

import (
	"encoding/json"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"

	"claude-usage/internal/types"
)

func FormatJSON(w io.Writer, res *types.AggResult) error {
	return json.NewEncoder(w).Encode(res)
}

func formatNum(n uint64) string {
	s := strconv.FormatUint(n, 10)
	if n < 1000 {
		return s
	}
	result := ""
	for i, r := range s {
		if (len(s)-i)%3 == 0 && i > 0 {
			result += ","
		}
		result += string(r)
	}
	return result
}

func formatCost(f float64) string {
	if f < 0.01 {
		return "$0.00"
	}
	return "$" + strconv.FormatFloat(f, 'f', 2, 64)
}

func FormatTable(w io.Writer, res *types.AggResult, compact, breakdown bool) {
	t := table.NewWriter()
	t.SetOutputMirror(w)
	t.SetStyle(table.StyleLight)

	if len(res.Projects) > 0 {
		// Project view - simpler and cleaner
		t.SetTitle("Claude Code Usage - By Project")
		t.AppendHeader(table.Row{"Project", "Date", "Model", "Input", "Output", "Cache", "Read", "Total", "Cost"})

		projKeys := make([]string, 0, len(res.Projects))
		for k := range res.Projects {
			projKeys = append(projKeys, k)
		}
		sort.Strings(projKeys)

		for _, projName := range projKeys {
			projBuckets := res.Projects[projName]
			if projBuckets == nil {
				continue
			}

			bucketKeys := make([]string, 0, len(projBuckets))
			for k := range projBuckets {
				bucketKeys = append(bucketKeys, k)
			}
			sort.Strings(bucketKeys)

			for _, dateStr := range bucketKeys {
				b := projBuckets[dateStr]
				if b == nil {
					continue
				}

				models := make([]string, 0, len(b.ModelBreakdown))
				for m := range b.ModelBreakdown {
					models = append(models, m)
				}
				sort.Strings(models)

				first := true
				for _, model := range models {
					mu := b.ModelBreakdown[model]
					if mu == nil {
						continue
					}
					if mu.InputTokens == 0 && mu.OutputTokens == 0 && mu.CacheCreate == 0 && mu.CacheRead == 0 {
						continue
					}

					totalTokens := mu.InputTokens + mu.OutputTokens + mu.CacheCreate + mu.CacheRead
					cost := formatCost(mu.Cost)

					row := table.Row{"", dateStr, cleanModelName(model), formatNum(mu.InputTokens), formatNum(mu.OutputTokens), formatNum(mu.CacheCreate), formatNum(mu.CacheRead), formatNum(totalTokens), ""}
					if first {
						row[0] = projName
						row[8] = cost
						first = false
					}
					t.AppendRow(row)
				}

				// Add daily total row only if there's data
				if b.InputTokens > 0 || b.OutputTokens > 0 || b.CacheCreate > 0 || b.CacheRead > 0 {
					totalTokens := b.InputTokens + b.OutputTokens + b.CacheCreate + b.CacheRead
					t.AppendRow(table.Row{"", "", "→ TOTAL", formatNum(b.InputTokens), formatNum(b.OutputTokens), formatNum(b.CacheCreate), formatNum(b.CacheRead), formatNum(totalTokens), formatCost(b.TotalCost)})
				}
			}
		}
	} else {
		// Regular daily view
		t.SetTitle("Claude Code Usage - Daily")
		if compact {
			t.AppendHeader(table.Row{"Date", "Model", "Input", "Output", "Cost"})
		} else {
			t.AppendHeader(table.Row{"Date", "Model", "Input", "Output", "Cache Create", "Cache Read", "Total", "Cost"})
		}

		for _, b := range res.Buckets {
			models := b.SortedModels
			if len(models) == 0 {
				models = []string{"-"}
			}

			for _, model := range models {
				mu := b.ModelBreakdown[model]
				if mu == nil {
					continue
				}
				if mu.InputTokens == 0 && mu.OutputTokens == 0 && mu.CacheCreate == 0 && mu.CacheRead == 0 {
					continue
				}

				totalTokens := mu.InputTokens + mu.OutputTokens + mu.CacheCreate + mu.CacheRead

				if compact {
					t.AppendRow(table.Row{
						b.Date,
						cleanModelName(model),
						formatNum(mu.InputTokens),
						formatNum(mu.OutputTokens),
						formatCost(b.TotalCost),
					})
				} else {
					t.AppendRow(table.Row{
						b.Date,
						cleanModelName(model),
						formatNum(mu.InputTokens),
						formatNum(mu.OutputTokens),
						formatNum(mu.CacheCreate),
						formatNum(mu.CacheRead),
						formatNum(totalTokens),
						formatCost(b.TotalCost),
					})
				}
			}
		}
	}

	t.Render()
}

func cleanModelName(m string) string {
	m = strings.TrimPrefix(m, "claude-")
	if strings.Contains(m, "-2025") {
		m = strings.Split(m, "-2025")[0]
	}
	if len(m) > 25 {
		m = m[:22] + "..."
	}
	return m
}
