package main

import (
	"github.com/charmbracelet/log"
	"github.com/hajimehoshi/ebiten/v2"
	"math"
	"time"
)

// Pool is a struct for storing particles
type Pool struct {
	// Arrays storing the particles information
	particles []*Particle // particles
	active    []bool      // Whether a particle is active or not
	lifetimes []time.Time // lifetimes of particles
	sprites   []int

	maxLifetime      time.Duration // How long for particles to live
	enforceLifetime  bool          // Whether to enforce the maxLifetime of particles
	fadeOverLifetime bool          // Whether to fade out particles over their lifetime

	disableCollision bool // Whether to collide with other particles in the same pool

	spriteSheet []*ebiten.Image // Sprites for particles
	drawScale   float64         // Scale for particles
}

// NewPool returns a new pool of n particles.
func NewPool(n int) *Pool {
	log.Debug("creating new pool", "n", n)
	return &Pool{
		particles: make([]*Particle, n),
		active:    make([]bool, n),
		lifetimes: make([]time.Time, n),
		sprites:   make([]int, n),
	}
}

// DisableCollision disables collision with other particles in the same pool.
func (p *Pool) DisableCollision() *Pool {
	log.Debug("pool collisions disabled")
	p.disableCollision = true
	return p
}

// EnforceLifetime enforces the maxLifetime of particles.
func (p *Pool) EnforceLifetime(lifetime time.Duration) *Pool {
	log.Debug("pool lifetime enforcement enabled", "lifetime", lifetime.String())
	p.enforceLifetime = true
	p.maxLifetime = lifetime
	return p
}

// FadeOverLifetime enables fading out particles over their lifetime.
func (p *Pool) FadeOverLifetime() *Pool {
	log.Debug("pool lifetime fading enabled")
	p.fadeOverLifetime = true
	return p
}

// SetSpriteSheet sets the sprites for the particles.
func (p *Pool) SetSpriteSheet(drawScale float64, sprites ...*ebiten.Image) *Pool {
	log.Debug("pool sprites set", "n", len(sprites), "drawScale", drawScale)

	p.drawScale = drawScale
	p.spriteSheet = sprites
	return p
}

// Draw draws all particles in the pool to the screen.
func (p *Pool) Draw(screen *ebiten.Image, relPos, frame Vector) {
	if p.spriteSheet == nil {
		return
	}

	for i := 0; i < len(p.particles); i++ {
		if p.active[i] {
			var colorScale *ebiten.ColorScale

			if p.fadeOverLifetime {
				colorScale = new(ebiten.ColorScale)
				colorScale.ScaleAlpha(1.0 - float32(time.Since(p.lifetimes[i]).Seconds()/p.maxLifetime.Seconds()))
			}

			p.particles[i].Draw(screen, p.spriteSheet[p.sprites[i]], p.drawScale, relPos, frame, colorScale)
		}
	}
}

// Update updates all particles in the pool.
func (p *Pool) Update(frame Vector, c float64, dt float64) {
	activeParticles := make([]*Particle, 0, len(p.particles))

	// Select all active particles
	for i := 0; i < len(p.particles); i++ {
		if p.enforceLifetime && p.active[i] {
			if time.Since(p.lifetimes[i]) > p.maxLifetime {
				_ = p.Deactivate(i)
				continue
			}
		}

		if p.active[i] {
			activeParticles = append(activeParticles, p.particles[i])
		}
	}

	// Update positions
	UpdatePositions(activeParticles, frame, c, dt)

	// If collisions are enabled, solve collisions
	if !p.disableCollision {
		SolveCollisions(activeParticles, frame, c)
	}
}

// UpdateWith updates all particles in the pool, plus solves collisions with the given particles.
func (p *Pool) UpdateWith(frame Vector, c float64, dt float64, particles ...*Particle) {
	activeParticles := make([]*Particle, 0, len(p.particles)+len(particles))

	// Select all active particles
	for i := 0; i < len(p.particles); i++ {
		if p.enforceLifetime && p.active[i] {
			if time.Since(p.lifetimes[i]) > p.maxLifetime {
				_ = p.Deactivate(i)
				continue
			}
		}

		if p.active[i] {
			activeParticles = append(activeParticles, p.particles[i])
		}
	}

	// Update positions
	UpdatePositions(activeParticles, frame, c, dt)

	// Add the given particles to the active particles
	activeParticles = append(activeParticles, particles...)

	// If collisions are enabled, solve collisions
	if !p.disableCollision {
		SolveCollisions(activeParticles, frame, c)
	}
}

