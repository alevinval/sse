# Go client for HTML5 Server-Sent Events

[![Build Status](https://travis-ci.org/mubit/sse.svg?branch=master)](https://travis-ci.org/mubit/sse)
[![GoDoc](https://godoc.org/github.com/go-rfc/sse?status.svg)](https://godoc.org/github.com/go-rfc/sse)
[![GoReport](https://goreportcard.com/badge/github.com/go-rfc/sse)](https://goreportcard.com/report/github.com/go-rfc/sse)

The package provides fast primitives to manipulate Server-Sent Events (SSE) as
standardized in the [HTML5 standard](https://html.spec.whatwg.org/multipage/comms.html).

> **Note**
> Types and interfaces exposed by the package are subjected to change until 
the project stabilises and matures, which is expected to happen by 1.0.0 release.

Install the package running `go get github.com/go-rfc/sse`

Check the [contributing guidelines](CONTRIBUTING.md) when submitting changes.

# Considerations

- Not aware of any use in production for this library.

- Priority is to be fully compliant with the specification. 
A stable 1.0.0 release will promise that.

- Feel free to report any bug or spec misalignment.

