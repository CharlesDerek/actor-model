package main

import (
	"fmt"

	"github.com/charlesderek/actor-model/actor"
	"github.com/charlesderek/actor-model/examples/remote/msg"
	"github.com/charlesderek/actor-model/remote"
)

type server struct{}

func newServer() actor.Receiver {
	return &server{}
}

func (f *server) Receive(ctx *actor.Context) {
	switch m := ctx.Message().(type) {
	case actor.Started:
		fmt.Println("server has started")
	case *actor.PID:
		fmt.Println("server has received:", m)
	case *msg.Message:
		fmt.Println("got message", m)
	}
}

func main() {
	e := actor.NewEngine()
	r := remote.New(e, remote.Config{ListenAddr: "127.0.0.1:4000"})
	e.WithRemote(r)

	e.Spawn(newServer, "server")
	select {}
}
