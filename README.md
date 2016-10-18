# HTML5 Server-Sent Events client for Golang

[![Build Status](https://travis-ci.org/mubit/sse.svg?branch=master)](https://travis-ci.org/mubit/sse)
[![GoDoc](https://godoc.org/github.com/mubit/sse?status.svg)](https://godoc.org/github.com/mubit/sse)

## What is sse?

Library to consume SSE streams according to the HTML5 standard.

## Installing

`go get github.com/mubit/sse`

### Running tests & benchmarks

`cd $GOPATH/src/github.com/mubit/sse && go test ./...`

`cd $GOPATH/src/github.com/mubit/sse/benchmarks && go test -bench=.`

### Benchmarks

````
BenchmarkDecodeEmptyEvent-4                  	 2000000	       804 ns/op	      64 B/op	       1 allocs/op
BenchmarkDecodeEmptyEventWithIgnoredLine-4   	 2000000	       892 ns/op	      64 B/op	       1 allocs/op
BenchmarkDecodeShortEvent-4                  	 2000000	       898 ns/op	      80 B/op	       2 allocs/op
BenchmarkDecode8kEvent-4                     	  200000	      6199 ns/op	    8256 B/op	       2 allocs/op
BenchmarkDecode16kEvent-4                    	  100000	     10309 ns/op	   16450 B/op	       2 allocs/op
PASS
ok  	github.com/mubit/sse/benchmarks	12.341s
````

Benchmarked against a MacBook Air 2013 (1,3 GHz Intel Core i5, 8 GB 1600 MHz DDR3)
