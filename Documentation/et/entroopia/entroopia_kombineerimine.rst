Entroopia kombineerimine
========================

Kirjeldus
---------

Kasutades erinevaid entroopia-allikaid, on vajalik nende väljund kombineerida ja
normaliseerida. Kombineerimisel on oluline, et säiliks sisendite sõltumatus, st.
kombineeritud väljund ei tohi olla kehvem ühestki sisendist.

IVXV raamistikus toimub entroopia kombineerimine SHAKE-256 muutuva
väljundipikkusega räsifunktsiooni abil (XOF), kasutades skeemi nagu on
kirjeldatud [BDPA10]_.

Täpsemalt, pärast kombineerija isendi tekitamist saab lisada erinevaid entroopia
allikaid. Lõpliku pikkusega allika lisamisel loetakse kogu väärtus ning antakse
see SHAKE-256 sisendiks. Piiramata pikkusega allika lisamisel salvestatakse
selle viide kombineerija mällu.

Pärides kombineerijast töödeldud juhuslikkust, loetakse kõigepealt igast
salvestatud entroopia allikast sama palju baite ning antakse see SHAKE-256
sisendiks. Seejärel kopeeritakse SHAKE-256 isend, muudetakse kopeeritud
SHAKE-256 režiim lugemisele ning loetakse nõutud baitide jagu väljundit.

Kopeerimata SHAKE-256 sisendile antakse sisendiks sama täidis, mis
režiimimuutuse korral. Seejärel kirjutatakse kopeerimata SHAKE-256 isendile sama
arv nulle täiteks, kui loeti kopeeritud isendist väljundit.

Kopeeritud SHAKE-256 isend kustutatakse mälust ning edaspidi kasutatakse
kopeerimata SHAKE-256 sisendit, mis toimib endiselt kirjutamisrežiimis (me
kirjutasime ainult sama täidise, kuid ei muutnud reaalselt olekut).

Täpsema kirjelduseks viitame kombineerija lähtekoodile. Kombineerija on
defineeritud klassis `ee.ivxv.common.crypto.rnd.CombineRnd`.

Viited
------

.. [BDPA10] Guido Bertoni, Joan Daemen, Michaël Peeters, Gilles Van Assche:
   Sponge-Based Pseudo-Random Number Generators. CHES 2010: 33-47
