package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/pyrex41/claude-usage/internal/aggregator"
	"github.com/pyrex41/claude-usage/internal/formatter"
	"github.com/pyrex41/claude-usage/internal/parser"
	"github.com/pyrex41/claude-usage/internal/pricing"
	"github.com/pyrex41/claude-usage/internal/types"
)

var rootCmd = &cobra.Command{
	Use:   "claude-usage",
	Short: "Fast Claude usage analyzer",
	Long:  "A fast CLI tool to analyze Claude Code usage and costs from JSONL logs.",
}

var dailyCmd = &cobra.Command{
	Use:   "daily",
	Short: "Daily report (default)",
	Run: func(cmd *cobra.Command, args []string) {
		runReport(aggregator.Daily, cmd)
	},
}

func init() {
	dailyCmd.Flags().StringP("since", "s", "", "Start date YYYYMMDD")
	dailyCmd.Flags().StringP("until", "u", "", "End date YYYYMMDD")
	dailyCmd.Flags().StringP("last", "l", "", "Time period: day, week, month, year")
	dailyCmd.Flags().StringP("project", "p", "", "Filter project")
	dailyCmd.Flags().Bool("instances", false, "Group by instance")
	dailyCmd.Flags().BoolP("json", "j", false, "JSON output")
	dailyCmd.Flags().Bool("compact", false, "Compact table")
	dailyCmd.Flags().BoolP("breakdown", "b", false, "Model breakdown")
	dailyCmd.Flags().BoolP("offline", "O", false, "Offline pricing")
	rootCmd.AddCommand(dailyCmd)
	rootCmd.PersistentFlags().StringP("path", "", os.ExpandEnv("$HOME/.claude/projects"), "Data path")
}

func runReport(rt aggregator.ReportType, cmd *cobra.Command) {
	sinceStr, _ := cmd.Flags().GetString("since")
	untilStr, _ := cmd.Flags().GetString("until")
	lastStr, _ := cmd.Flags().GetString("last")
	projectStr, _ := cmd.Flags().GetString("project")

	// --last overrides --since
	if lastStr != "" {
		now := time.Now()
		var t time.Time
		switch lastStr {
		case "day", "d":
			t = now
		case "week", "w":
			t = now.AddDate(0, 0, -6)
		case "month", "m":
			t = now.AddDate(0, -1, 0)
		case "year", "y":
			t = now.AddDate(-1, 0, 0)
		default:
			fmt.Fprintf(os.Stderr, "Unknown --last value %q (use day, week, month, year)\n", lastStr)
			os.Exit(1)
		}
		sinceStr = t.Format("20060102")
	}
	isJSON, _ := cmd.Flags().GetBool("json")
	isCompact, _ := cmd.Flags().GetBool("compact")
	isBreakdown, _ := cmd.Flags().GetBool("breakdown")
	isOffline, _ := cmd.Flags().GetBool("offline")
	instances, _ := cmd.Flags().GetBool("instances")
	pathStr, _ := cmd.Root().Flags().GetString("path")

	// Start pricing fetch in background (non-blocking)
	pricing.LoadPricesAsync(isOffline)

	// Use since date for file mtime filtering
	var sinceTime *time.Time
	if sinceStr != "" {
		t, err := time.Parse("20060102", sinceStr)
		if err == nil {
			sinceTime = &t
		}
	}

	paths, err := parser.FindFiles(pathStr, sinceTime)
	if err != nil || len(paths) == 0 {
		paths = []string{"testdata/sample.jsonl", "testdata/large.jsonl"}
	}
	p := parser.Parser{Filters: types.Filters{Since: sinceStr, Until: untilStr, Project: projectStr}}
	p.Init()
	ch := make(chan types.Event, 10000)
	var wg sync.WaitGroup
	p.ParseFiles(paths, ch, &wg)
	go func() {
		wg.Wait()
		close(ch)
	}()

	loc, _ := time.LoadLocation("Local")
	agg := aggregator.NewAgg(rt, loc)
	if instances {
		agg.DoProjects = true
	}
	agg.Aggregate(ch)

	// Wait for prices before formatting (costs may use fetched rates)
	pricing.WaitForPrices()

	res := agg.Result()
	if isJSON {
		formatter.FormatJSON(os.Stdout, res)
	} else {
		formatter.FormatTable(os.Stdout, res, isCompact, isBreakdown)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
