..  IVXV protocols

========================
Voting Result Audit
========================

Shuffle Proof Verification
================================

An algorithm is used to verify the shuffle proof, as defined in the
`Verificatum verifier implementing manual
<https://www.verificatum.org/files/vmnv-3.0.3.pdf>`_.

We would like to point out that when the shuffle proof is verified, the
ciphertext is edited with data on the election, the district, the station and
the question identifier. For addition, the relevant field is encoded as a group
element, using the random factor 0 to ensure obfuscation. For example: if at
first the ciphertext is :math:`c_0 = (c_00, c_01)`, using the public key
:math:`pk = (g, y)`, then for Verificatum’s input the wide ciphertext :math:`C =
(c_{id}, c_d, c_s, c_q, c_0)`, is used, where:

* the election identifier pseudo-ciphertext is given as  :math:`c_{id} = (1,
  encode(id))`, where the function :math:`encode` encodes the string as an
  element in the relevant group and `id` is the election identifier string
* the electoral district identifier pseudo-ciphertext is given as  :math:`c_d =
  (1, encode(d))`, where `d` is the district identifier string
* the polling station identifier pseudo-ciphertext is given as :math:`c_s = (1,
  encode(s))`, where `s` is the station identifier string
* the question identifier pseudo-ciphertext is given as :math:`c_q = (1,
  encode(q))`, where `q` is the question identifier string

In this case, the public key corresponding to the wide ciphertext is defined as
:math:`((g,1), (g,1), (g,1), (g,1), (g,y))`.

Verification of the Decryption Proof
==================================================

Let us have ciphertext  :math:`c = (c_0, c_1)`, which is decrypted into the
value :math:`d` with the given public key :math:`pk` over the parameters
:math:`(p,g)` and with the decryption proof :math:`(a,b,s)`.

To check that the decryption was correct, the challenge
:math:`k=H("DECRYPTION"||pk||c||d||a||b)` is calculated, and where :math:`H` is the
SHA-256 hash function.  Then it is verified that :math:`c_0^s = a * (c_1/d)^k`
and :math:`g^s = b * y^k`.

Correct Conversion Check
===============================

To make sure that the conversion between the IVXV ballot box and the Verificatum
ciphertexts has been performed correctly, the conversion has to be repeated
independently. After an independent conversion, the outputs have to be compared.
Since conversion is a deterministic procedure, a repeat conversion will
guarantee the correctness of the action.
