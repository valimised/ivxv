Miksimistõendi ülevaade
=======================

Ilma juhuslikkust kasutamata oleks iga krüpteerimise algoritm deterministlik,
st. valija valiku krüpteerimisel oleks tulemuseks alati sama väljund. See
lubaks ründajal tuvastada esialgse valiku, kui ta krüpteeriks kõikvõimalikud
valikud ja võrdleks seda nimekirja nähtud krüpteeritud valikuga.  Seega,
krüpteerimine peab kasutama juhuslikkust.

Juhuslikkuse kasutamine teeb krüptogrammid unikaalseks - isegi kui kaks korda
on krüpteeritud sama valik, siis krüptogrammid on erinevad. See tähendab, et
kui ründaja on võimeline mingil ajahetkel seostama krüptogrammi ja selle
krüpteerinud isikut, siis on tal võimalik hiljem ainult krüptogrammi nähes
tuletada seda andnud isik. Kuna sobivate parameetrite kasutamisel on ElGamali
avaliku võtme krüptosüsteem pikaajaliselt turvaline, siis see ei ole otseselt
probleemiks.

Probleem tekib siis, kui on tarvis tõestada, et krüptogramm on korrektselt
dekrüpteeritud. Kui kasutatav valikute agregeerimise algoritm töötab avateksti
kujul sõnede peal (nagu see IVXVs on), siis iga krüptogrammi kohta tekib üks
dekrüpteeritud avatekst. Korrektse dekrüpteerimise tõestus peab nii
krüptogrammi kui dekrüpteeritud avateksti siduma. Seega on võimalik ründajal
tekitada seos isiku ja krüptogrammi, ning krüptogrammi ja vastava avateksti
vahel, st. ta saab teada, mis valiku isik tegi.

Et seda seost eemaldada, kasutatakse IVXVs krüptogrammide segamist (mixnet).
Mixnet teeb korraga kahte operatsiooni -- järjestab sisendkrüptogrammid ümber
(permuteerib) ja uuendab krüptogrammis olevad juhuslikkust (rerandomiseerib).
See tähendab, et krüptogrammid, mis lähevad mixneti sisse on väliselt täiesti
sõltumatud krüptogrammidest, mis tulevad mixnetist välja. Kuna välise
sõltumatuse tõttu võiks teoreetiliselt mixnet krüptogramme asendada, siis on
tarvis lisada miksimistõend, mis tõestab krüptograafiliselt, et operatsioonid
on tehtud korrektselt ja ühtegi täiendavat operatsiooni pole tehtud.
Kontrollides miksimistõendit, on võimalik garanteerida, et mixnet on töötanud
korrektselt.

Väikeste parameetritega kirjutatud näide
----------------------------------------

IVXV kasutab valikute krüpteerimiseks ElGamali avaliku võtme krüptosüsteemi.
ElGamali krüptosüsteemi korral on fikseeritud algebralise rühma parameetrid
koos generaatoriga :math:`g`. Salajane võti :math:`x` valitakse vahemikus
:math:`[0, q-1]` ühtlase jaotusega, kus :math:`q` on rühma multiplikatiivse
alamrühma järk. Salajasele võtmele avalik võti on defineeritud:

.. math::
   pk = (g, g^x) = (g, y).

Krüpteerimaks sõne kujul valikut `V`, tuleb see kõigepealt kodeerida rühma
elemendiks

.. math::
   m = encode(V)

ja seejärel arvutatakse krüptogramm kasutades ühekordset juhuarvu
:math:`0<=r<q`

.. math::
   c = (c_1, c_2) = (m y^r, g^r).

Sellisel juhul piisab rerandomiseerimiseks teise ühekordse juhuarvuga
:math:`t` arvutada

.. math::
   c' = c * (y^t, g^t) = (m y^{r + t}, g^{r + t}).

Paneme veel tähele, et eelmises võrrandis :math:`(y^t, g^t)` on krüpteering
elemendist :math:`1`.

Dekrüpteerimiseks arvutatakse

.. math::
   d = c_1 / c_2^x,

ja dekoreeritakse :math:`S = decode(d)`.

