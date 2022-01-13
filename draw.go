package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"os"
	"strconv"
	"sync"
)

const (
	body_size = 0
)

func drawAll(x_start float64, x_end float64, y_start float64, y_end float64) {
	if x_start >= x_end || y_start >= y_end {
		panic("Bad x,y references")
	}

	worker_c := make(chan bool)
	go func() {
		for i := 0; i < n_workers; i++ {
			worker_c <- true
		}
	}()

	var wg sync.WaitGroup
	for i_step := uint64(0); i_step < sim_steps; i_step++ {
		<-worker_c
		wg.Add(1)
		go drawStep(i_step, worker_c, x_start, x_end, y_start, y_end, &wg)
	}
	wg.Wait()
}

func drawStep(i_step uint64, worker_c chan bool, x_start float64, x_end float64,
	y_start float64, y_end float64, wg *sync.WaitGroup) {
	defer func() { worker_c <- true }()
	defer func() { wg.Done() }()

	if i_step%log_step_render == 0 {
		fmt.Println("Rendering step", i_step)
	}

	stepFpath := "output/steps/" + strconv.FormatUint(i_step, 10) + ".json"
	file, err := os.Open(stepFpath)
	if err != nil {
		panic("err in open file for reading")
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		panic("err in reading file")
	}

	points := [n_bodies]point{}
	json.Unmarshal(data, &points)

	// Mantain a separate matrix for pixels, since image has no Get method
	pixels := [(h + 1) * (w + 1)]color.RGBA{}

	upLeft := image.Point{0, 0}
	lowRight := image.Point{w, h}
	img := image.NewRGBA(image.Rectangle{upLeft, lowRight})

	// Black background. ugly as f
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			pixels[x*h+y] = color.RGBA{0, 0, 0, 0xff}
		}
	}

	for _, body := range points {
		if body.X >= x_end || body.Y >= y_end {
			continue
		}
		if body.X < x_start || body.Y < y_start {
			continue
		}
		x := int(float64(w) * (body.X - x_start) / (x_end - x_start))
		y := h - int(float64(h)*(body.Y-y_start)/(y_end-y_start))

		const c_step = uint8(129)
		for dx := -body_size; dx <= body_size; dx++ {
			for dy := -body_size; dy <= body_size; dy++ {
				i_pix := (x+dx)*h + y + dy
				if i_pix < 0 || i_pix >= (h+1)*(w+1) {
					continue
				}
				pix := &(pixels[i_pix])
				if pix.B == 0xff {
					if pix.G == 0xff {
						if pix.R == 0xff {
							// do nothing
						} else {
							if pix.R+c_step < pix.R {
								pix.R = 0xff
							} else {
								pix.R += c_step
							}
						}
					} else {
						if pix.G+c_step < pix.G {
							pix.R = pix.G + c_step - 0xff
							pix.G = 0xff
						} else {
							pix.G += c_step
						}
					}
				} else {
					if pix.B+c_step < pix.B {
						pix.G = pix.B + c_step - 0xff
						pix.B = 0xff
					} else {
						pix.B += c_step
					}
				}
			}
		}
	}

	// Draw pixels on image
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			img.Set(x, y, pixels[x*h+y])
		}
	}

	f, err := os.Create(fmt.Sprintf("output/imgs/%09d.png", i_step))
	if err != nil {
		panic("err opening image to write")
	}
	defer f.Close()

	png.Encode(f, img)
}
