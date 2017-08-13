package api

// abstracts a common interface that can be run from cmd/client/main.go
type Client interface {
	Run()
}
