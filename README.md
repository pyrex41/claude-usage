# Claude Usage

A fast CLI tool to analyze your Claude Code usage and costs.

This is a Go rewrite of the original TypeScript `ccusage` tool, providing 5-10x better performance while maintaining full feature parity.

## Features

- **Fast**: Processes large usage logs quickly using concurrent parsing
- **Multiple views**: Daily reports, project-based breakdowns (`--instances`)
- **Flexible filtering**: Filter by date range, project, etc.
- **Multiple output formats**: Pretty tables, compact mode, or JSON
- **Accurate costing**: Proper token-based cost calculation

## Installation

### Quick Install (Recommended)

```bash
# Method 1: Using the install script
git clone https://github.com/pyrex41/claude-usage.git
cd claude-usage
./install.sh

# Method 2: Direct Go install
go install ./cmd/claude-usage
```

### From Makefile

```bash
git clone https://github.com/pyrex41/claude-usage.git
cd claude-usage
make install
```

### Option 2: Build locally

```bash
git clone https://github.com/pyrex41/claude-usage.git
cd claude-usage
go build -o claude-usage ./cmd/claude-usage
sudo mv claude-usage /usr/local/bin/
```

### Option 3: Direct build

```bash
go install github.com/pyrex41/claude-usage/cmd/claude-usage@latest
```

## Usage

### Basic Commands

```bash
# Daily report
claude-usage daily

# Group by project (most useful)
claude-usage daily --instances

# Compact table view
claude-usage daily --compact

# Filter by date range
claude-usage daily --since 20240101 --until 20240331

# Filter by specific project
claude-usage daily --project fg
```

### Other Options

```bash
claude-usage daily --json          # JSON output
claude-usage daily --breakdown     # Detailed model breakdown
claude-usage daily --help          # Show all options
```

**Pro tip:** Use `--instances --compact` for the most readable project-based view.

## Flags

- `--since, -s`: Start date (YYYYMMDD)
- `--until, -u`: End date (YYYYMMDD) 
- `--project, -p`: Filter by project name
- `--instances`: Group by project/instance
- `--compact`: Use compact table format
- `--breakdown, -b`: Show model breakdown
- `--json, -j`: Output as JSON
- `--offline, -O`: Use offline pricing
- `--path`: Path to Claude data (default: `~/.claude/projects`)

## Data Location

Claude Code stores usage data in:
- `~/.claude/projects/{project}/{session-id}.jsonl`

The tool automatically finds and parses all `.jsonl` files in the projects directory.

## Development

### Using Make (recommended)

```bash
make build        # Build binary
make install      # Install to GOPATH/bin
make test         # Run tests
make fmt          # Format code
make sample       # Test with sample data
make real         # Test with real data
make clean        # Clean build artifacts
```

### Manual commands

```bash
go build -o claude-usage ./cmd/claude-usage
go test ./...
go fmt ./...
```

## Project Structure

```
.
├── cmd/claude-usage/     # CLI entrypoint (main.go)
├── internal/
│   ├── parser/           # JSONL parsing + project name extraction
│   ├── aggregator/       # Usage aggregation + cost calculation
│   ├── formatter/        # Pretty table output (go-pretty)
│   ├── types/            # Core data structures
│   └── pricing/          # Token-based cost calculation
├── testdata/             # Sample data for testing
├── Makefile              # Build tasks and helpers
├── install.sh            # Simple installation script
├── README.md
└── go.mod
```

## License

MIT