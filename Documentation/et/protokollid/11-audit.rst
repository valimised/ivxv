..  IVXV protokollid

========================
Hääletamistulemuse audit
========================

Miksimistõendi kontroll
=======================

Miksimistõendi kontrollimiseks kasutatakse algoritmi nagu on defineeritud
`Verificatumi verifitseerija implementeerimise manuaalis
<http://verificatum.com/files/vmnv-3.0.2.pdf>`_.

Märgime, et miksimistõendi koostamisel lisatakse krüptogrammile andmed valimiste,
ringkonna, jaoskonna ja küsimuse identifikaatori kohta.  Lisamiseks kodeeritakse
vastav väli rühma elemendina, kasutades pimendamiseks juhuslikkust 0. Näitena,
kui esialgu on krüptogramm :math:`c_0 = (c_00, c_01)`, kasutades avalikku võtit
:math:`pk = (g, y)`, siis Verificatumi sisendina kasutatakse laia krüptogrammi
:math:`C = (c_{id}, c_d, c_s, c_q, c_0)`, kus:

* valimiste identifikaatori pseudokrüptogramm on antud kujul :math:`c_{id} = (1,
  encode(id))`, kus funktsioon :math:`encode` kodeerib sõne vastava rühma
  elemendina ja `id` on valimiste identifikaatori sõne.
* ringkonna identifikaatori pseudokrüptogramm on antud kujul :math:`c_d = (1,
  encode(d))`, kus `d` on ringkonna identifikaatori sõne.
* jaoskonna identifikaatori pseudokrüptogramm on antud kujul :math:`c_s = (1,
  encode(s))`, kus `s` on jaoskonna identifikaatori sõne.
* küsimuse identifikaatori pseudokrüptogramm on antud kujul :math:`c_q = (1,
  encode(q))`, kus `q` on küsimuse identifikaatori sõne.

Sellisel juhul defineeritakse laia krüptogrammile vastava avaliku võtmena
:math:`((g,1), (g,1), (g,1), (g,1), (g,y))`.

Korrektse dekrüpteerimise tõendi kontroll
=========================================

Olgu antud krüptogramm :math:`c = (c_0, c_1)`, mis deküpteeritakse väärtuseks
:math:`d` antud avaliku võtmega :math:`pk` üle parameetrite :math:`(p,g)` ja
dekrüpteerimistõendiga :math:`(a,b,s)`.

Korrektse dekrüpteerimise kontrollimise jaoks arvutatakse väljakutse
:math:`k=H("DECRYPTION"||pk||c||d||a||b)`, kus :math:`H` on SHA-256
räsifunktsioon. Seejärel kontrollitakse, et :math:`c_0^s = a * (c_1/d)^k` ja
:math:`g^s = b * y^k`.

Korrektse teisendamise kontroll
===============================

Kontrollimaks, et teisendus IVXV e-urni ja Verificatumi krüptogrammide vahel on
tehtud korrektselt, tuleb korrata teisendust sõltumatult. Pärast sõltumatut
teisendust tuleb võrrelda saadud väljundeid. Kuna teisendamine on deterministlik
protseduur, siis garanteerib kordamine tegevuse õigsuse.

.. vim: sts=3 sw=3 et:
