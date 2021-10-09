..  IVXV protokollid

========================
Hääletamistulemuse audit
========================

Miksimistõendi kontroll
=======================

Miksimistõendi kontrollimiseks kasutatakse algoritmi nagu on defineeritud
`Verificatumi verifitseerija implementeerimise manuaalis
<https://www.verificatum.org/files/vmnv-3.0.3.pdf>`_.

Märgime, et miksimistõendi koostamisel lisatakse krüptogrammile andmed valimiste,
ringkonna, jaoskonna ja küsimuse identifikaatori kohta.  Lisamiseks kodeeritakse
vastav väli rühma elemendina, kasutades pimendamiseks juhuslikkust 0. Näitena,
kui esialgu on krüptogramm :math:`c_0 = (c_{00}, c_{01})`, kasutades avalikku
võtit :math:`pk = (g, y)`, siis Verificatumi sisendina kasutatakse laia
krüptogrammi :math:`C = (c_{id}, c_d, c_s, c_q, c_0)`, kus:

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

Korrektse dekrüpteerimise kontrollimise jaoks on tarvis arvutada
mitte-interaktiivne kontrollija väljakutse. Selle jaoks kodeeritakse
:math:`"DECRYPTION" || pk || c || d || a || b` DER-kodeeringus. Baidijada
kasutatakse deterministliku juhuarvugeneraatori initsialiseerimiseks ja selle
väljundist loetakse rühma järgu pikkune täisarv :math:`k`.

Korrektse dekrüpteerimsie tõendi kontrolliks tuleb kontrollida, et :math:`c_0^s
= a * (c_1/d)^k` ja :math:`g^s = b * y^k`.

Korrektse teisendamise kontroll
===============================

Kontrollimaks, et teisendus IVXV e-valimiskasti ja Verificatumi krüptogrammide vahel on
tehtud korrektselt, tuleb korrata teisendust sõltumatult. Pärast sõltumatut
teisendust tuleb võrrelda saadud väljundeid. Kuna teisendamine on deterministlik
protseduur, siis garanteerib kordamine tegevuse õigsuse.
