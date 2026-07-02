# Things.go

A dead-simple and performant way to store objects in 100 lines of Go.

# Usage

1. Just copy and paste [things.go](./things.go) into a `things` folder in your project.

2. Make a file such as `kinds.go` in the same folder and paste the follwing contents:

```go
//go:generate go run github.com/BrownNPC/things.go/cmd@latest
package t

type Player struct {
	X, Y     float32
	Rotation float32
	Health   int
}
// add other kinds of structs like a Zombie struct:
```

3. run the command `go generate ./...` in your terminal.

**It must be run every time a struct is add/removed from `kinds.go`.**

Make sure you read the `things.go` file. It's only 100 lines.

# Demo

```go
/*
NOTE: dont forget to run the command:
      go generate ./...
*/
package main

import (
	t "game/things"
	"time"
)

// Things is a pool of Things. It is responsible for creating, deleting, reusing things.
var things t.Things

func main() {
	// You should store a ThingRef instead of a pointer.
	// ThingRefs are stable. Meaning they survive resizing of the underlying array,
	// and allow us to reuse slots in the underlying array without collisions.
	//
	// Reuse of slots is managed by the `things` struct.
	var plrRef t.ThingRef = things.Add(t.New(t.Player{
		X:      200,
		Y:      50,
		Health: 100,
	}))

	// assume this is your game loop
	for {
		// get Player pointer from plrRef
		plr, ok := things.Get(plrRef).(*t.Player)
		if !ok {
			println("Player is dead.")
			break
		}
		plr.Health -= 20
		println("Player took 20 damage")
		if plr.Health <= 0 {
			things.Delete(plrRef)
		}
		// delay
		time.Sleep(time.Millisecond * 50)
	}
	// add 30 Things
	for range 30 {
		things.Add(t.New(t.Player{}))
	}

	// iterate all things that exist and delete them.
	var ToDelete = make([]t.ThingRef, 0)
	for ref, thing := range things.Things() {
		switch thing.(type) {
		case *t.Player:
			// WARNING: deleting things while iterating over them is unsafe.
			// dont do this:
			// things.Delete(ref)

			// accumulate first.
			ToDelete = append(ToDelete, ref)
		}
	}
	// Accumulate, then delete.
	things.Delete(ToDelete...)
	ToDelete = ToDelete[:] // reset slice length
}

```


# Performance

The main performance gain from this approach is efficiency when creating/deleting objects.

If we allocate something, we keep reusing it instead of allocating a new object.

this approach makes the creation and deletion of objects instantaneous.

if you want numbers.

```go
things.Add(t.New(t.Player{})
```
takes **3 nanoseconds** to run.


while this line allocates memory on every call, and does not cache.
It takes **60 nanoseconds** to complete, which is **20x slower**.
```go
things.Add(new(t.Player{})
```
