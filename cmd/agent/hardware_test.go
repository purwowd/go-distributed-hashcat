package main

import (
	"strings"
	"testing"
)

const sampleHashcatIOutput = `hashcat (v6.1.1) starting...

OpenCL Info:
============

OpenCL Platform ID #1
  Vendor..: Mesa/X.org
  Name....: rusticl
  Version.: OpenCL 3.0 

  Backend Device ID #1
    Type...........: GPU
    Vendor.ID......: 1
    Vendor.........: AMD
    Name...........: AMD Radeon 780M Graphics (radeonsi, phoenix, LLVM 20.1.2, DRM 3.64, 6.17.0-35-generic)
    Version........: OpenCL 3.0 
    Processor(s)...: 12
    Clock..........: 2799
    Memory.Total...: 4096 MB (limited to 2047 MB allocatable in one block)
    Memory.Free....: 4032 MB
    OpenCL.Version.: OpenCL C 1.2 
    Driver.Version.: 25.2.8-0ubuntu0.24.04.2

OpenCL Platform ID #2
  Vendor..: The pocl project
  Name....: Portable Computing Language
  Version.: OpenCL 3.0 PoCL 5.0+debian  Linux, None+Asserts, RELOC, SPIR, LLVM 16.0.6, SLEEF, DISTRO, POCL_DEBUG

  Backend Device ID #2
    Type...........: CPU
    Vendor.ID......: 1
    Vendor.........: AuthenticAMD
    Name...........: cpu-skylake-avx512-AMD Ryzen 9 8945HS w/ Radeon 780M Graphics
    Version........: OpenCL 3.0 PoCL HSTR: cpu-x86_64-pc-linux-gnu-skylake-avx512
    Processor(s)...: 16
    Clock..........: 5263
    Memory.Total...: 25790 MB (limited to 8192 MB allocatable in one block)
    Memory.Free....: 25726 MB
    OpenCL.Version.: OpenCL C 1.2 PoCL
    Driver.Version.: 5.0+debian
`

func TestParseHashcatDevices_multiPlatform(t *testing.T) {
	devices := parseHashcatDevices(sampleHashcatIOutput)
	if len(devices) != 2 {
		t.Fatalf("expected 2 devices, got %d", len(devices))
	}

	if devices[0].Type != "GPU" {
		t.Errorf("device 1 type: got %q, want GPU", devices[0].Type)
	}
	if devices[0].Name == "Portable Computing Language" {
		t.Errorf("device 1 name must not be platform name PoCL, got %q", devices[0].Name)
	}
	if !strings.Contains(devices[0].Name, "Radeon 780M") {
		t.Errorf("device 1 name: got %q, want AMD Radeon 780M", devices[0].Name)
	}

	if devices[1].Type != "CPU" {
		t.Errorf("device 2 type: got %q, want CPU", devices[1].Type)
	}
	if !strings.Contains(devices[1].Name, "Ryzen 9 8945HS") {
		t.Errorf("device 2 name: got %q, want Ryzen 9 8945HS", devices[1].Name)
	}
}

func TestSelectBestDeviceHardware_prefersGPU(t *testing.T) {
	devices := parseHashcatDevices(sampleHashcatIOutput)
	hw := selectBestDeviceHardware(devices)

	if hw.Type != "GPU" {
		t.Errorf("hardware type: got %q, want GPU", hw.Type)
	}
	if hw.Processor == "Portable Computing Language" {
		t.Errorf("processor must not be PoCL platform name, got %q", hw.Processor)
	}
	if !strings.Contains(hw.Processor, "Radeon") {
		t.Errorf("processor: got %q, want Radeon GPU name", hw.Processor)
	}
}