Juhul kui krüptogramm on korrektselt konstrueeritud ja dekrüpteerimine on
korrektselt läbi viidud, siis :math:`d=m`, kuna

.. math::
   d = c_1 / c_2^x = (m y^r) / (g^{rx}) = (m y^r) / (y^r) = m.

Näiteks, oletame et rühm on täisarvud mooduli :math:`p = 227` järgi. Sellisel
juhul genereerib generaator :math:`g = 4` alamrühma järguga :math:`q = 113`.
Valime suvalise salajase võtme :math:`x = 100` ja sellele vastav avalik võti on
:math:`pk = (g, g^x) = (4, 21)`.

Oletame, et on neli erinevat valikut ning nende kodeeringud alarühma on
järgnevad:

===== =========
valik kodeering
===== =========
orav     16
jänes    64
hunt     29
kits    116
===== =========

Järgnevate valikute ja ühekordsete juhuarvude korral on krüptogrammid
järgnevad:

===== ========= ======= =============
valik kodeering juhuarv  krüptogramm
===== ========= ======= =============
kits     116       71    (62, 205)
hunt     29        80    (161, 221)
kits     116       64    (7, 147)
kits     116       47    (139, 36)
orav     16        76    (26, 172)
hunt     29        86    (30, 212)
kits     116       88    (155, 175)
orav     16        85    (87, 212)
orav     16        32    (132, 104)
jänes    64        22    (113, 171)
===== ========= ======= =============


Olgu mixneti kasutatav permutatsioon :math:`\pi` defineeritud järgnevalt:

====== =======
indeks väärtus
====== =======
   1      5
   2      8
   3      3
   4      6
   5      7
   6      2
   7      9
   8     10
   9      4
  10      1
====== =======

Olgu mixneti rerandomiseerimiseks kasutatavad juhuarvud:

====== ================
indeks täiendav juhuarv
====== ================
   1         43
   2        107
   3          6
   4         86
   5         56
   6         48
   7         35
   8        112
   9         55
  10        101
====== ================

Sellisel juhul pärast ümberjärjestamist on krüptogrammid järgnevas järjekorras:

==========   ==========  =============
 esialgne    uus indeks  permuteeritud
==========   ==========  =============
(62, 205)       5         (113, 171)
(161, 221)      8         (30, 212)
(7, 147)        3         (7, 147)
(139, 36)       6         (132, 104)
(26, 172)       7         (62, 205)
(30, 212)       2         (139, 36)
(155, 175)      9         (26, 172)
(87, 212)      10         (161, 221)
(132, 104)      4         (155, 175)
(113, 171)      1         (87, 212)
==========   ==========  =============

Pärast rerandomiseerimist on krüptogrammid järgnevad:

==========  ================  ==================  =================
 esialgne   täiendav juhuarv  korrutatav väärtus  rerandomiseeritud
==========  ================  ==================  =================
(113, 171)         43              (10, 103)         (222, 134)
(30, 212)         107              (28, 159)         (159, 112)
(7, 147)            6              (73, 10)          (57, 108)
(132, 104)         86              (100, 167)        (34, 116)
(62, 205)          56              (207, 113)        (122, 11)
(139, 36)          48              (78, 144)         (173, 190)
(26, 172)          35              (188, 73)         (121, 71)
(161, 221)        113              (173, 57)         (159, 112)
(155, 175)         55              (172, 85)         (101, 120)
(87, 212)         101              (103, 84)         (108, 102)
==========  ================  ==================  =================

Kontrollime, kuidas permuteeritud ja rerandomiseeritud krüptogrammid
dekrüpteeruvad:

===========  ==============  ============
krüptogramm  dekrüpteeritud  dekodeeritud
===========  ==============  ============
(222, 134)        64            jänes
(159, 112)        29            hunt
(57, 108)         116           kits
(34, 116)         16            orav
(122, 11)         116           kits
(173, 190)        116           kits
(121, 71)         16            orav
(159, 112)        29            hunt
(101, 120)        116           kits
(108, 102)        16            orav
===========  ==============  ============
