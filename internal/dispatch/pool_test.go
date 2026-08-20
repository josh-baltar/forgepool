package dispatch

import "testing"

func TestPool_MembershipAndEviction(t *testing.T) {
	p := NewPool(1, 2, 3)
	if p.Size() != 3 {
		t.Fatalf("want 3 members, got %d", p.Size())
	}
	if !p.Has(2) {
		t.Fatalf("conn 2 should be a member")
	}
	p.evict(2)
	if p.Has(2) || p.Size() != 2 {
		t.Fatalf("conn 2 should be evicted")
	}
	if got := p.nextOther(1); got != 3 {
		t.Fatalf("nextOther(1) want 3, got %d", got)
	}
}

func TestDispatch_HappyPath(t *testing.T) {
	p := NewPool(1)
	res := Dispatch(p, func(c ConnID) OutcomeKind { return OK })
	if !res.Succeeded || len(res.Attempts) != 1 {
		t.Fatalf("healthy dispatch should succeed in one attempt")
	}
}
