# Contributing

Contributions are welcome granted that:

- HTML5 server-sent events spec is respected.
- Source code is enhanced.
- Changes are covered by more tests.
- Existing tests do not break.

Any change must be accompanied by a unit test, in cases of bugs there must
be a specific test that reproduces the bug, and passes.

# Performance improvements

Performance improvements must be accompanied with a benchmark comparison.
The benchmarks should run at least 5 times. Use a tool such as benchstat to
ensure the results are statistically relevant. Include the output of benchstat
in the description of your pull request.

```bash
go test -bench=. -count=5 ./... > old.txt
go test -bench=. -count=5 ./... > new.txt
benchstat old.txt new.txt
```

Thanks for the contributions.
