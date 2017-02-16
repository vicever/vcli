package shared

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestCatenate(t *testing.T) {

	input := `This 	is 		a broken
	string`

	output := Catenate(input)
	if output != "This is a broken string" {
		t.Error("shared.Catenate() not working as intended")
	}

}

func TestNodeName(t *testing.T) {

	input := filepath.Join("alpha", "beta", "charlie", "delta", "echo", "foxtrot")
	output := NodeName(input)
	if output != "foxtrot" {
		fmt.Println(output)
		t.Error("shared.NodeName() not working as intended")
	}

}

func TestNodeParent(t *testing.T) {

	input := filepath.Join("alpha", "beta", "charlie", "delta", "echo", "foxtrot")
	expected := filepath.Join("alpha", "beta", "charlie", "delta", "echo")
	output := NodeParent(input)
	if output != expected {
		t.Error("shared.NodeParent() not working as intended")
	}
}

func TestValidHypervisor(t *testing.T) {

	input := "alpha"
	output := ValidHypervisor(input)
	if output {
		t.Error("shared.ValidHypervisor() not working as intended")
	}

	list := ListDetectedHypervisors()
	for i := 0; i < len(list)-1; i++ {
		output = ValidHypervisor(list[i])
		if !output {
			t.Error("shared.ValidHypervisor() not working as intended" + fmt.Sprintf("\nINPUT: %s\n", list[i]))
		}
	}

}

func TestValidHypervisorWithHidden(t *testing.T) {

	input := "alpha"
	output := ValidHypervisorWithHidden(input)
	if output {
		t.Error("shared.ValidHypervisor() not working as intended")
	}

	list := ListDetectedHypervisors()
	for i := 0; i < len(list)-1; i++ {
		output = ValidHypervisorWithHidden(list[i])
		if !output {
			t.Error("shared.ValidHypervisor() not working as intended" + fmt.Sprintf("\nINPUT: %s\n", list[i]))
		}
	}

}

func TestIsZip(t *testing.T) {

	zipFile := make([]byte, 0)
	zipFile = []byte{0x50, 0x4B, 0x03, 0x04, 0x0A, 0x00, 0x00, 0x00, 0x00, 0x00, 0x2E, 0x77, 0x4D, 0x4A, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x05, 0x00, 0x1C, 0x00, 0x54, 0x65, 0x73, 0x74, 0x2F, 0x55, 0x54, 0x09, 0x00, 0x03, 0xB8, 0x3C, 0xA1, 0x58, 0xC3, 0x3C, 0xA1, 0x58, 0x75, 0x78, 0x0B, 0x00, 0x01, 0x04, 0xE8, 0x03, 0x00, 0x00, 0x04, 0xE8, 0x03, 0x00, 0x00, 0x50, 0x4B, 0x01, 0x02, 0x1E, 0x03, 0x0A, 0x00, 0x00, 0x00, 0x00, 0x00, 0x2E, 0x77, 0x4D, 0x4A, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x05, 0x00, 0x18, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x10, 0x00, 0xED, 0x41, 0x00, 0x00, 0x00, 0x00, 0x54, 0x65, 0x73, 0x74, 0x2F, 0x55, 0x54, 0x05, 0x00, 0x03, 0xB8, 0x3C, 0xA1, 0x58, 0x75, 0x78, 0x0B, 0x00, 0x01, 0x04, 0xE8, 0x03, 0x00, 0x00, 0x04, 0xE8, 0x03, 0x00, 0x00, 0x50, 0x4B, 0x05, 0x06, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x4B, 0x00, 0x00, 0x00, 0x3F, 0x00, 0x00, 0x00, 0x00, 0x00}
	f, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fail()
	}

	f.Write(zipFile)
	okay := IsZip(f.Name())
	if !okay {
		t.Error("shared.IsZip() not working as intended")
	}
}

func TestVCFGHealthCheck(t *testing.T) {

	// Create empty BuildConfig object ...
	input := new(BuildConfig)
	input.Disk = new(DiskConfig)
	input.NTP = new(NTPConfig)
	input.Network = new(NetworkConfig)
	input.Redirects = new(RedirectConfig)
	input.Network.NetworkCards = make([]NetworkCardConfig, 4)
	input.Redirects.Rules = make([]Redirect, 4)

	// Create new temp file ...
	f, err := ioutil.TempFile("/tmp/", "vcfg-health-check")
	if err != nil {
		t.Error("error when executing ioutil.TempFile()")
	}

	// Marshal to tmp directory ...
	out, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		t.Error("error when executing json.MarshalIndent()")
	}

	err = ioutil.WriteFile(f.Name(), out, 0666)
	if err != nil {
		t.Error("error when executing ioutil.WriteFile()")
	}

	// Pass to HealthChecker ...
	err = VCFGHealthCheck("", f.Name(), true)
	if err != nil {
		t.Error("error when executing shared.VCFGHealthCheck()")
	}

}
