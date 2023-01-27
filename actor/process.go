package actor

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/charlesderek/actor-model/log"
)

type processer interface {
	PID() *PID
	Send(*PID, any)
	Shutdown()
}

type process struct {
	Opts

	inbox    chan any
	context  *Context
	pid      *PID
	restarts int
	quitch   chan struct{}
}

func newProcess(e *Engine, cfg Opts) *process {
	pid := NewPID(e.address, cfg.Name, cfg.Tags...)
	ctx := newContext(e, pid)
	p := &process{
		pid:     pid,
		inbox:   make(chan any, 1000),
		Opts:    cfg,
		context: ctx,
		quitch:  make(chan struct{}, 1),
	}
	p.start()
	return p
}

func (p *process) start() *PID {
	recv := p.Producer()
	if p.WithHooks {
		recv = hookReceiver{recv}
	}
	p.inbox <- Started{}

	go func() {
		defer func() {
			if err := recover(); err != nil {
				if msg, ok := err.(*InternalError); ok {
					time.Sleep(time.Second * 1)
					log.Errorw(msg.From, log.M{
						"error": msg.Err,
					})
					p.start()
					return
				}

				p.restarts++
				if p.restarts == p.MaxRestarts {
					log.Errorw("[PROCESS] max restarts exceeded, shutting down...", log.M{
						"pid": p.pid,
					})
					close(p.quitch)
				}

				log.Errorw("[PROCESS] actor restarting", log.M{
					"n":           p.restarts,
					"maxRestarts": p.MaxRestarts,
					"pid":         p.pid,
					"reason":      err,
				})
				fmt.Println(string(debug.Stack()))
				p.start()
			}
		}()

	loop:
		for {
			select {
			case msg := <-p.inbox:
				switch m := msg.(type) {
				case *WithSender:
					p.context.sender = m.Sender
					p.context.message = m.Message
				default:
					p.context.message = m
				}
				recv.Receive(p.context)
			case <-p.quitch:
				close(p.inbox)
				break loop
			}
		}
		p.cleanup()
	}()

	return p.pid
}

func (p *process) cleanup() {
	p.context.engine.registry.remove(p.pid)
	// We are a child if the parent context is not nil
	if p.context.parentCtx != nil {
		p.context.parentCtx.children.Delete(p.Name)
	}
	// We are a parent if we have children running
	if p.context.children.Len() > 0 {
		p.context.children.ForEach(func(name string, pid *PID) {
			p.context.engine.Poison(pid)
			log.Tracew("[PROCESS] shutting down child", log.M{
				"pid":   p.pid,
				"child": pid,
			})
		})
	}
	log.Tracew("[PROCESS] inbox shutdown", log.M{
		"pid": p.pid,
	})
}

func (p *process) PID() *PID            { return p.pid }
func (p *process) Send(_ *PID, msg any) { p.inbox <- msg }
func (p *process) Shutdown()            { close(p.quitch) }
