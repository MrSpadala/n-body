
## N-Body Simulation

In this repo there is my combined attempt to learn some Go programming and n-body simulation.

## Examples

### Collapsing rotating disc
Simulation of 2 * 10^4 particles, all with the same mass. Initially they are placed uniformly at random on the area of a circle with radius 5 and angular velocity of 0.015π rad/s. 

The angular velocity is low enough so that the disk collapses on itself.

https://user-images.githubusercontent.com/23057237/149504188-a4eb5245-9395-4c5e-aa8e-44102bb22d73.mp4

### Expanding rotating disc
Simulation of 2 * 10^4 particles on a rotating disc. 

The gravity is not enough to keep the disc tight and it expands.

https://user-images.githubusercontent.com/23057237/149547215-89915509-c56d-4967-bea8-a0e009cc933a.mp4


### Three orbiting masses
Three particles with different mass orbiting (similar to moon-earth-sun system). The smaller particle has a mass of 10^3, the medium particle 10^5 and the central particle 10^8. 

Eventually the smaller mass gets too close to the larger one and is blown away.

https://user-images.githubusercontent.com/23057237/149504154-3696ff20-5fc7-469a-a949-8a81c8b5bfa6.mp4

## Implementation

It runs an algorithm quadratic in the number of particles to calculate the force that one particle exerts on all the other ones. It runs multiple workers in parallel.

## How to run

Clone the repo, and then
```
go run .
```
This will simulate the requested steps and render the frames in .png (one frame per simulation step). To create a video of your simulation (requires `ffmpeg`):
```
./make-video
```