// Activate activates a particle with the given parameters. Returns true on success
func (p *Pool) Activate(pos, rap Vector, angPos, angVel, mass, radius float64, sprite int) {
	log.Debug("activating particle", "pos", pos.String(), "rap", rap.String(), "angPos", angPos, "angVel", angVel, "mass", mass, "radius", radius, "sprite", sprite)

	// Search for inactive particles to activate
	for i := 0; i < len(p.particles); i++ {
		// If the particle is inactive, activate it
		if !p.active[i] {
			p.active[i] = true
			p.lifetimes[i] = time.Now()
			p.sprites[i] = sprite
			p.particles[i] = &Particle{
				Pos:    pos,
				Rap:    rap,
				Vel:    Vector{},
				Acc:    Vector{},
				AngPos: angPos,
				AngVel: angVel,
				Mass:   mass,
				Radius: radius,
				Gamma:  1,
				Clock:  0,
			}
			return
		}
	}

	log.Debug("failed to find inactive particle")

	// If there are no more inactive particles, create a new one
	p.appendNew(pos, rap, angPos, angVel, mass, radius, sprite)
}

// appendNew adds a new particle to the pool by resizing the pool, to be used when the pool is full.
func (p *Pool) appendNew(pos, rap Vector, angPos, angVel, mass, radius float64, sprite int) {
	log.Debug("appending new particle", "pos", pos.String(), "rap", rap.String(), "angPos", angPos, "angVel", angVel, "mass", mass, "radius", radius, "sprite", sprite)
	p.active = append(p.active, true)
	p.lifetimes = append(p.lifetimes, time.Now())
	p.sprites = append(p.sprites, sprite)
	p.particles = append(p.particles, &Particle{
		Pos:    pos,
		Rap:    rap,
		Vel:    Vector{},
		Acc:    Vector{},
		AngPos: angPos,
		AngVel: angVel,
		Mass:   mass,
		Radius: radius,
		Gamma:  1,
		Clock:  0,
	})
}

// Deactivate deactivates a particle with the given index. Returns true on success
func (p *Pool) Deactivate(i int) bool {
	log.Debug("deactivating particle", "i", i)

	if p.active[i] {
		p.active[i] = false
		p.particles[i] = nil
		return true
	}

	return false
}

// PoolCollisions checks if any particles in a pool collide with any particles in another pool
func (p *Pool) PoolCollisions(other *Pool, frame Vector) [][2]int {
	out := make([][2]int, 0)

	for i := 0; i < len(p.particles); i++ {
		if p.active[i] {
			for j := 0; j < len(other.particles); j++ {
				if other.active[j] {
					if p.particles[i].CheckCollision(other.particles[j], frame) {
						out = append(out, [2]int{i, j})
					}
				}
			}
		}
	}

	return out
}

// Collisions checks if a particle in a pool collides with a particle
func (p *Pool) Collisions(particle *Particle, frame Vector) []int {
	out := make([]int, 0)

	for i := 0; i < len(p.particles); i++ {
		if p.active[i] {
			if p.particles[i].CheckCollision(particle, frame) {
				out = append(out, i)
			}
		}
	}

	return out
}

// Active returns all active particles in the pool.
func (p *Pool) Active() []*Particle {
	out := make([]*Particle, 0, len(p.particles))

	for i := 0; i < len(p.particles); i++ {
		if p.active[i] {
			out = append(out, p.particles[i])
		}
	}

	return out
}

// Closest returns the closest particle in the pool to the given position.
func (p *Pool) Closest(pos Vector) Vector {
	cPos := Vector{}
	minSqrDist := math.MaxFloat64

	for i := 0; i < len(p.particles); i++ {
		if p.active[i] {
			sqrDist := p.particles[i].Pos.SqrDist(pos)
			if sqrDist < minSqrDist {
				minSqrDist = sqrDist
				cPos = p.particles[i].Pos
			}
		}
	}

	return cPos
}

// Reset clears the pool
func (p *Pool) Reset() {
	log.Debug("resetting pool")

	for i := 0; i < len(p.particles); i++ {
		if p.active[i] {
			p.particles[i] = nil
			p.active[i] = false
			p.lifetimes[i] = time.Time{}
			p.sprites[i] = 0
		}
	}
}
