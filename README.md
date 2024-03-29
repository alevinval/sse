# Go client for HTML5 Server-Sent Events

[![GoDoc](https://pkg.go.dev/badge/github.com/alevinval/sse/pkg/eventsource)](https://pkg.go.dev/github.com/alevinval/sse/pkg/eventsource)

The package provides fast primitives to manipulate Server-Sent Events as
defined by the [HTML5 spec](https://html.spec.whatwg.org/multipage/server-sent-events.html).

Get the package with `go get github.com/alevinval/sse`

Check the [contributing guidelines](CONTRIBUTING.md) when submitting changes.

## Usage

```go
import "github.com/alevinval/sse/pkg/eventsource"

es, err := eventsource.New("http://foo.com/stocks/AAPL")

for {
    select {
    case event := <-es.MessageEvents():
        log.Printf("[Event] ID: %s\n Name: %s\n Data: %s\n\n", event.ID, event.Name, event.Data)
    case state := <-es.ReadyState():
        log.Printf("[ReadyState] %s (err=%v)", state.ReadyState, err)
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
import "github.com/alevinval/sse/pkg/decoder"

resp, _ := http.Get("http://foo.com/stocks/AAPL")
decoder = decoder.New(resp.Body) // any io.Reader works
for {
    event, err := decoder.Decode()
}
```

## Encoder

The encoder package allows encoding a stream of events

```go
import "github.com/alevinval/sse/pkg/encoder"

event := &base.MessageEvent{
    ID:   "some id",
    Name: "stock-update",
    Data: "AAPL 30.09",
}

out := new(bytes.Buffer)
encoder := New(out)

encoder.WriteComment("example event")
encoder.WriteEvent(event)

```

