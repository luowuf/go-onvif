package onvif

import (
	"fmt"
	"log"
	"testing"
)

func TestStartDiscovery(t *testing.T) {
	log.Println("Test StartDiscovery")

	devices, err := StartDiscovery(10000)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(devices)
}
