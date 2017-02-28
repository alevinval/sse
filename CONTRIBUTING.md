# Contributing

Contributions are welcome granted that:

- HTML5 server-sent events specification is respected.
- Existing tests do not break.
- Source code is enhanced.
- Code is formatted with `gofmt` (or `goimports`)

For bug fixes (excluding bugs in tests) also:

- Include a test that reproduces the bug, and passes.

# Performance improvements

Performance improvements to the library should be complimented with a benchmark
comparison. The benchmarks should run at least 10 times. Use a tool such as 
benchstat to ensure the results are statistically relevant.

```
go test -run=none -bench=. -c=10 github.com/go-rfc/sse/benchmarks > old.txt
go test -run=none -bench=. -c=10 github.com/go-rfc/sse/benchmarks > new.txt
benchstat old.txt new.txt
```

Thanks for the contributions.
