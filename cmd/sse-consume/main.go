package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/alevinval/sse/pkg/eventsource"
)

var (
	username = flag.String("username", "", "username to use for basic auth")
	password = flag.String("password", "", "password to use for basic auth")
	token    = flag.String("token", "", "authorization bearer token")
)

func main() {
	flag.Parse()

	if len(flag.Args()) != 1 {
		flag.Usage()
		return
	}

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)

	opts := []eventsource.RequestModifier{}
	if *username != "" && *password != "" {
		opts = append(opts, eventsource.WithBasicAuth(*username, *password))
	}
	if *token != "" {
		opts = append(opts, eventsource.WithBearerTokenAuth(*token))
	}

	url := flag.Arg(0)
	es, err := eventsource.New(url, opts...)
	if err != nil {
		log.Fatalf("cannot connect %s", err)
	}

	for {
		select {
		case event := <-es.MessageEvents():
			log.Printf("id: %s\nevent: %s\ndata: %s\n\n", event.ID, event.Name, event.Data)
		case status := <-es.ReadyState():
			if status.Err == nil {
				log.Printf("state=%s\n\n", status.ReadyState)
			} else {
				log.Printf("state=%s err=%v\n\n", status.ReadyState, status.Err)
			}
		case <-exit:
			es.Close()
			return
		}
	}
}
