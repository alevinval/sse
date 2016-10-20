# HTML5 Server-Sent Events client for Golang

[![Build Status](https://travis-ci.org/mubit/sse.svg?branch=master)](https://travis-ci.org/mubit/sse)
[![GoDoc](https://godoc.org/github.com/mubit/sse?status.svg)](https://godoc.org/github.com/mubit/sse)

This package provides a fast Decoder to consume ServerSentEvents (SSE) in
compliance to the the HTML5 standard.

## Installing

`go get github.com/mubit/sse`

### Running tests & benchmarks

`cd $GOPATH/src/github.com/mubit/sse && go test ./...`

`cd $GOPATH/src/github.com/mubit/sse/benchmarks && go test -bench=.`

# Considerations

- Not used in production tested, that I am aware.

- Decoder performance could drop due memory copy when the SSE stream
to be consumed provides line feeds using CRLF (`\r\n`), if you can,
rather use LF (`n`) to avoid the penalty.

- This package intends to be fully spec compliant, however, it might not be
there yet. If you find one, do not hesitate to file a bug.
