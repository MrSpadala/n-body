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
)

const (
	body_size = 1
)

func drawAll(x_start float64, x_end float64, y_start float64, y_end float64, h int, w int) {
	if x_start >= x_end || y_start >= y_end {
		panic("Bad x,y references")
	}

	worker_c := make(chan bool)
	go func() {
		for i := 0; i < n_workers; i++ {
			worker_c <- true
		}
	}()

	for i_step := uint64(0); i_step < sim_steps; i_step++ {
		<-worker_c
		go drawStep(i_step, worker_c,
			x_start, x_end, y_start, y_end, h, w)
	}
}

func drawStep(i_step uint64, worker_c chan bool, x_start float64, x_end float64,
		y_start float64, y_end float64, h int, w int) {
	defer func() { worker_c <- true }()

	if i_step%20 == 0 {
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

	bodies_json := [n_bodies]bodyJson{}
	json.Unmarshal(data, &bodies_json)

	upLeft := image.Point{0, 0}
	lowRight := image.Point{w, h}
	img := image.NewRGBA(image.Rectangle{upLeft, lowRight})

	// ugly as f
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			img.Set(x, y, color.Black)
		}
	}

	for i_body, body := range bodies_json {
		if body.X >= x_end || body.Y >= y_end {
			continue
		}
		if body.X < x_start || body.Y < y_start {
			continue
		}
		x := int(float64(w) * (body.X - x_start) / x_end)
		y := h - int(float64(h)*(body.Y-y_start)/y_end)

		for dx := -body_size; dx <= body_size; dx++ {
			for dy := -body_size; dy <= body_size; dy++ {
				if i_body == 0 {
					img.Set(x+dx, y+dy, color.RGBA{0xff, 0xff, 0x0, 0xff})
				} else if i_body == 1 {
					img.Set(x+dx, y+dy, color.RGBA{0x0, 0xff, 0x0, 0xff})
				} else if i_body == 2 {
					img.Set(x+dx, y+dy, color.RGBA{0xff, 0x0, 0x0, 0xff})
				} else {
					img.Set(x+dx, y+dy, color.White)
				}
			}
		}
	}

	f, err := os.Create(fmt.Sprintf("output/imgs/%09d.png", i_step))
	if err != nil {
		panic("err opening image to write")
	}
	defer f.Close()

	png.Encode(f, img)
}
