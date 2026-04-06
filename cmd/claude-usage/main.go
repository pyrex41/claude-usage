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
	projectStr, _ := cmd.Flags().GetString("project")
	isJSON, _ := cmd.Flags().GetBool("json")
	isCompact, _ := cmd.Flags().GetBool("compact")
	isBreakdown, _ := cmd.Flags().GetBool("breakdown")
	isOffline, _ := cmd.Flags().GetBool("offline")
	instances, _ := cmd.Flags().GetBool("instances")
	pathStr, _ := cmd.Root().Flags().GetString("path")

	pricing.LoadPrices(isOffline)

	paths, err := parser.FindFiles(pathStr)
	if err != nil || len(paths) == 0 {
		paths = []string{"testdata/sample.jsonl", "testdata/large.jsonl"}
	}
	p := parser.Parser{Filters: types.Filters{Since: sinceStr, Until: untilStr, Project: projectStr}}
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
