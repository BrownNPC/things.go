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
	// indexed by ThingRef.idx
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
func (t *Things) Add(thing Thing) ThingRef {
	var i uint
	if len(t.freeList) > 0 {
		// Reuse slot
		// get last element of freeList
		i = t.freeList[len(t.freeList)-1]
		t.freeList = t.freeList[:len(t.freeList)-1]
		t.things[i] = thing
		t.gen[i]++ // Increment generation on reuse
	} else {
		// New slot, add a new thing.
		i = uint(len(t.things))
		t.things = append(t.things, thing)
		t.gen = append(t.gen, 1) // Start generations at 1 (0 = uninitialized/invalid)
	}
	return ThingRef{idx: i, gen: t.gen[i]}
}

// Things returns an iterator over all the things that have been added.
func (t *Things) Things() iter.Seq2[ThingRef, Thing] {
	return func(yield func(ThingRef, Thing) bool) {
		for i, thing := range t.things {
			if thing != nil {
				ref := ThingRef{idx: uint(i), gen: t.gen[i]}
				if !yield(ref, thing) {
					return
				}
			}
		}
	}
}

// Get retrieves a Thing. Returns nil if the reference is stale or invalid.
func (t *Things) Get(ref ThingRef) Thing {
	if ref.idx >= uint(len(t.gen)) || ref.gen != t.gen[ref.idx] || ref.gen == 0 {
		return nil
	}
	return t.things[ref.idx]
}

// Delete invalidates the passed in ThingRefs.
//
// WARNING: you should not delete a Thing while you are iterating over things.
//
// In most cases you should collect all the things to delete inside a slice, and delete them all
// later, like at the very end of the frame in your game.
func (t *Things) Delete(refs ...ThingRef) {
	for _, ref := range refs {
		if ref.idx < uint(len(t.gen)) && ref.gen == t.gen[ref.idx] && ref.gen != 0 {
			// onDestroy method will store the object in a cache.
			t.things[ref.idx].onDestroy()
			t.things[ref.idx] = nil
			t.gen[ref.idx]++ // Increment gen to invalidate existing references
			t.freeList = append(t.freeList, ref.idx)
		}
	}
}

// caches allocations.
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
