package main

import (
	"bytes"
	"fmt"
	"github.com/charmbracelet/log"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/lucasb-eyer/go-colorful"
	"image/color"
	"image/png"
	"math"
	"math/rand"
	"strconv"
	"strings"
)

// Gamma returns the lorentz factor for the given velocity and speed of light
func Gamma(v, c float64) float64 {
	return 1.0 / math.Sqrt(1.0-(v*v)/(c*c))
}

// getDrag is used to calculate a force due to air resistance
// It is used to make the simulation closer to the original arcade game
func getDrag(v Vector, k float64) Vector {
	return v.SetMag(-k * v.SqrMag())
}

// getRedshift is used to convert the redshift data into a valid shader uniform
func getRedshift() [3][100]float64 {
	img, err := png.Decode(bytes.NewReader(spectraData))
	if err != nil {
		log.Fatal("failed to decode spectra data", "error", err)
	}

	out := [3][100]float64{}

	for i := 0; i < 100; i++ {
		col, _ := colorful.MakeColor(img.At(i, 0))

		out[0][i] = col.R
		out[1][i] = col.G
		out[2][i] = col.B
	}

	return out
}

// drawArrow is used to draw an arrow on the screen from the ship to a target
func drawArrow(screen *ebiten.Image, ship Vector, target Vector) {
	// Get the sprite dimensions
	spriteDims := Vector{
		X: float64(arrow.Bounds().Dx()),
		Y: float64(arrow.Bounds().Dy()),
	}

	// Get the screen dimensions
	screenDims := Vector{
		X: float64(screen.Bounds().Dx()),
		Y: float64(screen.Bounds().Dy()),
	}

	// Get the angle between the ship and the target
	angle := ship.AngleBetween(target)

	// Rotate the arrow by the angle
	ops := new(ebiten.DrawImageOptions)
	ops.GeoM.Translate(-spriteDims.X/2, -spriteDims.Y/2)
	ops.GeoM.Rotate(-angle)
	ops.GeoM.Translate(screenDims.X/2, screenDims.Y/2)

	// Give the arrow 50% opacity
	ops.ColorScale.ScaleAlpha(0.5)

	// Draw the arrow to the screen
	screen.DrawImage(arrow, ops)
}

// ScientificNotation is used to convert a float into a string with scientific notation
func scientificNotation(f float64) string {
	// Format the number with scientific notation and split the string on "e"
	parts := strings.Split(fmt.Sprintf("%e", f), "e")

	// Get the mantissa by parsing the first part
	mantissa, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		panic(err)
	}

	// Get the exponent by parsing the second part
	exponent, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		panic(err)
	}

	// Recombine the mantissa and exponent and format it correctly
	return fmt.Sprintf("%.2f×10^%d", mantissa, exponent)
}

// mapRange is used to map a value from one range to another
func mapRange(v, minIn, maxIn, minOut, maxOut float64) float64 {
	return (v-minIn)/(maxIn-minIn)*(maxOut-minOut) + minOut
}

// colorMix is used to mix two colors by a factor x
func colorMix(c1, c2 color.Color, x float64) color.Color {
	// Get uint32 values for each color channel of the two colors
	c1R, c1G, c1B, c1A := c1.RGBA()
	c2R, c2G, c2B, c2A := c2.RGBA()

	// Combine the channels
	return color.RGBA64{
		R: uint16(math.Round(float64(c1R)*x + float64(c2R)*(1.0-x))),
		G: uint16(math.Round(float64(c1G)*x + float64(c2G)*(1.0-x))),
		B: uint16(math.Round(float64(c1B)*x + float64(c2B)*(1.0-x))),
		A: uint16(math.Round(float64(c1A)*x + float64(c2A)*(1.0-x))),
	}
}

// newAsteroidPool returns a new pool of n asteroids.
func newAsteroidPool(n int, area float64) *Pool {
	// Create a new pool
	p := NewPool(n)

	// Populate the pool with n asteroids
	for i := 0; i < n; i++ {
		// 2 in 5 asteroids should be a small asteroid
		if rand.Int63n(5) < 2 {
			radius := rand.Float64()*0.5 + 0.5
			mass := math.Pi * math.Pow(radius, 2)

			p.particles[i] = &Particle{
				Pos:    randUnit().Scl(area),
				Rap:    randUnit().Scl(0.5),
				Vel:    Vector{},
				Acc:    Vector{},
				AngPos: rand.Float64() * 2 * math.Pi,
				AngVel: rand.Float64()*0.25 - 0.125,
				Mass:   mass,
				Radius: radius,
				Gamma:  1,
				Clock:  0,
			}

			p.sprites[i] = 1
		} else {
			radius := rand.Float64() + 1
			mass := math.Pi * math.Pow(radius, 2)

			p.particles[i] = &Particle{
				Pos:    randUnit().Scl(area),
				Rap:    randUnit().Scl(0.5),
				Vel:    Vector{},
				Acc:    Vector{},
				AngPos: rand.Float64() * 2 * math.Pi,
				AngVel: rand.Float64()*0.25 - 0.125,
				Mass:   mass,
				Radius: radius,
				Gamma:  1,
				Clock:  0,
			}

			p.sprites[i] = 0
		}

		p.active[i] = true
	}

	return p
}

// explode is used to place explosion particles in a pool
func explode(pool *Pool, target *Particle) {
	radius := rand.Float64()*0.5 + 0.5
	mass := math.Pi * math.Pow(radius, 2)

	n := 16

	for i := 0; i < n; i++ {
		pool.Activate(
			target.Pos.Add(Vector{1, 0}.Rotate(float64(i)*math.Pi*2.0/float64(n)).Scl(0.2)),
			target.Rap.Add(Vector{1, 0}.Rotate(float64(i)*math.Pi*2.0/float64(n)).Scl(0.5)),
			rand.Float64()*math.Pi*2,
			rand.Float64()*0.25-0.125,
			mass, radius,
			0,
		)
	}

}
