You are an expert Go performance engineer. 

**Guiding Principles (must follow):**
- **Simple**: Prefer standard library and minimal dependencies. Avoid complexity.
- **Flexible**: Design for easy extension and adaptation to different data volumes and use cases.
- **Just Works**: Prioritize robustness, great defaults, intuitive CLI, and zero-config where possible. Make it reliable on real-world messy data.

Your task is to completely rewrite the ccusage CLI tool (a command-line utility for analyzing Claude Code / Codex CLI usage from local JSONL files) to make it dramatically faster while maintaining 100% feature parity.

Original repository: https://github.com/ryoppippi/ccusage

**Core Requirements (must preserve exactly):**
- Parses one or more JSONL files containing Claude usage events (each line is a JSON object with tokens, model, timestamp, project/instance info, cache creation/read tokens, etc.).
- Supports the following report types: daily, monthly, session, blocks (5-hour billing windows), and any status-line or default views.
- Features to keep:
  - Date-range filtering (--since, --until)
  - Project/instance filtering and grouping (--project, --instances)
  - Model breakdown (--breakdown)
  - Cache token tracking (creation vs read)
  - Timezone and locale support
  - JSON output (--json)
  - Compact table mode (--compact)
  - Offline mode with cached pricing (--offline)
  - MCP server integration for exposing usage data
  - Colorful terminal tables with proper formatting
  - Ultra-small binary size (prefer minimal dependencies)
- Command-line interface using the existing flags and subcommands where possible.
- Support for custom data path (--path) and default Claude data directory.

**Performance Goals (critical):**
The current implementation is dreadfully slow on large datasets. Rewrite it to be as fast as possible, targeting at least 5-10x speedup on files with tens of thousands to millions of lines.

**Key optimization strategies (use as appropriate, prioritize simplicity):**
1. **Streaming & Memory Efficiency** — Never load the entire file(s) into memory. Use bufio.Scanner or line-by-line reading.
2. **Single-Pass Parsing** — Perform all aggregation in one pass over the data using efficient accumulators.
3. **Efficient Parsing** — Parse only needed fields. Use encoding/json efficiently.
4. **Optimized Data Structures** — Use maps, structs with appropriate types for aggregation. Use fast date handling (time package).
5. **Smart Parallelism** — Process multiple files concurrently using goroutines where beneficial, with result merging (sync.WaitGroup, channels).
6. **Minimize Overhead** — Reduce unnecessary allocations and computations.
7. **Fast Output** — Optimize terminal and JSON output for speed.
8. **Efficient Caching** — O(1) lookups for pricing.

**Technical Stack Recommendations (prioritize minimalism and simplicity):**
- Pure Go.
- Minimal dependencies: use standard library wherever possible (bufio, encoding/json, flag or cobra minimally, time, sync, etc.). For colors and tables consider small focused libraries or implement simple ANSI/color support.
- Focus on maintainability and "just works" over micro-optimizations that add complexity.
- Produce a fast, static binary suitable for distribution.

**Implementation Structure:**
- Modular architecture: separate parser, aggregator, formatter, CLI, and mcp packages.
- Robust error handling for malformed lines (skip gracefully with optional verbose logging).
- Comprehensive tests for correctness (unit tests for aggregation logic with sample JSONL data using Go's testing package).
- Maintain the same CLI interface and output format as closely as possible.

**Deliverables:**
Provide the complete rewritten codebase, organized clearly (e.g., cmd/ccusage/main.go, internal/parser/parser.go, internal/aggregator/aggregator.go, etc.).
Include go.mod with minimal dependencies.
Add a brief performance comparison section in the README (expected speedup).
Ensure the tool produces a small, fast static binary.

Start by outlining the high-level architecture and key optimizations, then provide the full code.

Begin now.
