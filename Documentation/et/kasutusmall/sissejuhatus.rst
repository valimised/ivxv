..  IVXV kasutusmallid

Sissejuhatus
============

Dokument kirjeldab elektroonilise hääletamise süsteemi kasutusmallimudeli. Dokument on jaotatud kaheks osaks: tegijad ja kasutusmallid. Esimeses osas on kirjeldatud kasutusmallimudeli tegijad. Teises osas on kirjeldatud elektroonilise hääletamise süsteemi kasutusmallid.

Viited
------
1.  [UML] – UMLi kontsentraat. Objektmodelleerimise standardkeele UML 2.0 lühijuhend. Martin Fowler. Cybernetica AS.

Kasutatud metoodika
-------------------

Kasutusmallimudel
`````````````````
    Kasutusmallid on abivahend, mis aitab aru saada süsteemi talitlusnõuetest. Kasutusmallid esitavad süsteemi välisvaate kirjeldades süsteemi kasutajate ja süsteemi vahelisi interaktsioone kasutajate eesmärkide täitmiseks. Kasutusmallimudel koondab endas kõik analüüsitava süsteemi olulised kasutusmallid [UML].

Tegija
``````
    Tegija on roll, mida kasutaja mängib süsteemi suhtes. Tegija ei pea olema inimene. Kui modelleeritav süsteem annab mingit teenust teisele arvutisüsteemile, on see teine süsteem tegija.

Kasutusmall
```````````

Kirjeldus
'''''''''

Kasutusmall on kogum stsenaariume, mida omavahel ühendab ühine kasutaja eesmärk. Iga stsenaarium on süsteemis tehtavate toimingute jada, mis annab tegijale nähtava ja kasuliku tulemuse.

Eeltingimus
'''''''''''

Eeltingimus kirjeldab tingimusi, mille täidetuse peab süsteem tagama, enne kui ta laseb alustada kasutusmalli täitmist.

Päästik
'''''''

Päästik spetsifitseerib sündmuse, mis käivitab kasutusmalli.

Põhiprotsess
''''''''''''

Põhiprotsess kirjeldab kasutusmalli peastsenaariumi.

Laiendid
''''''''

Laiend kasutusmallis nimetab oleku, mis tuleneb teistsugustest interaktsioonidest kui need, mis on kirjeldatud õnnestumise peastsenaariumis, ja ütleb, millised on need erinevused.

Järeltingimus
'''''''''''''

Järeltingimuse täidetus garanteeritakse süsteemi poolt kasutusmalli täitmise lõpus.

E-hääletamise etapid
--------------------

E-hääletamine jaguneb korralduslikult viieks etapiks:

- hääletamiseelne etapp
- hääletamisetapp
- töötlusetapp
- lugemisetapp
- auditeerimisetapp

Mõisted
-------

Mõiste *isik* võib tähistada nii füüsilist kui juriidilist isikut. Eeldame, et konkreetsete toimingute läbiviija on alati üheselt indentifitseeritav füüsiline isik, kes võib tegutseda juriidilise isiku volitatud esindajana.

Mõiste *serverisüsteem* tähistab terviklikku kogumit tarkvaralisi ja riistvaralisi komponente, mis koostoimes realiseerivad konkreetset protokolli ning osutavad pikema aja vältel teenust paljudele kasutajatele.

Mõiste *liides* tähistab selgelt spetsifitseeritud kokkupuutepunkti süsteemi komponentide vahel, mis võimaldab komponentide vahelist infovahetust.

Mõiste *rakendus* tähistab tarkvaralist komponenti, mida käivitatakse konkreetsel ajahetkel ühel riistvaralisel komponendil ühe kasutaja poolt. Rakendus võib oma ülesande täitmiseks suhelda servrisüsteemiga.

Mõiste *teenus* tähistab süsteemivälist komponenti, millega süsteem oma ülesannete täitmiseks konkreetsete liideste kaudu andmeid vahetab.



.. vim: sts=3 sw=3 et:
