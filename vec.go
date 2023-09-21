package main

import (
	"fmt"
	"math"
	"math/rand"
)

// randUnit is used to generate a random unit vector
func randUnit() Vector {
	a := rand.Float64() * 2 * math.Pi

	return Vector{
		X: math.Cos(a),
		Y: math.Sin(a),
	}
}

// Vector represents a vector in 2D space
type Vector struct {
	X, Y float64
}

// Add returns the sum of two vectors
func (v Vector) Add(w Vector) Vector {
	return Vector{
		X: v.X + w.X,
		Y: v.Y + w.Y,
	}
}

// Sub returns the difference between two vectors
func (v Vector) Sub(w Vector) Vector {
	return Vector{
		X: v.X - w.X,
		Y: v.Y - w.Y,
	}
}

// Neg returns the negation of a vector
func (v Vector) Neg() Vector {
	return Vector{
		X: -v.X,
		Y: -v.Y,
	}
}

// Abs returns the absolute value of a vector
func (v Vector) Abs() Vector {
	return Vector{
		X: math.Abs(v.X),
		Y: math.Abs(v.Y),
	}
}

// Dot returns the dot product of two vectors
func (v Vector) Dot(w Vector) float64 {
	return v.X*w.X + v.Y*w.Y
}

// Scl returns the scaled vector
func (v Vector) Scl(s float64) Vector {
	return Vector{
		X: v.X * s,
		Y: v.Y * s,
	}
}

// Mul returns element-wise multiplication of two vectors
func (v Vector) Mul(w Vector) Vector {
	return Vector{
		X: v.X * w.X,
		Y: v.Y * w.Y,
	}
}

// Div returns element-wise division of two vectors
func (v Vector) Div(w Vector) Vector {
	return Vector{
		X: v.X / w.X,
		Y: v.Y / w.Y,
	}
}

// Mag returns the magnitude of the vector
func (v Vector) Mag() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

// SqrMag returns the squared magnitude of the vector
func (v Vector) SqrMag() float64 {
	return v.X*v.X + v.Y*v.Y
}

// Dist returns the Euclidean distance between two vectors
func (v Vector) Dist(w Vector) float64 {
	return v.Sub(w).Mag()
}

// SqrDist returns the squared Euclidean distance between two vectors
func (v Vector) SqrDist(w Vector) float64 {
	return v.Sub(w).SqrMag()
}

// Unit returns the unit vector of the vector
func (v Vector) Unit() Vector {
	return v.Scl(1 / v.Mag())
}

// Angle returns the angle of the vector in radians
func (v Vector) Angle() float64 {
	return math.Atan2(v.Y, v.X)
}

// SetMag returns a vector in the same direction with the given magnitude
func (v Vector) SetMag(s float64) Vector {
	if v.Mag() == 0 {
		return v
	}

	return v.Scl(s / v.Mag())
}

func (v Vector) Rotate(theta float64) Vector {
	return Vector{
		X: v.X*math.Cos(theta) - v.Y*math.Sin(theta),
		Y: v.X*math.Sin(theta) + v.Y*math.Cos(theta),
	}
}

// ToUniform returns the vector as a valid shader uniform object
func (v Vector) ToUniform() [2]float64 {
	return [2]float64{v.X, v.Y}
}

// String returns a string representation of the vector
func (v Vector) String() string {
	return fmt.Sprintf("(%.2f, %.2f)", v.X, v.Y)
}

// AngleBetween returns the angle between two vectors in radians
func (v Vector) AngleBetween(w Vector) float64 {
	if v.X < w.X {
		return -math.Acos(Vector{0, -1}.Dot(w.Sub(v).Unit()))
	} else {
		return math.Acos(Vector{0, -1}.Dot(w.Sub(v).Unit()))
	}
}
