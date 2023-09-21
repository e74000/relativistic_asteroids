package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"math"
)

// Particle is a particle in the simulation
type Particle struct {
	Pos Vector // Position of the particle
	Rap Vector // Rapidity of the particle
	Vel Vector // Velocity of the particle
	Acc Vector // Acceleration of the particle

	AngPos float64 // Angular position of the particle
	AngVel float64 // Angular velocity of the particle

	Mass   float64 // Mass of the particle
	Radius float64 // Radius of the particle

	Gamma float64 // Lorentz factor of the particle
	Clock float64 // Time measured by a Clock on the particle
}

// Update Updates the particle's parameters
func (p *Particle) Update(force, frame Vector, c, dt float64) {
	relativeVelocity := p.Vel.Sub(frame)
	p.Gamma = Gamma(relativeVelocity.Mag(), c) // Calculate the lorentz factor of from the velocity
	dtr := dt * p.Gamma                        // Get Δt adjusted by lorentz factor

	p.Clock += dtr // Add Δt to clock

	p.AngPos += p.AngVel * dtr // Update angular position

	// Calculate new properties of the particle
	acc := force.Scl(1 / p.Mass)                                   // Calculate acceleration from force
	rap := p.Rap.Add(p.Acc.Add(acc).Scl(dtr / 2))                  // Update rapidity from acceleration
	vel := p.Rap.SetMag(c * math.Tanh(p.Rap.Mag()/c))              // Calculate velocity from rapidity
	pos := p.Pos.Add(p.Rap.Scl(dtr)).Add(p.Acc.Scl(dtr * dtr / 2)) // Calculate new position with respect to acceleration and rapidity

	// Update all variables
	p.Acc = acc
	p.Rap = rap
	p.Vel = vel
	p.Pos = pos
}

// CheckCollision returns true if the particle collides with another particle whilst accounting for length contraction (approximately)
func (p *Particle) CheckCollision(q *Particle, frame Vector) bool {
	// Calculate the collision axis
	axis := p.Pos.Sub(q.Pos).Unit()

	// Calculate the radii along the collision axis of the two particles whilst accounting for length contraction
	pRadius := p.ScaledRadius() * (1 - p.Vel.Sub(frame).Unit().Dot(axis)*(1-1/p.Gamma))

	qRadius := q.ScaledRadius() * (1 - q.Vel.Sub(frame).Unit().Dot(axis)*(1-1/q.Gamma))

	if math.IsNaN(pRadius) {
		pRadius = p.ScaledRadius()
	}

	if math.IsNaN(qRadius) {
		qRadius = q.ScaledRadius()
	}

	// Calculate the distance between the centers of the two particles
	distance := p.Pos.Sub(q.Pos).Mag()

	// check whether the distance is greater than the sum of the radii
	return distance < pRadius+qRadius
}

// Momentum returns the momentum of the particle
func (p *Particle) Momentum() Vector {
	return p.Vel.Scl(p.Mass * p.Gamma)
}

// MassRel returns the relative mass of the particle
func (p *Particle) MassRel() float64 {
	return p.Mass * p.Gamma
}

