# HTML5 Server-Sent Events client for Golang

[![Build Status](https://travis-ci.org/mubit/sse.svg?branch=master)](https://travis-ci.org/mubit/sse)
[![GoDoc](https://godoc.org/github.com/mubit/sse?status.svg)](https://godoc.org/github.com/mubit/sse)

This package provides a fast Decoder to consume ServerSentEvents (SSE) in
compliance to the the [HTML5 standard](https://html.spec.whatwg.org/multipage/comms.html).

## Installing

`go get github.com/mubit/sse`

### Running tests & benchmarks

`cd $GOPATH/src/github.com/mubit/sse && go test ./...`

`cd $GOPATH/src/github.com/mubit/sse/benchmarks && go test -bench=.`

# Considerations

- Not used in production, that I am aware.

- Decoder performance could drop due memory copy when the SSE stream
to be consumed provides line feeds using CRLF (`\r\n`), the library
performs best when LF (`\n`) is used.

- This package intends to be spec compliant, however: it might not be
there yet. If you find a bug, do not hesitate to report it.
