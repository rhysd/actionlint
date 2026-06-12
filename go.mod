module github.com/rhysd/actionlint

go 1.25.0

require (
	github.com/bmatcuk/doublestar/v4 v4.10.0
	github.com/fatih/color v1.19.0
	github.com/google/go-cmp v0.7.0
	github.com/mattn/go-colorable v0.1.14
	github.com/mattn/go-runewidth v0.0.21
	github.com/mattn/go-shellwords v1.0.12
	github.com/robfig/cron/v3 v3.0.1
	github.com/yuin/goldmark v1.7.17
	go.yaml.in/yaml/v4 v4.0.0-rc.3
	golang.org/x/net v0.53.0
	golang.org/x/sync v0.20.0
	golang.org/x/sys v0.43.0
)

require (
	github.com/BurntSushi/toml v1.4.1-0.20240526193622-a339e1f7089c // indirect
	github.com/clipperhouse/uax29/v2 v2.7.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	golang.org/x/exp/typeparams v0.0.0-20231108232855-2478ac86f678 // indirect
	golang.org/x/mod v0.35.0 // indirect
	golang.org/x/telemetry v0.0.0-20260421165255-392afab6f40e // indirect
	golang.org/x/tools v0.44.0 // indirect
	golang.org/x/vuln v1.3.0 // indirect
	honnef.co/go/tools v0.7.0 // indirect
)

tool (
	golang.org/x/vuln/cmd/govulncheck
	honnef.co/go/tools/cmd/staticcheck
)
