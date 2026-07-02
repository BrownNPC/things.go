package t

import (
	"iter"
	"slices"
)

// Thing is an object that's used in your application.
type Thing interface{ onDestroy() }

// ThingRef is a stable Reference to a Thing.
// This can be stored instead of a pointer.
type ThingRef struct{ idx, gen uint }

// Things is a pool of Things. It is responsible for creating, deleting, reusing things.
// The zero value is usable.
type Things struct {
	things   []Thing
	gen      []uint
	freeList []uint
}

// Reserve will reserve memory for at least n Things.
func (t *Things) Reserve(n int) {
	t.gen = slices.Grow(t.gen, n)
	t.things = slices.Grow(t.things, n)
}

// Add will add Thing to the pool.
// It returns a reference that will only ever point to this Thing.
func (p *Things) Add(e Thing) ThingRef {
	var i uint
	if len(p.freeList) > 0 {
		// Reuse slot
		i = p.freeList[len(p.freeList)-1]
		p.freeList = p.freeList[:len(p.freeList)-1]
		p.things[i] = e
		p.gen[i]++ // Increment generation on reuse
	} else {
		// New slot
		i = uint(len(p.things))
		p.things = append(p.things, e)
		p.gen = append(p.gen, 1) // Start generations at 1 (0 = uninitialized/invalid)
	}
	return ThingRef{idx: i, gen: p.gen[i]}
}

// Things returns an iterator over all the things that have been added.
func (p *Things) Things() iter.Seq2[ThingRef, Thing] {
	return func(yield func(ThingRef, Thing) bool) {
		for i, thing := range p.things {
			if thing != nil {
				ref := ThingRef{idx: uint(i), gen: p.gen[i]}
				if !yield(ref, thing) {
					return
				}
			}
		}
	}
}

// Get retrieves a Thing. Returns nil if the reference is stale or invalid.
func (p *Things) Get(ref ThingRef) Thing {
	if ref.idx >= uint(len(p.gen)) || ref.gen != p.gen[ref.idx] || ref.gen == 0 {
		return nil
	}
	return p.things[ref.idx]
}

// Delete invalidates the passed in ThingRefs, by making the slot the thing is stored in reusable.
//
// WARNING: you should not delete a Thing while you are iterating all the things.
//
// In most cases you should collect all the things to delete inside a slice, and delete them all
// later, like at the very end of the frame in your game.
func (p *Things) Delete(refs ...ThingRef) {
	for _, ref := range refs {
		if ref.idx < uint(len(p.gen)) && ref.gen == p.gen[ref.idx] && ref.gen != 0 {
			p.things[ref.idx].onDestroy()
			p.things[ref.idx] = nil
			p.gen[ref.idx]++ // Increment gen to invalidate existing references
			p.freeList = append(p.freeList, ref.idx)
		}
	}
}
// cache allocations.
type cache[T any] struct {
	objects []*T
}

func (p *cache[T]) Reserve(n int) {
	p.objects = slices.Grow(p.objects, n)
	for range n {
		p.objects = append(p.objects, new(T))
	}
}

func (p *cache[T]) New(v T) any {
	if len(p.objects) == 0 {
		obj := new(T)
		*obj = v
		return obj
	}

	obj := p.objects[len(p.objects)-1]
	p.objects = p.objects[:len(p.objects)-1]
	*obj = v
	return obj
}

func (p *cache[T]) Store(v *T) {
	var zero T
	*v = zero
	p.objects = append(p.objects, v)
}
