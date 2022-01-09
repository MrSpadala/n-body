package main

import (
    "fmt"
    "os"
    "math"
    "sync"
    "bufio"
    "strconv"
    "encoding/json"
)

func main() {
    fmt.Println("hello world")
    mainLoop()
}

// ROADMAP:
// 1. simulate gravity
// 2. basic print on image
// 3. collisions

// Simulation
const (
    n_workers         = 1
    n_bodies          = 100
    sim_steps uint64  = 20
    sim_step  float64 = 1 //seconds
)

// Environment
const (
    G = 0.01
)

type body struct {
    x    float64
    y    float64
    vx   float64
    vy   float64
    ax   float64
    ay   float64
    mass float64
}

type bodyJson struct {
    X    float64
    Y    float64
    Vx   float64
    Vy   float64
    Ax   float64
    Ay   float64
    Mass float64
}

func (b *body) tobodyJson() bodyJson {
    return bodyJson{X: b.x, Y: b.y, Vx: b.vx, Vy: b.vy, Ax: b.ax, Ay: b.ay, Mass: b.mass}
}

func (b *body) print() {
    fmt.Println("x ", b.x)
    fmt.Println("y ", b.y)
    fmt.Println("vx ", b.vx)
    fmt.Println("vy ", b.vy)
    fmt.Println("ax ", b.ax)
    fmt.Println("ay ", b.ay)
}

func (b *body) copy() body {
    return body{x: b.x, y: b.y, vx: b.vx, vy: b.vy, ax: b.ax, ay: b.ay, mass: b.mass}
}

func populateRange(c chan int, n int) {
    go func() {
        for i := 0; i < n; i++ {
            c <- i
        }
        close(c)
    }()
}

func simInit() [n_bodies]body {
    const offset float64 = 20
    const step float64 = 10
    bodies := [n_bodies]body{}
    for i := 0; i < n_bodies; i++ {
        bodies[i] = body{x: offset + step*float64(i), y: offset, mass: 1.0}
    }
    return bodies
}

func mainLoop() {
    // Init bodies
    bodies := simInit()
    bodies_next := [n_bodies]body{}

    for i := uint64(0); i < sim_steps; i++ {
        go dump(i, &bodies)

        step(i, &bodies, &bodies_next)
        
        bodies = bodies_next
        bodies[0].print()
        bodies_next = [n_bodies]body{}
    }
}

func step(i_step uint64, bodies *[n_bodies]body, bodies_next *[n_bodies]body) {
    fmt.Println("Starting step", i_step)
    indices := make(chan int)
    populateRange(indices, n_bodies)

    var wg sync.WaitGroup
    for i := 0; i < n_workers; i++ {
        wg.Add(1)
        go calcGravity(indices, bodies, bodies_next, &wg)
    }
    wg.Wait()
}

func calcGravity(indices chan int, bodies *[n_bodies]body, bodies_next *[n_bodies]body, wg *sync.WaitGroup) {
    defer wg.Done()
    for i_body := range indices {
        var fx float64 = 0
        var fy float64 = 0
        for j_body := 0; j_body < n_bodies; j_body++ {
            if j_body == i_body {
                continue
            }
            tmpx := bodies[i_body].x - bodies[j_body].x
            tmpy := bodies[i_body].y - bodies[j_body].y
            tmpG := - G * bodies[i_body].mass * bodies[j_body].mass
            tmpDen := math.Pow(math.Abs(tmpx), 3) + math.Pow(math.Abs(tmpy), 3)
            fx += (tmpG / tmpDen) * tmpx
            fy += (tmpG / tmpDen) * tmpy
        }
        bodies_next[i_body] = bodies[i_body].copy()
        b := &bodies_next[i_body]
        // Update gravity acceleration
        b.ax += fx / b.mass
        b.ay += fy / b.mass
        // Update velocity
        b.vx += b.ax * sim_step
        b.vy += b.ay * sim_step
        // Update position
        b.x += b.vx * sim_step
        b.y += b.vy * sim_step
    }
}

func dump(i_step uint64, bodies *[n_bodies]body) {
    // Jsonize bodies at the start of step i_step
    bodies_json := [n_bodies]bodyJson{}
    for i := 0; i < n_bodies; i++ {
        bodies_json[i] = bodies[i].tobodyJson()
    }
    f, _ := os.Create("steps/"+strconv.FormatUint(i_step, 10)+".json")
    defer f.Close()
    w := bufio.NewWriter(f)
    data, _ := json.Marshal(bodies_json)
    w.Write(data)
}

