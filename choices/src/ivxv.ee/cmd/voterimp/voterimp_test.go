package main

import "testing"

const (
	// Tabs at the end of voter list entries must be preserved!
	testInitial = `1
test election
algne
48908209998	CØNTROLINA ÅLT-DELETÈ	lisamine	100	1	100	1		
`
	testChanges = `1
test election
muudatused
48908209998	CØNTROLINA ÅLT-DELETÈ	kustutamine	100	1	100	1		
48908209998	CØNTROLINA ÅLT-DELETÈ	lisamine	200	2	200	2		
`
)

func TestListVersion(t *testing.T) {
	type value struct {
		name    string
		data    string
		version string
	}
	tests := []struct {
		name   string
		values []value
	}{
		{"simple", []value{
			{"foo", "foo", "2rpzjSZKiuwsKm1EAjBlfBlASUYdS3Pmh57mxM2slDU="},
			{"bar", "bar", "W6Phkw+VYIRm3EcL3IrDvgWLtQ2D10IZ94prP8kAZkE="},
			{"baz", "baz", "eV25KM6JmfU6mNL6y5AmWWf7NaMYfmQApBZ+A9yYT8s="},
			{"quux", "quux", "zvwtNgLcdOoOondCfEeC71iGGjkZeABolKVwGFGYdMM="},
		}},
		{"actual", []value{
			{"initial", testInitial, "wSG2fM3kAtVaeSqTgAfLi99NaeIQMkdVPm0hN7LW4xM="},
			{"changes", testChanges, "FFbhxo1vNBSM09dI34uzlNEKqGc4ciXLToasH+LiFTY="},
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var ver string
			for _, val := range test.values {
				ver = listver(ver, []byte(val.data))
				if ver != val.version {
					t.Errorf("unexpected %q version: got %q, want %q",
						val.name, ver, val.version)
				}
			}
		})
	}
}
