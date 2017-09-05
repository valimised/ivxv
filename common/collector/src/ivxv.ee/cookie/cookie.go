/*
Package cookie creates and opens session cookies which are protected from
reading and modification by a shared secret.

Note! Session cookies are not inherently protected from replay attacks.
Additional steps must be taken if resubmitting or delaying a cookie can cause
problems.
*/
package cookie // import "ivxv.ee/cookie"

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"sync"
)

// Key is the type of the shared secret used to create and access cookies.
type Key []byte

// C is a cookie manager which can create or open cookies.
type C struct {
	aead  cipher.AEAD // Authenticated encryption state.
	nonce []byte      // Next nonce: random initial value, incremented after each use.
	lock  sync.Mutex  // Synchronizes access to the nonce.
}

// New creates a new cookie manager with the provided key. The key must be
// valid for the underlying cipher, currently AES.
func New(key Key) (c *C, err error) {
	c = new(C)

	// Set up authenticated encryption.
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, KeyError{Err: err}
	}
	if c.aead, err = cipher.NewGCM(block); err != nil {
		return nil, CipherError{Err: err}
	}

	// Initialize the nonce.
	c.nonce = make([]byte, c.aead.NonceSize())
	if _, err = rand.Read(c.nonce); err != nil {
		return nil, InitNonceError{Err: err}
	}
	return
}

// Create creates a new cookie which contains the data, but cannot be read or
// modified without knowing the shared secret passed to New.
func (c *C) Create(data []byte) (cookie []byte) {
	// Cookies are a concatenation of the fixed length nonce, which we will
	// copy, and the ciphertext, which aead.Seal will append.
	cookie = make([]byte, len(c.nonce), len(c.nonce)+len(data)+c.aead.Overhead())

	// Copy the current nonce and increment it for next use.
	c.lock.Lock()
	copy(cookie, c.nonce)

	i := c.aead.NonceSize() - 1
	c.nonce[i]++
	for i > 0 && c.nonce[i] == 0 { // If a byte wraps, then increment the next one.
		i--
		c.nonce[i]++
	}
	c.lock.Unlock()

	// Encrypt data, append the result to cookie, and return.
	return c.aead.Seal(cookie, cookie, data, nil)
}

// Open opens a cookie using the shared secret passed to New and returns the
// data within.
func (c *C) Open(cookie []byte) (data []byte, err error) {
	n := c.aead.NonceSize()
	if len(cookie) <= n+c.aead.Overhead() {
		return nil, ShortCookieError{Len: len(cookie), Err: err}
	}

	nonce := cookie[:n]
	cipher := cookie[n:]

	if data, err = c.aead.Open(data, nonce, cipher, nil); err != nil {
		return nil, OpenError{Err: err}
	}
	return
}
