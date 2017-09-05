..  IVXV protokollid

================================================================================
Elektroonilise hääle kontrollimine
================================================================================

Elektroonilist häält kontrollitakse töötlemisrakenduses, kogumisteenuses,
valijarakenduses ja kontrollrakenduses. Kõige põhjalikuma kontrolli läbib
elektrooniline hääl e-urni koosseisus töötlemisrakenduses, kus otsustatakse
konkreetse hääle lugemisele saatmine või mittesaatmine. Iga üksiku hääle kohta
läbitakse töötlemisrakendusega analoogsel tasemel kontroll valijarakenduses, kus
veendutakse, et kogumisteenus on hääle kvalifitseerinud selliselt, et
töötlemisrakenduses tehtavad kontrollid õnnestuvad. Valijarakendusega analoogsed
kontrollid viib läbi kontrollrakendus.

Kontrollid kogumisteenuses
--------------------------

Valijarakendus saadab kogumisteenusele allkirjastatud hääle koosseisus

#. krüpteeritud sedeli
#. valija allkirja krüpteeritud sedelil
#. valija allkirjastamissertifikaadi

Kogumisteenus viib läbi minimaalselt järgmised kontrollid:

#. hääle allkirjastaja on valijate nimekirjas,
#. allkirjastatud hääl on esitatud korrektses konteinervormingus,
#. digitaalallkiri krüpteeritud sedelil on korrektne,
#. hääle allkirjastaja sertifikaat on kehtiv hääle vastuvõtmise ajahetkel.

Hääle allkirjastaja sertifikaadi kehtivuse kontrolliks teeb kogumisteenus
päringu kehtivuskinnitusteenusele. Kogumisteenus verifitseerib
kehtivuskinnitusteenuse vastust sertifikaadi oleku kohta ning lisab selle
vastuse häält kvalifitseerivate elementide hulka.

Kogumisteenus registreerib hääle talletamise fakti välises
registreerimisteenuses allkirjastades registreerimispäringu ning talletades
registreerimisteenuse poolt allkirjastatud registreerimistõendi häält
kvalifitseerivate elementide hulka.

Kogumisteenus tagastab kõik tema poolt hangitud häält kvalifitseerivad elemendid
valijarakendusele koos hääle unikaalse identifikaatoriga.

Kontrollid valijarakenduses
---------------------------

Valijarakendus moodustab valija avakujul tahteavalduse põhjal krüpteeritud
sedeli ning allkirjastab selle valija allkirja andmise vahendiga.

Valijarakenduse rolliks peale hääle allkirjastamist on veenduda, et
kogumisteenus käitus häält kvaliftseerivate elementide võtmisel
protokollikohaselt ning et hääl on talletatud selliselt, et ta saab
töötlemisrakenduse poolt arvesse võetud.

Valijarakendus viib läbi minimaalselt järgmised kontrollid:

#. Kogumisteenus võttis kehtivuskinnituse valija sertifikaadile volitatud
   kehtivuskinnitusteenuselt. Valijarakendus kontrollib allkirja
   kehtivuskinnitusteenuse vastusel.

#. Kogumisteenus registreeris valija poolt allkirjastatud hääle volitatud
   registreerimisteenuses. Valijarakendus kontrollib, et kogumisteenuse poolt
   moodustatud päring oli kogumisteenuse poolt signeeritud ning viitas
   korrektselt allkirjastatud häälele. Valijarakendus kontrollib, et
   registreerimisteenuse vastus on allkirjastatud õige registreerimisteenuse
   osutaja poolt ning selles sisaldub kogumisteenuse poolt allkirjastatud
   päring.

Kui hääle kvalifitseerimiseks vajalike elementide kontroll ei õnnestu, siis
teavitab valijarakendus sellest kasutajat.


Kontrollid kontrollrakenduses
-----------------------------

Kontrollrakendus saab valijarakendusest järgmise info:

#. Krüpteeritud sedeli moodustamisel kasutatud juhuslikkuse
#. Allkirjastatud hääle unikaalse identifikaatori kogumisteenuses

Kontrollrakendus kasutab hääle unikaalset identifikaatorit kogumisteenusest
järgmise info saamiseks:

