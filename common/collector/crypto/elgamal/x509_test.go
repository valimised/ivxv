package elgamal

import (
	"testing"

	"tivi.io/core/crypto"
	x_509 "tivi.io/core/crypto/x509"
)

const (
	modP3072IVXVPEMPublicKey = "-----BEGIN PUBLIC KEY-----\nMIIDMjCCAaEGCSsGAQQBl1UCATCCAZICggGBAP//////////yQ/aoiFowjTExmKLgNwc0SkCTgiK\nZ8x0Agu+pjsTmyJRSgh5jjQE3e+VGbPNOkMbMCsKbfJfFDdP4TVtbVHCReSFtXZiXn7G9ExC6aY3\n7WsL/1y29Aa37e44a/taiZ+lrp8kEXxLH+ZJKGZR7ORbPcIAfLihY78FmNpINhxV05ppFj+o/STP\nX4NlXSPco62WHGLzViCFUrue1SkHcJaWbWcMNU5KvJgE8XRsCMoYIXwykF5GLjbOO+OedywYDoYD\nmyeDouwHoo+1xV3wb0xSyd4ry/aVWBcYOZVJfOqVauUV0iYYmPoFEBVyjlqKqsQtrTMXDQRQejOo\nVSGr3xy6ZOz7hQRY2+8KiupxV10GDH2zlw+FpuHkx6v1rozbCTPXHoyU4EolYZ3O49ImGtLua/Ev\n+gbZighk2HYCcz7IamRSHysYF3sgDLvhF1d6YV1sdwmIwLrZRuII4k+gdOWrMUPbW/zg/RCOS4LR\nIKk60sr//////////wIBAhsISEEtU0VUVVADggGJADCCAYQCggGAeitj08u4VcaxSlj1vbnO8QKD\nZAmFy72WeziY8otUuEHI5GQZVXZOCew/9EI/cSYTGf1tqDlg3m1wTIRkX5VY/zGUWPhiXEdLqiuK\nOzDDHBlKWE1Sikq81qdMwtn+Jz4e83/Cnefj/q3C30GMhiWKRg3TTR0SmS7hnuboOEDXAd0NAHP2\n5/eYj4UKiFVbRz0bmw80x3jCtGVZts/bdq5SO1jPAh9822HnxWU7jADRKFggItodgAgQFaFQQAad\n49ae3G8tgk8pj+W9kc3aCZEdN2hUj4LRy7h8mWky2abi1HWPZcbPuldHnSyCYvXawHZyyhNn2pos\nvpo6fkorkmBfBuxzudr0K8bw95iaRBVBa5Yug7hY3BHr/shwmYYMci10u8IfnHJqHCUweJr0pP31\nFsO2OkHCPnwyFJ0vhbgky2O693owerdaucZm2bXYANN2T//sonU5JdFq2WYB9DK5e+6e0GeOJxhb\nlOJeUpgz3vFpM6WPB3HMaY3+m/YE1JJO\n-----END PUBLIC KEY-----"
	ecP384IVXVPEMPublicKey   = "-----BEGIN PUBLIC KEY-----\nMIGGMBwGCSsGAQQBho0fATAPGwVQLTM4NBsGUksyMDI0A2YAMGMEYQSJJIKZPcSGFK4FxmrLBPaY\nV8meE4KNJxDUKk7cY6aBwckU+FqdY8C+MmSp3+MY42nxufI92C2EOBsaDfloRw3NM9x7FdXwhrgf\nTOJR5J20rWN2jSJCV65GQA8uh+NnL7Y=\n-----END PUBLIC KEY-----"
)

func TestX509Unmarshaller(t *testing.T) {
	unmarshaller := x_509.NewUnmarshaller[crypto.EncryptionKey](NewX509Unmarshaller())
	publicKeys := []struct {
		name string
		key  []byte
	}{
		{
			name: "RFC3526ModPGroup3072",
			key:  []byte(modP3072IVXVPEMPublicKey),
		},
		{
			name: "NIST-P384",
			key:  []byte(ecP384IVXVPEMPublicKey),
		},
	}

	for _, publicKey := range publicKeys {
		params, pkey, err := unmarshaller.Unmarshal(publicKey.key)
		if err != nil {
			t.Fatal(err)
		}

		if pkey == nil {
			t.Fatal("Public key is nil")
		}

		if params.Group().Name() != publicKey.name {
			t.Fatalf("expected group name %s, but got %s", publicKey.name, params.Group().Name())
		}
	}
}
