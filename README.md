# Go client for HTML5 Server-Sent Events

[![Build Status](https://travis-ci.org/go-rfc/sse.svg?branch=master)](https://travis-ci.org/go-rfc/sse)
[![GoDoc](https://godoc.org/github.com/go-rfc/sse?status.svg)](https://godoc.org/github.com/go-rfc/sse)

The package provides fast primitives to manipulate Server-Sent Events as
defined by the [HTML5 standard](https://html.spec.whatwg.org/multipage/comms.html).

Get the package with `go get github.com/go-rfc/sse`

Check the [contributing guidelines](CONTRIBUTING.md) when submitting changes.

## Usage

```go
import "github.com/go-rfc/sse/pkg/eventsource"

es, err := eventsource.New("http://foo.com/stocks/AAPL")

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

```go
eventsource.New("http://foo.com/stocks/AAPL", eventsource.WithBasicAuth("user", "password"))

eventsource.New("http://foo.com/stocks/AAPL", eventsource.WithAuthorizationBearer("token"))
```

Create your own `RequestModifier` in case you need further manipulation of the
underlying HTTP request.
## Decoder

The decoder package allows decoding events from any `io.Reader` source

```go
import "github.com/go-rfc/sse/pkg/decoder"

resp, _ := http.Get("http://foo.com/stocks/AAPL")
decoder = decoder.New(resp.Body) // any io.Reader works
for {
    event, err := decoder.Decode()
}
```

## Encoder

The encoder package allows encoding a stream of events

```go
import "github.com/go-rfc/sse/pkg/encoder"

encoder := encoder.New(out)
encoder.SetRetry(1000)

event := &sse.MessageEvent{
    LastEventID: "",
    Name: "stock-update",
    Data: "AAPL 30.09",
}
encoder.Write(event)
```
