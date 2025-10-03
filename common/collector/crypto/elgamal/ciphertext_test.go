package elgamal

import (
	"encoding/base64"
	"testing"

	"tivi.io/core/math/group"
	_ "tivi.io/core/math/group/all"
)

const (
	modP3072Group = "RFC3526ModPGroup3072"
	ecP384Group   = "NIST-P384"

	base64DERModP3072Ciphertext = "MIIDGjALBgkrBgEEAZdVAgEwggMJAoIBgQCDkyKlAYLbY7VkpU2ayZkDQi3AKfXNKdQaUDnn+/jUB6W3Bn1jE2ijXtunt8g5fKeiDxuP59AcoI3pSKIZOqFFYOTKI+kDXe3bPz9RDziBWXtI6TawYj1C07dPCivdazxc6qUBUO1teeUZsQQV37q6FZ4KLc3r13gtCzQPCAykd/UIRxPSmp1hWZ4ytbg15YLt56E6wOLbHI6mbfqRgteWcsVM+YEicqPhVXDyLlkNMBk6OYXbjALjieX4B4K1sCPvB01ZIizjdHLZPzTas6juZ8Ga1izrS/tVaIRQVgQY2NXcgAYopZaW6K9uHIOApvTy+7ykOfhfrKmBoMy+54eOq5BN9rkcCkRhu+7L1qigmPM5SR1xiwgVAoN9lrp1lhZdfQAPmUVmchfTCyvGX05ZlqP2AxYgBbtiprbXqkr0K2SdCtee2ta1HLzs/oDHylzjdbQmqbXBaHEbJ90eFpm585Nq2jaQLW9V9wGJ39vzcQSZe4i1fdK546EBfNNCT2oCggGAEjdBWB8wPj+0Ag/lIxLWTbdSQViDNUIln/7QxZxwBrxIP5sRoVRnScxHu4TMskj8kQ9kwnmSeiP7yTnbUsHdko/bJPFhhMLeEe3sXMhA9J0WdRDGeWV1gtORFOIkD/jB8LAO6TpI3r9yZ8F+9SAxnHLyZ9KxDgfMk9/S0qCNYMTo/uFBlQiNl1cDxL92yoVfTkT6ufFa6T7VHhXKHgBNrR/r6VKVmcmpLseQHwmqHInjKAZ+S/PqxLY4yRmHNJXbun7bjaBvXMotvcKWenZbe+G2K+E0Q55V19PfivIrb2KrKumpSuS7wMSgLyuDdgC1AwKL9qeAwXfaDeHQ4DMK1WW/sg4NXow6/3e+WkgRI3ncFIfHVIxpcUsYqai+8XPomiomhhTNVLDAmTswdjnnsPgmDqAjIWLUyFvQup1iIYP/y3CJs4U0Cnc2MZ4CQIsaZ4i49dsR7kxaCLU+sPbUhDRaXK+LCRlDKc+vVTGi9O+r+Ztf+etwbr3zyRSqVK6v"
	base64DERECp384Ciphertext   = "MIHWMAsGCSsGAQQBho0fATCBxgRhBI/cd6M0Dyjedj2fOF1IA6ikpa4cdwniqMU49KibgsOtffMvtglYJei8B7eHyjE7D8Wl9HFmeBoLgeUTre0Mfk4Lknd4v1c/GJu1bYZlAMvmg46ysrItQ1sXD/lOwPNeiARhBHGq8thwH1qCSWQae5McuJB5GM4MO4m4tz3YQ6qnuKomG3b+QTS9ToU9ZK3OQj1Jjf6RA08tmERk417DC+IC+hbH6JjCCrVs2pvTBx5H0OQ+2uZ2yEhMYKwyLsfRGzu3rA=="
)

func TestASN1CiphertextVerify(t *testing.T) {
	testSamples := []struct {
		groupName        string
		base64Ciphertext string
	}{
		{
			groupName:        modP3072Group,
			base64Ciphertext: base64DERModP3072Ciphertext,
		},
		{
			groupName:        ecP384Group,
			base64Ciphertext: base64DERECp384Ciphertext,
		},
	}

	for _, testSample := range testSamples {
		t.Run(testSample.groupName, func(t *testing.T) {
			g, err := group.Get(testSample.groupName)
			if err != nil {
				t.Fatal(err)
			}

			ciphertext, err := base64.StdEncoding.DecodeString(testSample.base64Ciphertext)
			if err != nil {
				t.Fatal(err)
			}

			_, err = UnmarshalCiphertext(g, ciphertext)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
