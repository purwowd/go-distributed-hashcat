# Benchmarks

Simple performance benchmarks untuk distributed hashcat system.

## 🚀 Quick Start

```bash
# Run benchmarks
go test -bench=. -benchmem ./tests/benchmarks/

# Run specific benchmark
go test -bench=BenchmarkAgentCreation -benchmem ./tests/benchmarks/

# Run with custom duration
go test -bench=. -benchtime=5s -benchmem ./tests/benchmarks/
```

## 📊 Available Benchmarks

- **BenchmarkAgentCreation**: Agent creation via HTTP handler
- **BenchmarkJobCreation**: Job creation performance
- **BenchmarkAgentListing**: Agent listing with multiple records
- **BenchmarkDirectAgentCreation**: Direct usecase calls (fastest)
- **BenchmarkLimitedConcurrentAgentCreation**: Concurrent agent creation

## 📈 Sample Results (Apple M3)

```
BenchmarkAgentCreation-8                    72000    18590 ns/op   11075 B/op   84 allocs/op
BenchmarkJobCreation-8                      37804    30760 ns/op   13457 B/op   99 allocs/op  
BenchmarkDirectAgentCreation-8             111704    13757 ns/op    2083 B/op   40 allocs/op
BenchmarkAgentListing-8                     16782    71040 ns/op   44715 B/op  278 allocs/op
BenchmarkLimitedConcurrentAgentCreation-8  367640     4189 ns/op   10487 B/op   75 allocs/op
```

## 🔧 Performance Tips

- Use in-memory databases for fastest setup
- Run benchmarks multiple times for stable results
- Monitor memory allocations (`-benchmem` flag)
- Use `GOMAXPROCS` to control CPU cores usage 