// Draw draws the particle on the screen
func (p *Particle) Draw(screen, sprite *ebiten.Image, scale float64, relPos, frame Vector, colorScale *ebiten.ColorScale) {
	// Get the sprite dimensions
	spriteDims := Vector{
		X: float64(sprite.Bounds().Dx()),
		Y: float64(sprite.Bounds().Dy()),
	}

	// Get the screen dimensions
	screenDims := Vector{
		X: float64(screen.Bounds().Dx()),
		Y: float64(screen.Bounds().Dy()),
	}

	axis := p.Vel.Sub(frame).Angle()
	offset := p.Pos.Sub(relPos).Rotate(-axis).Mul(Vector{1 / p.Gamma, 1}).Rotate(axis)

	// Define the drawImageOptions
	ops := new(ebiten.DrawImageOptions)

	// First center the sprite
	ops.GeoM.Translate(-spriteDims.X/2, -spriteDims.Y/2)

	// Then rotate the sprite according to the angle of the particle and the angle of the reference frame
	ops.GeoM.Rotate(p.AngPos - axis)

	// Then scale the particle, according the particle size and the length contraction acting on the particle
	ops.GeoM.Scale((scale/spriteDims.X)*(p.Radius/p.Gamma), (scale/spriteDims.Y)*p.Radius)

	// Rotate back to the correct angle for the particle to be displayed at
	ops.GeoM.Rotate(axis)

	// Apply the color scale if necessary
	if colorScale != nil {
		ops.ColorScale = *colorScale
	}

	// Rotate back to the correct angle
	ops.GeoM.Translate(scale*offset.X+screenDims.X/2, scale*offset.Y+screenDims.Y/2) // Translate to the correct position

	// Draw the sprite to the screen
	screen.DrawImage(sprite, ops)
}

// String returns the string representation of the particle as JSON
func (p *Particle) String() string {
	return fmt.Sprintf(
		"{\n  \"pos\": %v,\n  \"vel\": %v,\n  \"acc\": %v,\n  \"rap\": %v,\n}",
		p.Pos,
		p.Vel,
		p.Acc,
		p.Rap,
	)
}

// ScaledRadius returns the scaled radius of the particle
func (p *Particle) ScaledRadius() float64 {
	return p.Radius * physScale
}

// UpdatePositions updates the positions of particles
func UpdatePositions(particles []*Particle, frame Vector, c, dt float64) {
	for _, particle := range particles {
		particle.Update(Vector{}, frame, c, dt)
	}
}

// SolveCollisions is used to handle collisions between particles
func SolveCollisions(particles []*Particle, frame Vector, c float64) {
	for i := 0; i < len(particles)-1; i++ {
		for j := i + 1; j < len(particles); j++ {
			if particles[i].CheckCollision(particles[j], frame) {
				// Get the distance between the two particles
				distance := particles[i].Pos.Sub(particles[j].Pos).Mag()
				// Get the collision axis of the two particles
				axis := particles[i].Pos.Sub(particles[j].Pos).Unit()
				// Calculate the step size to move the two particles apart
				stepSize := particles[i].ScaledRadius() + particles[j].ScaledRadius() - distance

				// Move the two particles apart by half the step size
				particles[i].Pos = particles[i].Pos.Add(axis.Scl(0.5 * stepSize))
				particles[j].Pos = particles[j].Pos.Sub(axis.Scl(0.5 * stepSize))

				// Calculate the momentum of the system
				momentum := particles[i].Momentum().Add(particles[j].Momentum())

				// Get the center of momentum frame of the system
				CoM := momentum.Scl(1 / (particles[i].MassRel() + particles[j].MassRel()))

				// Get the velocity of particle i in the center of momentum frame
				iVCoM := particles[i].Vel.Sub(CoM).Scl(1 / (1 - particles[i].Vel.Dot(CoM)/(c*c)))

				// Get the velocity of particle j in the center of momentum frame
				jVCoM := particles[j].Vel.Sub(CoM).Scl(1 / (1 - particles[j].Vel.Dot(CoM)/(c*c)))

				// Perform the collision on particle i and transform back to original reference frame
				iv := iVCoM.Neg().Add(CoM).Scl(1 / (1 - iVCoM.Dot(CoM)/(c*c)))

				// Perform the collision on particle j and transform back to original reference frame
				jv := jVCoM.Neg().Add(CoM).Scl(1 / (1 - jVCoM.Dot(CoM)/(c*c)))

				// Update the rapidity of the two particles
				particles[i].Rap = iv.SetMag(c * math.Tanh(iv.Mag()/c))
				particles[j].Rap = jv.SetMag(c * math.Tanh(jv.Mag()/c))

				// Update the velocity of the two particles
				particles[i].Vel = iv
				particles[j].Vel = jv

			}
		}
	}
}
