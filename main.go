package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"strconv"
	"sync"
)

/*
The program is divided into two steps:
	1. Simulation: the motion of the particles is simulated. At each simulation step,
		the position of the particles is saved on disk
	2. Rendering: it reads the files generated during the simulation and generates one image per frame.
*/
func main() {
	// Simulation
	mainLoop()
	// Rendering
	drawAll(380, 420, 1480, 1520)
}

// Simulation
const (
	n_workers         = 16    //number of parallel workers
	n_bodies          = 10000 //number of bodies
	sim_steps uint64  = 40    //simulation steps
	sim_step  float64 = 0.2   //duration in secods of each simulation step
)

// Environment
const (
	G        = 0.000013 //gravity constant
	min_dist = 0.01     //minimum distance of particles on which calculate gravity
)

// Image output heigth and width
const (
	h = 900
	w = 900
)

// Misc
const (
	log_step_sim    = 1
	log_step_render = 20
)

type body struct {
	x    float64
	y    float64
	vx   float64
	vy   float64
	mass float64
}

// This struct is only used when exporting on disk the result of a simulation
type point struct {
	X float64
	Y float64
}

func (b *body) toPoint() point {
	return point{X: b.x, Y: b.y}
}

func (b *body) print() {
	fmt.Println("x ", b.x)
	fmt.Println("y ", b.y)
	fmt.Println("vx ", b.vx)
	fmt.Println("vy ", b.vy)
}

func (b *body) copy() body {
	return body{x: b.x, y: b.y, vx: b.vx, vy: b.vy, mass: b.mass}
}

// Function that populates a int channel with numbers from 0 to n-1 and then closes channel
func populateRange(c chan int, n int) {
	go func() {
		for i := 0; i < n; i++ {
			c <- i
		}
		close(c)
	}()
}

// Initialize bodies for the simulation
func simInit() [n_bodies]body {
	bodies := [n_bodies]body{}

	// Single rotating disc
	bodies_ := RotatingDisc(5, 400, 1500, 0, 0, 0.015*math.Pi, n_bodies)
	for i := 0; i < n_bodies; i++ {
		bodies[i] = bodies_[i]
	}

	// Double rotating disc
	/*
		if n_bodies%2 != 0 {
			panic("even number of bodies required")
		}
		bodies_1 := RotatingDisc(5, 400, 1500, 0, 0, 0.01*math.Pi, n_bodies/2)
		bodies_2 := RotatingDisc(5, 370, 1500, 0.6, 0, 0.01*math.Pi, n_bodies/2)
		for i := 0; i < n_bodies; i++ {
			if i < n_bodies/2 {
				bodies[i] = bodies_1[i]
			} else {
				bodies[i] = bodies_2[i-n_bodies/2]
			}
		}
	*/

	// Three masses
	/*
	   bodies[0] = body{x: 500, y: 50, mass: 100000, vx: 3}
	   bodies[1] = body{x: 480, y: 80, mass: 1000, vx: 4, vy: 0.0}
	   bodies[2] = body{x: 500, y: 1800, mass: 100000000, vx: 0, vy: 0}
	*/

	// Line
	/*
		const offset float64 = 20
		const step float64 = 5
		for i := 0; i < n_bodies; i++ {
			bodies[i] = body{x: offset + 300 + step*float64(i%20), y: offset + float64(i/20)*step, mass: 1.0, vx: 1, vy: 0}
		}
	*/

	// Square
	/*
	   if n_bodies % 5 != 0 {
	       panic("divisible by 5")
	   }
	   for i := 0; i < 5; i++ {
	       for j := 0; j < n_bodies / 5; j++ {
	           bodies[i + j*5] = body{x: (offset + step*float64(j))/4, y: offset + step*float64(i), mass: 1.0}
	       }
	   }
	*/
	return bodies
}

// Simulation loop
func mainLoop() {

	// Init bodies
	bodies := simInit()
	// Keep array of bodies of the next step
	bodies_next := [n_bodies]body{}

	fmt.Println("Simulating", sim_steps, "steps")
	for i_step := uint64(0); i_step < sim_steps; i_step++ {
		if i_step%log_step_sim == 0 {
			fmt.Println("Simulating step", i_step)
		}

		// Safe to asynchronously dump the bodies
		go dump(i_step, &bodies)

		// Channel of indices, one per body. They will be consumed in parallel by the workers
		indices := make(chan int)
		populateRange(indices, n_bodies)

		// Launch and wait workers
		var wg sync.WaitGroup
		for i := 0; i < n_workers; i++ {
			wg.Add(1)
			go calcGravity(indices, &bodies, &bodies_next, &wg)
		}
		wg.Wait()

		// Update bodies for the next iteration
		bodies = bodies_next
		bodies_next = [n_bodies]body{}
	}

	fmt.Println("Simulation finished")
}

func calcGravity(indices chan int, bodies *[n_bodies]body, bodies_next *[n_bodies]body, wg *sync.WaitGroup) {
	defer wg.Done()
	for i_body := range indices {
		var fx float64 = 0
		var fy float64 = 0
		b_i := bodies[i_body]
		for j_body := 0; j_body < n_bodies; j_body++ {
			b_j := &bodies[j_body]
			tmpx := b_i.x - b_j.x
			tmpy := b_i.y - b_j.y
			tmpG := -G * b_i.mass * b_j.mass
			tmpDen := math.Pow(math.Abs(tmpx), 3) + math.Pow(math.Abs(tmpy), 3)
			tmpDen = math.Max(tmpDen, min_dist)
			fx += (tmpG / tmpDen) * tmpx
			fy += (tmpG / tmpDen) * tmpy
		}
		bodies_next[i_body] = bodies[i_body].copy()
		b := &bodies_next[i_body]

		// Calculate acceleration on body
		ax := fx / b.mass
		ay := fy / b.mass
		// Update velocity
		b.vx += ax * sim_step
		b.vy += ay * sim_step
		// Update position
		b.x += b.vx * sim_step
		b.y += b.vy * sim_step
	}
}

// Dump on disk bodies at the given step as the binary array of structs.
func dump(i_step uint64, bodies *[n_bodies]body) {
	points := [n_bodies]point{}
	for i := 0; i < n_bodies; i++ {
		points[i] = bodies[i].toPoint()
	}
	f, e := os.Create("output/steps/" + strconv.FormatUint(i_step, 10) + ".bin")
	if e != nil {
		panic("err in open file")
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	defer w.Flush()
	e = binary.Write(w, binary.LittleEndian, points)
	if e != nil {
		panic("err in binary marshalling")
	}
}
