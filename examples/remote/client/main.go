package main

import (
	"time"

	"github.com/charlesderek/actor-model/actor"
	"github.com/charlesderek/actor-model/examples/remote/msg"
	"github.com/charlesderek/actor-model/remote"
)

func main() {
	e := actor.NewEngine()
	r := remote.New(e, remote.Config{ListenAddr: "127.0.0.1:3000"})
	e.WithRemote(r)

	pid := actor.NewPID("127.0.0.1:4000", "server")
	for {
		e.Send(pid, &msg.Message{Data: "hello!"})
		time.Sleep(time.Second)
	}
}
