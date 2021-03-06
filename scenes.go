package main

import (
	"fmt"
	"math"
	"math/rand"
)

/*
Generates particles uniformly inside a rotating circle with the specified radius,
position of the center (cx, cy) and angular velocity of the disc (in rad/s).
A translation velocity can also be specified (v0x, v0y)
*/
func RotatingDisc(radius float64, cx float64, cy float64, v0x float64, v0y float64, w float64, n_particles uint64) []body {
	fmt.Println("Generating rotating disc scene...")
	res := make([]body, n_particles)
	rand := rand.New(rand.NewSource(42))
	for i := uint64(0); i < n_particles; i++ {
		// Generate points uniformly on circle, https://stackoverflow.com/questions/5837572/generate-a-random-point-within-a-circle-uniformly
		r := math.Sqrt(rand.Float64()) * radius
		theta := rand.Float64() * 2 * math.Pi

		// Calculate initial particle position and velocity
		v := w * r
		x := cx + r*math.Cos(theta)
		y := cy + r*math.Sin(theta)
		vx := v0x + -v*math.Sin(theta)
		vy := v0y + v*math.Cos(theta)

		res[i] = body{
			x:    x,
			y:    y,
			vx:   vx,
			vy:   vy,
			mass: 1}
	}
	fmt.Println("Rotating disc scene generated")
	return res
}
