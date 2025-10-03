# TIVI Core Go library

TIVI Core Go library is a library written in [Go](https://go.dev/). It is a library for providing
common functionality for successful e-voting, mainly targeting cryptographic operations.

**TIVI Core Go's key features are:**

- ElGamal/Lifted ElGamal encryption support
- Homomorphic ElGamal/Lifted ElGamal encryption support
- Abstract group based cryptography
- ModP and elliptic curve support for ElGamal encryption through abstract groups
- Adaptive X509 certificates and PKCS8 private keys parsing

## Getting started

### Prerequisites

TIVI Core Go requires [Go](https://go.dev/) version [1.23](https://go.dev/doc/devel/release#go1.23.0) or above.

### Getting TIVI Core Go

```sh
GOPRIVATE='tivi.io' go get tivi.io/core@feature/crypto-api-2
```