#. krüpteeritud sedel,
#. valija allkiri krüpteeritud sedelil,
#. valija allkirjastamissertifikaat,
#. häält kvalifitseerivad elemendid, kaasaarvatud kehtivuskinnitus ja
   registreerimistõend

Kontrollrakendus teostab järgmised kontrollid:

#. allkirjastatud hääl on esitatud korrektses konteinervormingus,
#. digitaalallkiri krüpteeritud sedelil on korrektne,
#. hääle allkirjastaja sertifikaat on kehtiv hääle vastuvõtmise ajahetkel, mida
   kinnitab korrektne kehtivuskinnitus,
#. hääl on korrektselt registreeritud õiges registreerimisteenuses.

Nende kontrollide teostamise järel kuvab kontrollrakendus hääle allkirjastanud
isiku andmeid.

Täiendavalt kasutab kontrollrakendus krüpteeritud sedeli moodustamisel kasutatud
juhuslikkust krüpteeritud sedeli dekrüpteerimiseks. NB! Ühe hääle krüpteerimisel
kasutatud juhuslikkust saab kasutada ainult selle hääle dekrüpteerimiseks. Mitme
erineva hääle dekrüpteerimiseks läheb vaja häälte salastamise võtme
privaatkomponenti.

Kontrollrakendus veendub, et dekrüpteerimisel saadud avatekst vastab avakujul
tahteavalduse vorminõuetele.

Kontrollrakendus kuvab vorminõuetele vastava tahteavalduse võimaldamaks
kontrollijal selle tahteavalduse korrektsuses veenduda.


Kontrollid töötlemisrakenduses
------------------------------

Töötlemisrakendus kontrollib iga üksikut häält eraldi, veendudes muuhulgas, et
iga kogumisteenuse ning registreerimisteenuse poolt esitatud vaated e-urni sisu
kohta on konsistentsed. Seejärel otsustab töötlemisrakendus iga valija häälte
kohta, milline neist on ajaliselt viimane ning suunatakse töötlemise järgmisesse
etappi, mille tulemusena hääl võib jõuda lugemisele.

Töötlemisrakenduse sisendiks on:

#. Loend registreerimisteenuse poolt vastuvõetud registreerimispäringutest
#. Loend kogumisteenuses rakendatud valijanimekirjadest
#. Kogumisteenuse poolt üle antud e-urn, mis sisaldab iga hääle kohta
   krüpteeritud sedelit, valija allkirja krüpteeritud sedelil, valija
   allkirjastamissertifikaati, sertifikaadi kehtivuskinnitust ning
   registreerimistõendit.

Töötlemisrakendus kontrollib registreerimisteenuse ja kogumisteenuse kooskõla
ning väljastab erinevused:

#. Hääled, mille kohta on olemas registreerimispäring kogumisteenuses, kuid
   vastus ei ole jõudnud kogumisteenusesse
#. Hääled, mille kohta on olemas registreerimispäring registreerimisteenuses,
   kuid mida kogumisteenus ei ole üle andnud

Töötlemisrakendus kontrollib iga üksikut häält:

#. hääle allkirjastaja oli valijate nimekirjas,
#. allkirjastatud hääl on esitatud korrektses konteinervormingus,
#. digitaalallkiri krüpteeritud sedelil on korrektne,
#. hääle allkirjastaja sertifikaat on kehtiv hääle vastuvõtmise ajahetkel, mida
   kinnitab korrektne kehtivuskinnitus,
#. hääl on korrektselt registreeritud õiges registreerimisteenuses.

Töötlemisrakendus otsustab, milline valija häältest oli viimane ning liigub
töötlemise järgmisesse etappi. S.t. üks häält kvalifitseerivatest elementidest
täidab hääle talletamise aja fikseerimise rolli ning selle elemendi põhjal
moodustatakse üksikute häälte ajaline järgnevus. Olenevalt IVXV profiilist võib
see element olla kehtivuskinnituse koosseisus (BDOC-TM), eraldi ajatemplina
(BDOC-TS) või registreerimistõendi koosseisus (BDOC-TS).

.. vim: sts=3 sw=3 et:
