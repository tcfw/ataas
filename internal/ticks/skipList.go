package ticks

import (
	"time"

	"github.com/valyala/fastrand"
)

const (
	maxL = 10
)

type skipListLink struct {
	next *skipListNode //prev
}

type skipListNode struct {
	ts     time.Time
	offset uint64
	tower  [maxL]skipListLink
}

type skipList struct {
	head *skipListNode
	last [maxL]*skipListNode
}

func (sk *skipList) search(ts time.Time) *skipListNode {
	p := sk.head
	i := 0

	if p == nil {
		return nil
	}

	var isAfter bool
	for i < maxL {
		next := p.tower[i].next
		if next != nil {
			isAfter = ts.After(next.ts) || ts == next.ts
		}

		if next == nil || !isAfter {
			i++
		} else if isAfter {
			p = next
		} else {
			break
		}
	}

	return p
}

func (sk *skipList) coinFlip() bool {
	return fastrand.Uint32n(2) == 1
}

func (sk *skipList) insert(ts time.Time, offset uint64) *skipListNode {
	q := sk.newNode(ts, offset)

	if sk.head == nil {
		//first add
		sk.head = q
		for level := 0; level < maxL; level++ {
			sk.last[level] = q
		}
		return q
	}

	for level := maxL - 1; level >= 0; level-- {
		sk.attach(sk.last[level], q, level)

		sk.last[level] = q

		if sk.coinFlip() {
			break
		}
	}

	return q
}

func (sk *skipList) attach(p, q *skipListNode, h int) {
	q.tower[h].next = p.tower[h].next
	p.tower[h].next = q
	// q.tower[h].prev = p
}

func (sk *skipList) newNode(ts time.Time, offset uint64) *skipListNode {
	p := &skipListNode{
		ts:     ts,
		offset: offset,
		tower:  [maxL]skipListLink{},
	}

	return p
}
