//go:generate go run github.com/BrownNPC/things.go/cmd@latest
package t

type Player struct {
	X, Y     float32
	Rotation float32
	Health   int
}

type Zombie struct {
	X, Y     float32
	Health  int
	Chasing ThingRef
}

type Bullet struct {
	Damage int
}
