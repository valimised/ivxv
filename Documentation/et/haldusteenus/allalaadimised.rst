..  IVXV kogumisteenuse haldusliidese kasutusjuhend

.. _allalaadimised:

Väljavõtete allalaadimised
==========================

Väljavõtete allalaadimise leht avaneb menüüvalikust ``Allalaadimised``.


Hääletamise detailstatistika allalaadimine
------------------------------------------

Detailstatistika on JSON-vormingus ja see genereeritakse hääletamisteenuses.


Hääletamise seansside väljavõtte allalaadimine
----------------------------------------------

Hääletamise ja hääle kontrollimise seansside väljavõte on CSV-vormingus ja see
koostatakse logiseire teenuses.


E-valimiskasti allalaadimine
----------------------------

E-valimiskasti allalaadimise vormil on võimalik:

* Algatada e-valimiskasti koostamist;

* Koostatud e-valimiskaste alla laadida.

E-valimiskasti on võimalik koostada konsolideerimata ja konsolideeritud kujul.
Konsolideerimine on ressursinõudlik tegevus ja seda on vaja kasutada vaid
juhul, kui häälte kogumisel on esinenud probleeme talletusteenusega.

Enne hääletamisperioodi lõppu allalaaditav e-valimiskast on mõeldud varundamiseks,
pärast hääletusperioodi lõppu allalaaditud e-valimiskast on mõeldud häälte
kokkulugemiseks.

.. attention::

   E-valimiskasti koostamine võib olla aeganõudev tegevus! See sõltub kogutud häälte
   arvust, varundatud e-valimiskastide arvust, teenusmasinate riistvaralisest
   võimekusest, teenusmasinaid ühendava võrgu läbilaskevõimest ja süsteemi
   koormusest.


Töötlemisrakenduse sisendi aluse allalaadimine
----------------------------------------------

Töötlemisrakenduse sisendi alus on häälte töötlemiseks vajalike sisendfailide
komplekt ZIP-vormingus konteinerina ja see koostatakse haldusteenuses.
