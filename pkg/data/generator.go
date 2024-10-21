package data

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

// Generator provides methods to generate realistic data for traces
type Generator struct {
	rand *rand.Rand
}

// NewGenerator creates a new data generator
func NewGenerator() *Generator {
	return &Generator{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GenerateID generates a random UUIDv4
func (g *Generator) GenerateID() string {
	return uuid.New().String()
}

// GenerateIPAddress generates a random IP address
func (g *Generator) GenerateIPAddress() string {
	return fmt.Sprintf("%d.%d.%d.%d", g.rand.Intn(256), g.rand.Intn(256), g.rand.Intn(256), g.rand.Intn(256))
}

// GeneratePort generates a random port number
func (g *Generator) GeneratePort() int {
	return g.rand.Intn(65536)
}

// GenerateUserID generates a random user ID
func (g *Generator) GenerateUserID() string {
	return fmt.Sprintf("user_%d", g.rand.Intn(1000000))
}

// GenerateOrderID generates a random order ID
func (g *Generator) GenerateOrderID() string {
	return fmt.Sprintf("order_%d", g.rand.Intn(1000000))
}

// GenerateProductID generates a random product ID
func (g *Generator) GenerateProductID() string {
	return fmt.Sprintf("prod_%d", g.rand.Intn(100000))
}

// GenerateLatency generates a random latency between min and max milliseconds
func (g *Generator) GenerateLatency(min, max int) time.Duration {
	return time.Duration(g.rand.Intn(max-min+2)+min) * time.Millisecond
}

// GenerateStatusCode generates a random HTTP status code
func (g *Generator) GenerateStatusCode() int {
	codes := []int{200, 201, 204, 400, 401, 403, 404, 500}
	return codes[g.rand.Intn(len(codes))]
}

// GenerateErrorMessage generates a random error message
func (g *Generator) GenerateErrorMessage() string {
	errors := []string{
		"Connection timeout",
		"Database unavailable",
		"Invalid input",
		"Resource not found",
		"Internal server error",
	}
	return errors[g.rand.Intn(len(errors))]
}

// GenerateFloat32 generates a random float32 between 0 and 1
func (g *Generator) GenerateFloat32() float32 {
	return g.rand.Float32()
}

// GenerateInt generates a random int between 0 and max-1
func (g *Generator) GenerateInt(max int) int {
	return g.rand.Intn(max)
}
