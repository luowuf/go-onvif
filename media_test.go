package onvif

import (
	"fmt"
	"log"
	"testing"
)

func TestGetProfiles(t *testing.T) {
	log.Println("Test GetProfiles")

	deviceServices, err := testDevice.GetServices()

	if err != nil {
		t.Error(err)
		return
	}
	res, err := testDevice.GetProfiles(deviceServices)
	if err != nil {
		t.Error(err)
		return
	}

	js := prettyJSON(&res)
	fmt.Println(js)
}

func TestGetStreamURI(t *testing.T) {
	log.Println("Test GetStreamURI")

	deviceServices, err := testDevice.GetServices()
	if err != nil {
		t.Error(err)
		return
	}

	mediaProfiles, err := testDevice.GetProfiles(deviceServices)
	if err != nil {
		t.Error(err)
		return
	}
	for _, mediaProfile := range mediaProfiles {
		fmt.Println(mediaProfile)

		mediaURI, err := testDevice.GetStreamURI(deviceServices, mediaProfile.Token, "UDP")
		if err != nil {
			t.Error(err)
		}
		js := prettyJSON(&mediaURI)
		fmt.Println(js)
	}

}
