package t

import (
	"iter"
	"slices"
)

// Thing is an object that's used in your application.
type Thing interface{ onDestroy() }

// ThingRef is a stable Reference to a Thing.
//
// ThingRefs are stable meaning they survive resizing of the underlying array,
// and allow us to reuse slots in the underlying array without collisions.
type ThingRef struct{ index, generation uint }

// Things is a pool of Things. It is responsible for creating, deleting, reusing things.
// The zero value is usable.
type Things struct {
	things      []Thing
	generations []uint
	freeList    []uint
}

// Reserve will reserve memory for at least n Things.
func (t *Things) Reserve(n int) {
	t.generations = slices.Grow(t.generations, n)
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
		p.generations[i]++ // Increment generation on reuse
	} else {
		// New slot
		i = uint(len(p.things))
		p.things = append(p.things, e)
		p.generations = append(p.generations, 1) // Start generations at 1 (0 = uninitialized/invalid)
	}
	return ThingRef{index: i, generation: p.generations[i]}
}

// WARNING: Do not call Add while iterating.
// Things returns an iterator over all Things that have been added.
//
// Delete is safe during iteration.
func (p *Things) Things() iter.Seq2[ThingRef, Thing] {
	return func(yield func(ThingRef, Thing) bool) {
		for i, thing := range p.things {
			if thing != nil {
				ref := ThingRef{index: uint(i), generation: p.generations[i]}
				if !yield(ref, thing) {
					return
				}
			}
		}
	}
}

// Get retrieves a Thing. Returns nil if the reference is stale or invalid.
func (p *Things) Get(ref ThingRef) Thing {
	if ref.index >= uint(len(p.generations)) || ref.generation != p.generations[ref.index] || ref.generation == 0 {
		return nil
	}
	return p.things[ref.index]
}

// Delete invalidates the passed in ThingRefs, by making the slot the thing is stored in reusable.
// It is safe to delete while iterating.
func (p *Things) Delete(refs ...ThingRef) {
	for _, ref := range refs {
		if ref.index < uint(len(p.generations)) && ref.generation == p.generations[ref.index] && ref.generation != 0 {
			p.things[ref.index].onDestroy()
			p.things[ref.index] = nil
			p.generations[ref.index]++ // Increment gen to invalidate existing references
			p.freeList = append(p.freeList, ref.index)
		}
	}
}

// cache allocations.
type cache[T any] struct {
	objects []*T
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
