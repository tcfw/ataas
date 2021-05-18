package ticks

import (
	"math/rand"
	"time"
)

const (
	maxL = 21
)

type skipListLink struct {
	prev, next *skipListNode
}

type skipListNode struct {
	ts     time.Time
	offset uint64
	tower  [maxL]skipListLink
}

type skipList struct {
	head *skipListNode
}

func (sk *skipList) search(ts time.Time) *skipListNode {
	p := sk.head
	i := maxL - 1

	if p == nil {
		return nil
	}

	for i >= 0 {
		if p.tower[i].next == nil {
			i--
			continue
		}

		next := p.tower[i].next

		if ts.After(next.ts) || ts == next.ts {
			p = next
		} else {
			break
		}
	}

	return p
}

func (sk *skipList) coinFlip() bool {
	return rand.Intn(2) == 1
}

func (sk *skipList) insert(ts time.Time, offset uint64) *skipListNode {
	q := sk.newNode(ts, offset)

	if sk.head == nil {
		//first add
		sk.head = q
		return q
	}

	p := sk.search(ts)
	sk.attach(p, q, 0)

	for i := maxL - 1; i >= 0; i-- {
		if p.tower[i].next != nil || p == sk.head {
			sk.attach(p, q, i)
		} else {
			//move back
			if p.tower[maxL-1].prev != nil {
				p = p.tower[maxL-1].prev
				continue
			}
		}

		if sk.coinFlip() {
			break
		}
	}

	return q
}

func (sk *skipList) attach(p, q *skipListNode, h int) {
	q.tower[h].next = p.tower[h].next
	q.tower[h].prev = p
	p.tower[h].next = q
}

func (sk *skipList) newNode(ts time.Time, offset uint64) *skipListNode {
	p := &skipListNode{
		ts:     ts,
		offset: offset,
		tower:  [maxL]skipListLink{},
	}

	return p
}
