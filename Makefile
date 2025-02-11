.PHONY: agent-run
agent-run:
	@go run cmd/agent/main.go

.PHONY: server-run
server-run:
	@go run cmd/server/main.go

.PHONY: multichecker
multichecker:
	@go run cmd/multichecker/main.go $(ARGS)

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