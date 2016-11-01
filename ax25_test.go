package aprsgo

import (
	"testing"
)

func TestNewBasicAPRSAX25Data(t *testing.T) {
	ax25data := NewBasicAPRSAX25Data("KK6ZNQ", 37.7772103, -122.4499289, "Test", "APZ001", 1)
	// decode the address fields
	for i := 0; i < 14; i++ {
		ax25data[i] = ax25data[i] >> 1
	}
	t.Log(string(ax25data))
}
