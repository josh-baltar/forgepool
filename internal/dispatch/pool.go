// Package dispatch routes build requests from a CI worker to a pool of warm
// build daemons over persistent connections. Each daemon keeps compiled
// artifacts and analysis caches hot, so keeping a connection alive across
// requests is a large performance win — a cold daemon must re-read and
// re-analyze the whole workspace before it can serve anything.
package dispatch

// ConnID identifies a persistent connection to one warm build daemon.
type ConnID int

// OutcomeKind classifies what happened when a request was sent over a
// connection.
type OutcomeKind int

const (
	// OK means the daemon served the request successfully.
	OK OutcomeKind = iota
	// Stale means the wire failed: the daemon process is gone, the socket was
	// reset, the read hit EOF, or the request timed out before any response.
	// Nothing was computed; the connection is dead.
	Stale
	// BuildErr means the daemon is alive and answered, but the build work
	// itself failed (a compile error, a failing task). The connection is
	// perfectly healthy — the *request* is what failed.
	BuildErr
)

// Pool holds the set of live connections the worker may use. Membership is
// what matters: a connection in the pool is considered usable for the next
// request; a connection removed from the pool has been torn down and its
// daemon's warm caches are lost.
type Pool struct {
	members map[ConnID]bool
	// order records connections in the order they were added, so the next
	// free connection is chosen deterministically.
	order []ConnID
}

// NewPool builds a pool seeded with the given connections (in order).
func NewPool(conns ...ConnID) *Pool {
	p := &Pool{members: map[ConnID]bool{}}
	for _, c := range conns {
		if !p.members[c] {
			p.members[c] = true
			p.order = append(p.order, c)
		}
	}
	return p
}

// Has reports whether c is currently a pool member.
func (p *Pool) Has(c ConnID) bool { return p.members[c] }

// Size returns the number of live connections in the pool.
func (p *Pool) Size() int { return len(p.members) }

// Members returns the current pool members in insertion order.
func (p *Pool) Members() []ConnID {
	out := []ConnID{}
	for _, c := range p.order {
		if p.members[c] {
			out = append(out, c)
		}
	}
	return out
}

// evict removes c from the pool (its daemon connection is torn down).
func (p *Pool) evict(c ConnID) { delete(p.members, c) }

// nextOther returns the first pool member that is not `avoid`, or -1 if there
// is none. Used to pick a different connection than the one that just failed.
func (p *Pool) nextOther(avoid ConnID) ConnID {
	for _, c := range p.order {
		if p.members[c] && c != avoid {
			return c
		}
	}
	return -1
}
