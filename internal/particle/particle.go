package particle

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Particle represents a single particle in the system.
type Particle struct {
	X, Y       float64 // Current position
	VX, VY     float64 // Velocity
	Life       float64 // Remaining lifetime in seconds
	TotalLife  float64 // Initial lifetime
	Size       float32 // Size
	Color      color.Color
	UseGravity bool
}

// System manages a collection of particles.
type System struct {
	Particles []*Particle
	Gravity   float64
}

// NewSystem creates a particle system.
func NewSystem(gravity float64) *System {
	return &System{
		Particles: make([]*Particle, 0, 100), // Pre-allocate some capacity
		Gravity:   gravity,
	}
}

// Update updates all particles in the system.
func (s *System) Update(deltaTime float64) {
	aliveParticles := s.Particles[:0] // Re-slice to keep only alive particles

	for _, p := range s.Particles {
		p.Life -= deltaTime
		if p.Life > 0 {
			p.X += p.VX * deltaTime
			p.Y += p.VY * deltaTime
			if p.UseGravity {
				p.VY += s.Gravity * deltaTime
			}
			aliveParticles = append(aliveParticles, p)
		}
	}
	s.Particles = aliveParticles
}

// Emit creates new particles at a specific location.
type EmitConfig struct {
	X, Y           float64
	Count          int
	UseGravity     bool
	Color          color.Color
	BaseVelocityX  float64
	BaseVelocityY  float64
	VelocitySpread float64
	MinLifetime    float64
	MaxLifetime    float64
	MinSize        float32
	MaxSize        float32
}

func (s *System) Emit(config EmitConfig) {
	for i := 0; i < config.Count; i++ {
		lifetime := config.MinLifetime + rand.Float64()*(config.MaxLifetime-config.MinLifetime)
		angle := rand.Float64() * 2 * math.Pi
		speed := config.VelocitySpread * rand.Float64()
		vx := config.BaseVelocityX + math.Cos(angle)*speed
		vy := config.BaseVelocityY + math.Sin(angle)*speed
		size := config.MinSize + rand.Float32()*(config.MaxSize-config.MinSize)

		p := &Particle{
			X:          config.X,
			Y:          config.Y,
			VX:         vx,
			VY:         vy,
			Life:       lifetime,
			TotalLife:  lifetime,
			Size:       size,
			Color:      config.Color,
			UseGravity: config.UseGravity,
		}
		s.Particles = append(s.Particles, p)
	}
}

// Draw renders all particles.
func (s *System) Draw(screen *ebiten.Image) {
	for _, p := range s.Particles {
		// Calculate alpha based on remaining life for fade effect
		alphaFactor := p.Life / p.TotalLife
		if alphaFactor < 0 {
			alphaFactor = 0
		}
		if alphaFactor > 1 {
			alphaFactor = 1
		}

		// Get original color components
		r, g, b, a := p.Color.RGBA()
		// Modulate original alpha by the life factor
		finalA := uint8(float64(a>>8) * alphaFactor) // Need to shift A down

		// Create the final color with modulated alpha
		finalColor := color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: finalA}

		halfSize := p.Size / 2.0
		vector.DrawFilledRect(screen, float32(p.X-float64(halfSize)), float32(p.Y-float64(halfSize)), p.Size, p.Size, finalColor, false)
	}
}
