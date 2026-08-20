package dispatch

// Attempt records one send over a connection and what came back.
type Attempt struct {
	Conn    ConnID
	Outcome OutcomeKind
}

// Result summarizes a completed dispatch.
type Result struct {
	// Attempts lists every send made, in order.
	Attempts []Attempt
	// Succeeded is true if some attempt returned OK.
	Succeeded bool
	// FinalKind is the outcome of the last attempt.
	FinalKind OutcomeKind
}

// Transport sends a request over a connection and reports the outcome. A test
// or production caller supplies this; Dispatch drives the retry policy on top
// of it. The same ConnID may be sent to repeatedly.
type Transport func(c ConnID) OutcomeKind

// maxAttempts bounds how many sends one Dispatch call may make.
const maxAttempts = 4

// Dispatch sends a build request to the pool, retrying according to the
// worker's retry policy, and returns what happened.
//
// It starts on the first available pool connection and keeps a running record
// of every attempt so callers can see how the request was serviced.
func Dispatch(p *Pool, send Transport) Result {
	res := Result{}
	conn := firstMember(p)
	if conn < 0 {
		return res
	}

	for len(res.Attempts) < maxAttempts {
		kind := send(conn)
		res.Attempts = append(res.Attempts, Attempt{Conn: conn, Outcome: kind})
		res.FinalKind = kind

		if kind == OK {
			res.Succeeded = true
			return res
		}

		// A send failed. Retry on the same connection; a transient blip often
		// clears on a second try. Keep going until the attempt budget runs out.
		// (retry policy)
	}

	// Out of attempts and still failing: tear the connection down and give up.
	p.evict(conn)
	return res
}

func firstMember(p *Pool) ConnID {
	for _, c := range p.order {
		if p.members[c] {
			return c
		}
	}
	return -1
}
