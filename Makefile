
.PHONY: agent-run
buildVersion := $(git describe --tags --abbrev=0)
buildCommit := $(git rev-parse HEAD)
buildDate := $(date +'%Y/%m/%d-%H:%M:%S')
agent-run:
	@go run -ldflags "-X main.buildVersion=$(buildVersion) -X main.buildCommit=$(buildCommit) -X main.buildDate=$(buildDate)" cmd/agent/main.go

.PHONY: server-run
server-run:
	@go run -ldflags "-X main.buildVersion=$(buildVersion) -X main.buildCommit=$(buildCommit) -X main.buildDate=$(buildDate)" cmd/server/main.go -l "DEBUG"

.PHONY: multichecker
multichecker:
	@go run cmd/multichecker/main.go ./...

.PHONY: proto-gen
proto-gen:
	@protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative pkg/proto/metricserv.proto
# -------------------- Benchmarking server --------------------

.PHONY: bench-server-mem-storage
TS := $(shell date +"%d_%m_%y-%T")
bench-server-mem-storage:
	@cd pkg/storage && go test -v -bench . -benchmem -$(type)profile=../../profiles/$(type)_$(TS).out
	@go tool pprof -http=":9090" profiles/$(type)_$(TS).out

.PHONY: bench-server
bench-server: bench-server-mem-storage

# -------------------- Profiling -----------------------
.PHONY: server-pprof
server-pprof:
	@go tool pprof -http=":9090" -seconds=30 http://localhost:8081/debug/pprof/profile

.PHONY: server-profiling
server-profiling: server-run agent-run server-pprof
