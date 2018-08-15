package broker

import (
	"testing"
)

func TestAddFlags(t *testing.T) {
	opts := Options{}
	AddFlags(&opts)
}
