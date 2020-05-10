package gu

// Common Interfaces

type Producer interface {
	Produce(message []byte) error
	ProducerHealth() bool
}
