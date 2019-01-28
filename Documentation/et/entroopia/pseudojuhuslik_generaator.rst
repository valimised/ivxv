Pseudojuhuslik generaator
=========================

Pseudojuhuslik generaator (PRNG) on mõeldud baidijadade genereerimiseks, kasutades
etteantud seemneväärtust (seed). PRNG on olekuline, st. pärast sobiva
baidijada genereerimist on võimalik selle olek salvestada ning kasutada seda
täiendavate baidijadade genereerimiseks.

PRNG on teostatud räsifunktsiooni SHA2-256 abil, mida tähistame `H`.
Raundiseemne jaoks kasutatakse loendurit, mida esitatakse 64-bitise märgita
tavajärjestuses täisarvuna. Generaatori väljund on järgnev:::

    H(1 || seeme) || H(2 || seeme) || ...

Kataloogis `referents_implementatsioon` on Python3 programmeerimiskeeles
kirjutatud referents implementatsioon.

IVXV raamistikus on implementeeritud pseudojuhuslik generaator
`ee.ivxv.common.crypto.rnd.DPRNG`.
