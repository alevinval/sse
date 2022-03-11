# Go client for HTML5 Server-Sent Events

[![Build Status](https://travis-ci.org/go-rfc/sse.svg?branch=master)](https://travis-ci.org/go-rfc/sse)
[![GoDoc](https://godoc.org/github.com/go-rfc/sse?status.svg)](https://godoc.org/github.com/go-rfc/sse)

The package provides fast primitives to manipulate Server-Sent Events (SSE) as
defined in the [HTML5 standard](https://html.spec.whatwg.org/multipage/comms.html).

Get the package with `go get github.com/go-rfc/sse`

Check the [contributing guidelines](CONTRIBUTING.md) when submitting changes.

## Usage

```go
es, err := sse.NewEventSource("http://foo.com/stocks/AAPL")

for {
    select {
    case event := <- es.MessageEvents():
        processEvent(event)
    case <- es.ReadyState():
        // You can hook custom logic on ReadyState changes
        continue
    }
}
```

The library includes `WithBasicAuth` and `WithAuthorizationBearer` modifiers.

Create your own `RequestModifier` in case you need further manipulation of the
underlying HTTP request.

```go
sse.NewEventSource("http://foo.com/stocks/AAPL", sse.WithBasicAuth("user", "password"))
sse.NewEventSource("http://foo.com/stocks/AAPL", sse.WithAuthorizationBearer("token"))
```

On the server side, use the encoder to emit the events.

```go
encoder := sse.NewEncoder(out)
encoder.SetRetry(1000)

event := &sse.MessageEvent{
    LastEventID: "",
    Name: "stock-update",
    Data: "AAPL 30.09",
}
encoder.Write(event)
```
