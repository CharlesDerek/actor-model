package actor

import (
	"github.com/charlesderek/actor-model/log"
	cmap "github.com/orcaman/concurrent-map/v2"
)

type registry struct {
	lookup cmap.ConcurrentMap[string, cmap.ConcurrentMap[string, *process]]
}

func newRegistry() *registry {
	return &registry{
		lookup: cmap.New[cmap.ConcurrentMap[string, *process]](),
	}
}

func (r *registry) remove(pid *PID) {
	addrs, ok := r.lookup.Get(pid.Address)
	if !ok {
		return
	}
	addrs.Remove(pid.ID)
	if addrs.Count() == 0 {
		r.lookup.Remove(pid.Address)
	}
}

func (r *registry) add(proc *process) {
	pid := proc.pid
	if addrs, ok := r.lookup.Get(pid.Address); ok {
		if !addrs.Has(pid.ID) {
			addrs.Set(pid.ID, proc)
		} else {
			log.Warnw("[REGISTRY] process already registered", log.M{
				"pid": pid,
			})
		}
		return
	}
	addrs := cmap.New[*process]()
	addrs.Set(pid.ID, proc)
	r.lookup.Set(pid.Address, addrs)
}

func (r *registry) get(pid *PID) *process {
	maddr, ok := r.lookup.Get(pid.Address)
	if !ok {
		return nil
	}
	proc, ok := maddr.Get(pid.ID)
	if !ok {
		return nil
	}
	return proc
}

// always local
func (r *registry) getByID(id string) *process {
	addrs, ok := r.lookup.Get("local")
	if !ok {
		return nil
	}
	proc, _ := addrs.Get(id)
	return proc
}
