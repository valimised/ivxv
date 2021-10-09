#!/usr/bin/env sage

from sage.all import *

n = 3072
p_found = False
c = 0

while not(p_found):
    if (Mod(c,10000) == 0):
        print("c is: ", c)
    p = 2**n - 2**(n-64) - 1 + 2**64*(floor(2**(n-130)*pi.n(prec=10000))+c)
    if is_pseudoprime(p):
        print(p, " is prime")
        q = (p-1)/2
        if is_pseudoprime(q):
            print(p, " is safe prime")
            p_found = True
    c = c + 1
