package sse

type Event struct {
	ID    string
	Name  string
	Data  string
	retry int
}
