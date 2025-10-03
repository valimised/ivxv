..  IVXV protokollid

================================================================================
Elektroonilise hääle kontrollimine eraldi ja e-valimiskasti koosseisus
================================================================================

Elektroonilist häält kontrollitakse töötlemisrakenduses, kogumisteenuses,
valijarakenduses ja kontrollrakenduses.

Kõige põhjalikuma kontrolli läbib elektrooniline hääl e-valimiskasti koosseisus
töötlemisrakenduses, kus otsustatakse konkreetse hääle lugemisele saatmine või
mittesaatmine. Nende kontrollide käigus vaadeldakse e-häält nii eraldi kui ka
suhtes kõigi teiste sama valija poolt antud häältega. Täiendavalt kõrvutatakse
e-valimiskasti registreerimisteenuse väljavõttega.

Iga üksiku hääle kohta läbitakse töötlemisrakendusega analoogsel tasemel
kontroll valijarakenduses, kus veendutakse, et kogumisteenus on hääle
kvalifitseerinud selliselt, et töötlemisrakenduses tehtavad kontrollid
õnnestuvad. Valijarakenduse kontrollidega samaväärsed kontrollid viib läbi
kontrollrakendus.

Kogumisteenus on lisaks vastutav mitme hääle lõplikuks kvalifitseerumiseks
vajaliku elemendi hankimise eest ning teostab ka nende hankimise järgselt kontrollid.


Lõpliku elektroonilise hääle komponendid
----------------------------------------

Elemendid, mis on kättesaadavad otsustamise hetkel, kas hääl kvalifitseerub
lugemisele saatmiseks või mitte:

   #. elektroonilist häält sisaldav konteiner - :ref:`entity-haale-konteiner`;

      #. krüpteeritud sedel - :ref:`entity-krypteeritud-sedel`;

      #. valija signatuur krüpteeritud sedelil - :ref:`entity-haale-signatuur`;

      #. valija allkiri krüpteeritud sedelil - :ref:`entity-haale-allkiri`;

      #. valija allkirjastamissertifikaat - :ref:`entity-valija-sertifikaat`;

      #. valija isikukood - :ref:`entity-valija-identiteet`;

   #. kvalifitseeriv element - valija sertifikaadi kehtivuskinnitus -
      :ref:`entity-kehtivuskinnitus`;

   #. kvalifitseeriv element - ajatempel allkirjastatud krüpteeritud sedelile -
      :ref:`entity-ajatempel`;

   #. kvalifitseeriv element - registreerimispäringu konteiner
      :ref:`entity-registreerimisparing-konteiner`;

      #. kvalifitseeriv element - registreerimispäring allkirjastatud
         krüpteeritud sedelile - :ref:`entity-registreerimisparing`;

   #. kvalifitseeriv element - registreerimistõend allkirjastatud krüpteeritud
      sedeli kohta - :ref:`entity-registreerimistoend`;

   #. valija ringkonnakuuluvuse tõend hääle andmise hetkel -
      :ref:`entity-nimekirjatunnus`.


Elemendid, mis on kättesaadavad otsustamise hetkel, kuidas häält kokkulugemisel
arvesse võtta:

   #. miksitud krüpteeritud sedel - :ref:`entity-miksitud-krypteeritud-sedel`;

   #. valija tahteavaldus avakujul - :ref:`entity-tahteavaldus`;

   #. ringkonnakuuluvuse tunnus - :ref:`entity-ringkonnatunnus`;

   #. ringkonnapõhine valikute nimekiri - :ref:`entity-ringkonna-valikutenimekiri`.


Täiendavad elemendid:

   #. krüpteerimisel kasutatud juhuslikkus - :ref:`entity-juhuslikkus`; -
      luuakse valijarakenduses hääle krüpteerimisel ning vahendatakse ainult
      kontrollrakendusele.


.. _entity-tahteavaldus:

````````````
TAHTEAVALDUS
````````````

Olem moodustatakse valijarakenduses ning tuvastatakse võtmerakenduses. Tegemist
on EHS spetsiifilises vormingus UTF-8 kodeeringus baidijadaga.

.. _check-tahteavaldus-correctness:

TAHTEAVALDUS, vormingu korrektsus
`````````````````````````````````

Tahteavalduse vormingu korrektsuse kontroll käesoleva spetsifikatsiooni suhtes.


.. _check-tahteavaldus-ringkonnatunnus-valikutenimekiri-consistency:

TAHTEAVALDUS, RINGKONNATUNNUS, RINGKONNA VALIKUTENIMEKIRI kooskõlalisus
```````````````````````````````````````````````````````````````````````

Tahteavalduses sisalduva valiku kontroll ringkonnatunnuse ja valikute nimekirja
suhtes. Kontroll õnnestub, kui tuvastatud valik on kättesaadav antud
ringkonnas.

.. _check-tahteavaldus-ringkonna-valikutenimekiri-consistency:

TAHTEAVALDUS, RINGKONNA VALIKUTENIMEKIRI, kooskõlalisus
```````````````````````````````````````````````````````

Tahteavalduses sisalduva valiku kontroll ringkonnapõhise valikute nimekirja
suhtes. Kontroll õnnestub, kui tuvastatud valik on kättesaadav antud ringkonnas.


.. _entity-juhuslikkus:

```````````
JUHUSLIKKUS
```````````

Olem moodustatakse valijarakenduses ning kasutatakse ka kontrollrakenduses.
Olem esitatakse EHS spetsiifilises vormingus baidijadana ning peab olema
ühilduv avaliku võtme parameetrite poolt määratud matemaatilise rühmaga.

.. _check-juhuslikkus-correctness:

JUHUSLIKKUS, vormingu korrektsus
````````````````````````````````

Juhuslikkuse vormingu korrektsuse kontroll käesoleva spetsifikatsiooni suhtes.

.. _check-juhuslikkus-public-key-consistency:

JUHUSLIKKUS, kooskõlalisus avaliku võtmega
``````````````````````````````````````````

Juhuslikkuse kooskõlalisuse kontroll avaliku võtmega. Kontroll õnnestub kui
juhuslikkus on kasutatav skalaarina arvutusteks avaliku võtme parameetrite
poolt määratud matemaatilises rühmas.

.. _entity-krypteeritud-sedel:

``````````````````
KRÜPTEERITUD SEDEL
``````````````````

Olem moodustatakse valijarakenduses. Tegemist on EHS spetsiifilises vormingus
DER-kodeeritud andmestruktuuriga, mille kontrollimise aluseks on EHS avalik
võti, mis muuhulgas määratleb krüpteerimisel kasutatud matemaatilise rühma.

.. _check-krypteeritud-sedel-correctness:

KRÜPTEERITUD SEDEL, vormingu korrektsus
```````````````````````````````````````

Krüpteeritud sedeli vormingu korrektsuse kontroll käesoleva spetsifikatsiooni suhtes.


.. _check-krypteeritud-sedel-public-key-consistency:

KRÜPTEERITUD SEDEL, kooskõlalisus avaliku võtmega
`````````````````````````````````````````````````

Krüpteeritud sedeli kooskõlalisuse kontroll avaliku võtmega. Kontroll õnnestub kui
krüpteeritud sedeli komponendid on kasutatavad rühma liikmetena arvutusteks
avaliku võtme parameetrite poolt määratud matemaatilises rühmas.

.. _check-krypteeritud-sedel-juhuslikkus-consistency:

KRÜPTEERITUD SEDEL, JUHUSLIKKUS, kooskõlalisus
``````````````````````````````````````````````

Krüpteeritud sedeli kooskõlalisuse kontroll juhuslikkusega lähtudes rühma
parameetritest. Kontroll õnnestub kui õnnestub näidata, et antud juhuslikkust
on kasutatud krüpteeritud sedeli komponendi `uBlind` arvutamiseks antud rühma
parameetrite järgi.


.. _entity-valija-sertifikaat:

``````````````````
VALIJA SERTIFIKAAT
``````````````````

Olem on valijale omistatud süsteemiväliselt. Tegemist on X.509 vormingus
sertifikaadiga, mille kehtivuse kontrolli aluseks on samasse avaliku võtme
infrastruktuuri kuuluvad juursertifikaadid ja kehtivuskinnitusteenuse osutaja.

.. _check-valija-sertifikaat-correctness:

VALIJA SERTIFIKAAT, vormingu korrektsus
```````````````````````````````````````

Valija sertifikaadi vormingu korrektsuse kontroll lähtudes X.509 spetsifikatsioonist.

.. _check-valija-sertifikaat-consistency-protocol-settings:

VALIJA SERTIFIKAAT, protokollikohane kooskõlalisus valimise seadistustega
`````````````````````````````````````````````````````````````````````````

Valija sertifikaadi protokollikohase kooskõlalisuse kontroll valimise
seadistustega. Kontroll õnnestub kui:

   #. Valija sertifikaat on kehtiv lähtudes sertifikaadis sisalduvast kehtivusajast;

   #. Valija sertifikaat kuulub mõnda valimise seadistustes kirjeldatud
      sertifitseerimishierarhiatest.


.. _check-valija-sertifikaat-kehtivuskinnitus-consistency:

VALIJA SERTIFIKAAT, KEHTIVUSKINNITUS, protokollikohane kooskõlalisus valimise seadistustega
```````````````````````````````````````````````````````````````````````````````````````````

Valija sertifikaadi ja kehtivuskinnituse protokollikohase kooskõlalisuse
kontroll valimise seadistustega. Kontroll õnnestub kui:

   #. Kehtivuskinnitus on korrektselt allkirjastatud kehtivuskinnituse teenuse
      osutaja poolt, kellel on seadistuste kohaselt voli valija sertifikaadi
      kehtivust tõendada

   #. Kehtivuskinnitus on väljastatud valija sertifikaadi kohta;

   #. Kehtivuskinnitus näitab valija sertifikaadi OCSP olekuks 'Kehtiv'.


.. _entity-valija-identiteet:

`````````````````
VALIJA IDENTITEET
`````````````````

Olem on valijale omistatud süsteemiväliselt. Tegemist on isikukoodiga, mille
kontrollimise aluseks on VALIJA SERTIFIKAAT.

.. _check-valija-identiteet-correctness:

VALIJA IDENTITEET, vormingu korrektsus
``````````````````````````````````````

Valija identiteedi vormingu korrektsuse kontroll lähtudes Eesti Vabariigi
isikukoodi vormingust.

.. _check-valija-identiteet-nimekirjatunnus-eligibility:

VALIJA IDENTITEET, NIMEKIRJATUNNUS, hääleõigus
``````````````````````````````````````````````

Hääleõiguse kontroll valija identiteedi ja nimekirjatunnuse alusel. Kontroll
õnnestub, kui valija identiteet on kantud nimekirjatunnusele vastavasse
valijate nimekirja. Sellisel juhul on identiteedile omistatud ka
ringkonnatunnus, mis määrab ringkonnaspetsiifilise valikute nimekirja.

.. _entity-haale-signatuur:

```````````````
HÄÄLE SIGNATUUR
```````````````

Olem arvutatakse allkirjastamise vahendi poolt lähtudes valijarakenduse poolt
sisendiks antud KRÜPTEERITUD SEDELI räsist.

.. _check-haale-signatuur-correctness:

HÄÄLE SIGNATUUR, vormingu korrektsus
````````````````````````````````````

Hääle signatuuri vormingu korrektsuse kontroll lähtudes konkreetsest
signeerimismeetodist.

.. _check-haale-signatuur-krypteeritud-sedel-valija-sertifikaat-consistency:

HÄÄLE SIGNATUUR, KRÜPTEERITUD SEDEL, VALIJA SERTIFIKAAT, kooskõlalisus
``````````````````````````````````````````````````````````````````````

Kontroll, mis õnnestub siis ja ainult siis kui õnnestub verifitseerida, et
hääle signatuur on arvutatud krüpteeritud sedeli räsist kasutades valija
sertifikaadis leiduvale avalikule võtmele vastavat privaatvõtit.

.. _entity-haale-allkiri:

`````````````
HÄÄLE ALLKIRI
`````````````

Olem moodustatakse valijarakenduses, talletades hääle signatuuri allkirja vormingusse.

.. _check-haale-allkiri-correctness:

HÄÄLE ALLKIRI, vormingu korrektsus
``````````````````````````````````

Hääle allkirja vormingu korrektsuse kontroll käesoleva spetsifikatsiooni suhtes.

.. _check-haale-allkiri-haale-signatuur-consistency:

HÄÄLE ALLKIRI, HÄÄLE SIGNATUUR, kooskõlalisus
`````````````````````````````````````````````

Hääle allkirja ja hääle signatuuri kooskõlalisuse kontroll. Kontroll õnnestub,
kui antud allkiri sisaldab antud signatuuri.

.. _entity-haale-konteiner:

```````````````
HÄÄLE KONTEINER
```````````````

Olem moodustatakse valijarakenduses, luues vormingukohane konteiner, mis
sisaldab muuhulgas krüpteeritud sedelit, hääle allkirja ja valija sertifikaati.

.. _check-haale-konteiner-correctness:

HÄÄLE KONTEINER, vormingu korrektsus
````````````````````````````````````

Hääle konteineri vormingu korrektsuse kontroll käesoleva spetsifikatsiooni suhtes.

.. _check-haale-konteiner-haale-allkiri-valija-sertifikaat-krypteeritud-sedel-consistency:

HÄÄLE KONTEINER, HÄÄLE ALLKIRI, VALIJA SERTIFIKAAT, KRÜPTEERITUD SEDEL, kooskõlalisus
`````````````````````````````````````````````````````````````````````````````````````

Hääle konteineri, hääle allkirja, valija sertifikaadi ja krüpteeritud sedeli
kooskõlalisuse kontroll. Kontroll õnnestub kui antud konteiner sisaldab
konkreetset hääle allkirja, valija sertifikaati ja krüpteeritud sedelit.

.. _entity-kehtivuskinnitus:

````````````````
KEHTIVUSKINNITUS
````````````````

Olemi hankimise eest vastutab kogumisteenus. Tegemist on OCSP vormingus
kinnitusega valija sertifikaadi oleku kohta.

.. _check-kehtivuskinnitus-correctness:

KEHTIVUSKINNITUS, vormingu korrektsus
`````````````````````````````````````

Kehtivuskinnituse vormingu korrektsuse kontroll OCSP spetsifikatsiooni suhtes.


.. _check-kehtivuskinnitus-ajatempel-order:

KEHTIVUSKINNITUS, AJATEMPEL, ajaline järgnevus
``````````````````````````````````````````````

Kehtivuskinnituse ja ajatempli ajalise järgnevuse kontroll. Kontroll õnnestub, kui

   #. ajatempel ei ole väljastatud hiljem kui kehtivuskinnitus;

   #. ajatempli ja kehtivuskinnituse väljastamise ajaline vahe on väiksem kui
      valimise seadistustes ette nähtud aeg.


.. _entity-ajatempel:

`````````
AJATEMPEL
`````````

Olemi hankimise eest vastutab kogumisteenus. Tegemist on PKIX vormingus ajatempliga.

.. _check-ajatempel-correctness:

AJATEMPEL, vormingu korrektsus
``````````````````````````````

Ajatempli vormingu korrektsuse kontroll PKIX spetsifikatsiooni suhtes.


.. _check-ajatempel-consistency-protocol-settings:

AJATEMPEL, HÄÄLE ALLKIRI, protokollikohane kooskõlalisus valimise seadistustega
```````````````````````````````````````````````````````````````````````````````

Ajatempli protokollikohase kooskõlalisuse kontroll valimise seadistustega.
Kontroll õnnestub kui:

   #. ajatempel on korrektselt allkirjastatud valimise seadistuses volitatud
      ajatempliteenuse osutaja poolt;

   #. ajatempel on võetud hääle allkirjale.

.. _entity-registreerimisparing:

````````````````````
REGISTREERIMISPÄRING
````````````````````

Olem moodustatakse kogumisteenuses. Olemi talletamise eest vastutab
registreerimisteenuse osutaja.

.. _check-registreerimisparing-correctness:

REGISTREERIMISPÄRING, vormingu korrektsus
`````````````````````````````````````````

Registreerimispäringu vormingu korrektsuse kontroll käesoleva spetsifikatsiooni suhtes.


.. _check-registreerimisparing-haale-allkiri-consistency:

REGISTREERIMISPÄRING, HÄÄLE ALLKIRI, protokollikohane kooskõlalisus valimise seadistustega
``````````````````````````````````````````````````````````````````````````````````````````

Registreerimispäringu ja hääle allkirja protokollikohase kooskõlalisuse
kontroll valimise seadistustega. Kontroll õnnestub kui:

   #. Registreerimispäring on koostatud hääle allkirjale;

   #. Registreerimispäring on korrektselt allkirjastatud valimise seadistuses
      viidatud kogumisteenuse poolt.


.. _check-registreerimisparing-registreerimistoend-consistency:

REGISTREERIMISPÄRING, REGISTREERIMISTÕEND, kooskõlalisus
```````````````````````````````````````````````````````````````````````

Registreerimispäringu ja registreerimistõendi protokollikohase kooskõlalisuse
kontroll valimise seadistustega. Kontroll õnnestub kui:

   #. Registreerimispäring on korrektselt allkirjastatud valimise seadistuses
      viidatud kogumisteenuse poolt.

   #. Registreerimistõend on koostatud vastuseks samale registreerimispäringule;

   #. Registreerimistõend on korrektselt allkirjastatud valimise seadistuses
      viidatud registreerimisteenuse poolt.


.. _entity-registreerimisparing-konteiner:

```````````````````````````````
REGISTREERIMISPÄRINGU KONTEINER
```````````````````````````````

Olem moodustatakse kogumisteenuses. Olemi talletamise eest vastutab
registreerimisteenuse osutaja.

.. _check-registreerimisparing-konteiner-correctness:

REGISTREERIMISPÄRINGU KONTEINER, vormingu korrektsus
````````````````````````````````````````````````````

Registreerimispäringu konteineri vormingu korrektsuse kontroll käesoleva
spetsifikatsiooni suhtes.


.. _check-registreerimisparing-konteiner-registreerimispäring-consistency:

REGISTREERIMISPÄRINGU KONTEINER, REGISTREERIMISPÄRING, protokollikohane kooskõlalisus valimise seadistustega
````````````````````````````````````````````````````````````````````````````````````````````````````````````

Registreerimispäringu konteineri ja registreerimispäringu protokollikohase kooskõlalisuse
kontroll valimise seadistustega. Kontroll õnnestub kui:

   #. Registreerimispäringu konteiner sisaldab registreerimispäringut;

   #. Registreerimispäring on korrektselt allkirjastatud valimise seadistuses
      viidatud kogumisteenuse poolt.


.. _entity-registreerimistoend:

```````````````````
REGISTREERIMISTÕEND
```````````````````

Olemi hankimise eest vastutab kogumisteenus, olemi moodustab
registreerimisteenuse osutaja.

.. _check-registreerimistoend-correctness:

REGISTREERIMISTÕEND, vormingu korrektsus
````````````````````````````````````````

Registreerimistõendi vormingu korrektsuse kontroll käesoleva spetsifikatsiooni suhtes.


.. _check-registreerimistoend-haale-allkiri-consistency:

REGISTREERIMISTÕEND, HÄÄLE ALLKIRI, protokollikohane kooskõlalisus valimise seadistustega
`````````````````````````````````````````````````````````````````````````````````````````

Registreerimistõendi ja hääle allkirja protokollikohase kooskõlalisuse
kontroll valimise seadistustega. Kontroll õnnestub kui:

   #. Registreerimistõend on koostatud hääle allkirjale;

   #. Registreerimistõend sisaldab kogumisteenuse poolt korrektselt
      allkirjastatud registreerimispäringut;

   #. Registreerimistõend on korrektselt allkirjastatud valimise seadistuses
      viidatud registreerimisteenuse poolt.


.. _entity-nimekirjatunnus:

```````````````
NIMEKIRJATUNNUS
```````````````

Olem moodustatakse kogumisteenuses ning identifitseerib ühe konkreetse
versiooni valijate nimekirjast.

.. _check-nimekirjatunnus-correctness:

NIMEKIRJATUNNUS, vormingu korrektsus
````````````````````````````````````

Nimekirjatunnuse vormingu korrektsuse kontroll käesoleva spetsifikatsiooni suhtes.

.. _check-nimekirjatunnus-consistency-protocol-settings:

NIMEKIRJATUNNUS, protokollikohane kooskõlalisus valimise seadistusega
`````````````````````````````````````````````````````````````````````

Nimekirjatunnuse protokollikohase kooskõlalisuse kontroll valimise
seadistusega. Kontroll õnnestub kui valimise seadistuses sisalduvate
muudatusnimekirjade alusel on võimalik koostada nimekirjatunnusele vastav
nimekiri.


.. _entity-miksitud-krypteeritud-sedel:

```````````````````````````
MIKSITUD KRÜPTEERITUD SEDEL
```````````````````````````

Olem moodustatakse miksimisrakenduses. Miksitud krüpteeritud sedel on
sisuliselt krüpteeritud sedel ning rakenduvad samad kontrollid, mis
krüpteeritud sedelilegi.

.. _check-miksitud-krypteeritud-sedel-correctness:

MIKSITUD KRÜPTEERITUD SEDEL, vormingu korrektsus
````````````````````````````````````````````````

Miksitud krüpteeritud sedeli vormingu korrektsuse kontroll käesoleva
spetsifikatsiooni suhtes.

.. _check-miksitud-krypteeritud-sedel-public-key-consistency:

MIKSITUD KRÜPTEERITUD SEDEL, kooskõlalisus avaliku võtmega
``````````````````````````````````````````````````````````

Miksitud krüpteeritud sedeli kooskõlalisuse kontroll avaliku võtmega. Kontroll
õnnestub kui miksitud krüpteeritud sedeli komponendid on kasutatavad rühma
liikmetena arvutusteks avaliku võtme parameetrite poolt määratud matemaatilises
rühmas.


.. _entity-ringkonnatunnus:

```````````````
RINGKONNATUNNUS
```````````````

Olem moodustatakse töötlemisrakenduses ning viitab ringkonnapõhisele valikute
nimekirjale.

.. _check-ringkonnatunnus-correctness:

RINGKONNATUNNUS, vormingu korrektsus
````````````````````````````````````

Ringkonnatunnuse vormingu korrektsuse kontroll käesoleva spetsifikatsiooni suhtes.

.. _entity-ringkonna-valikutenimekiri:

`````````````````````````````````
RINGKONNAPÕHINE VALIKUTE NIMEKIRI
`````````````````````````````````

Olem on valijale omistatud süsteemiväliselt.

.. _check-ringkonna-valikutenimekiri-correctness:

RINGKONNAPÕHINE VALIKUTE NIMEKIRI, vormingu korrektsus
``````````````````````````````````````````````````````

Ringkonnapõhise valikute nimekirja vormingu korrektsuse kontroll VIS liideste
spetsifikatsiooni suhtes.

Kontrollid kogumisteenuses
--------------------------

Kogumisteenus käsitleb talletatavat häält sõltumatult teistest häältest.
Kogumisteenuse ülesandeks on hääle talletamine kogu hääletamisperioodi vältel
ja häälele töötlemise käigus kvalifitseerumiseks vajalike elementide hankimine.

Kogumisteenus saab valijarakenduselt järgmised olemid:

   #. :ref:`entity-krypteeritud-sedel`

   #. :ref:`entity-valija-sertifikaat`

   #. :ref:`entity-valija-identiteet`

   #. :ref:`entity-haale-signatuur`

   #. :ref:`entity-haale-allkiri`

   #. :ref:`entity-haale-konteiner`

Kogumisteenus hangib hääle töötlemise käigus välistelt teenustelt järgmised olemid:

   #. :ref:`entity-kehtivuskinnitus`

   #. :ref:`entity-ajatempel`

   #. :ref:`entity-registreerimistoend`

Kogumisteenus tuvastab / loob ise järgmised olemid:

   #. :ref:`entity-registreerimisparing`

   #. :ref:`entity-registreerimisparing-konteiner`

   #. :ref:`entity-nimekirjatunnus`

   #. :ref:`entity-ringkonnatunnus`

   #. :ref:`entity-ringkonna-valikutenimekiri`

Kogumisteenus ei puutu vahetult kokku järgmiste olemitega:

   #. :ref:`entity-tahteavaldus`

   #. :ref:`entity-juhuslikkus`

   #. :ref:`entity-miksitud-krypteeritud-sedel`

Kogumisteenus viib läbi järgmised tegevused ja teostab järgmised kontrollid:

   #. Hääle talletamise päringu vastuvõtmine valijarakenduselt

      #. :ref:`check-valija-sertifikaat-correctness`
      #. :ref:`check-valija-sertifikaat-consistency-protocol-settings`
      #. :ref:`check-valija-identiteet-correctness`
      #. :ref:`check-valija-identiteet-nimekirjatunnus-eligibility`
      #. :ref:`check-haale-signatuur-correctness`
      #. :ref:`check-haale-signatuur-krypteeritud-sedel-valija-sertifikaat-consistency`
      #. :ref:`check-haale-allkiri-correctness`
      #. :ref:`check-haale-allkiri-haale-signatuur-consistency`
      #. :ref:`check-haale-konteiner-correctness`
      #. :ref:`check-haale-konteiner-haale-allkiri-valija-sertifikaat-krypteeritud-sedel-consistency`
      #. :ref:`check-krypteeritud-sedel-correctness`
      #. :ref:`check-krypteeritud-sedel-public-key-consistency`

   #. Kehtivuskinnituse hankimine - :ref:`entity-kehtivuskinnitus`

      #. :ref:`check-kehtivuskinnitus-correctness`
      #. :ref:`check-valija-sertifikaat-kehtivuskinnitus-consistency`

   #. Ajatempli hankimine - :ref:`entity-ajatempel`

      #. :ref:`check-ajatempel-correctness`
      #. :ref:`check-ajatempel-consistency-protocol-settings`

   #. Registreerimispäringu loomine - :ref:`entity-registreerimisparing`

   #. Registreerimistõendi hankimine - :ref:`entity-registreerimistoend`

      #. :ref:`check-registreerimisparing-registreerimistoend-consistency`
      #. :ref:`check-registreerimistoend-correctness`
      #. :ref:`check-registreerimistoend-haale-allkiri-consistency`
      #. :ref:`check-kehtivuskinnitus-ajatempel-order`

   #. Hääle talletamine, kvalifitseerivate elementide ja unikaalse
      identifikaatori tagastamine valijarakendusele.

Kontrollid valijarakenduses
---------------------------

Valijarakendus moodustab valija avakujul tahteavalduse põhjal krüpteeritud
sedeli ning allkirjastab selle valija allkirja andmise vahendiga.

Valijarakenduse rolliks peale hääle allkirjastamist on veenduda, et
kogumisteenus käitus häält kvalifitseerivate elementide võtmisel
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
   osutaja poolt ning sisaldab kogumisteenuse poolt allkirjastatud päringut.

Kui hääle kvalifitseerimiseks vajalike elementide kontroll ei õnnestu, siis
teavitab valijarakendus sellest kasutajat.

Valijarakendus loob ise järgmised olemid:

   #. :ref:`entity-tahteavaldus`

   #. :ref:`entity-juhuslikkus`

   #. :ref:`entity-krypteeritud-sedel`

   #. :ref:`entity-haale-allkiri`

   #. :ref:`entity-haale-konteiner`

Valijarakendus saab järgmised olemid teistelt osapooltelt:

   #. :ref:`entity-valija-sertifikaat`

   #. :ref:`entity-valija-identiteet`

   #. :ref:`entity-ringkonna-valikutenimekiri`

   #. :ref:`entity-haale-signatuur`

   #. :ref:`entity-kehtivuskinnitus`

   #. :ref:`entity-ajatempel`

   #. :ref:`entity-registreerimistoend`

   #. :ref:`entity-registreerimisparing`

Valijarakendus ei puutu vahetult kokku järgmiste olemitega:

   #. :ref:`entity-registreerimisparing-konteiner`

   #. :ref:`entity-nimekirjatunnus`

   #. :ref:`entity-ringkonnatunnus`

   #. :ref:`entity-miksitud-krypteeritud-sedel`

Valijarakendus viib läbi järgmised tegevused ja teostab järgmised kontrollid:

   #. eID vahendi aktiveerimine ja valija identiteedi tuvastamine

      #. :ref:`entity-valija-sertifikaat`
      #. :ref:`entity-valija-identiteet`

      #. :ref:`check-valija-sertifikaat-correctness`
      #. :ref:`check-valija-sertifikaat-consistency-protocol-settings`
      #. :ref:`check-valija-identiteet-correctness`

   #. Ringkonnapõhise valikutenimekirja tuvastamine -
      :ref:`entity-ringkonna-valikutenimekiri`

   #. Tahteavalduse moodustamine - :ref:`entity-tahteavaldus`

   #. Juhuarvu genereerimine - :ref:`entity-juhuslikkus`

   #. Sedeli krüpteerimine - :ref:`entity-krypteeritud-sedel`

   #. Krüpteeritud sedeli signeerimine - :ref:`entity-haale-signatuur`

   #. Allkirja moodustamine signatuurist - :ref:`entity-haale-allkiri`

      #. :ref:`check-haale-signatuur-correctness`
      #. :ref:`check-haale-signatuur-krypteeritud-sedel-valija-sertifikaat-consistency`

   #. Allkirjastatud konteineri moodustamine - :ref:`entity-haale-konteiner`

   #. Allkirjastatud konteineri edastamine kogumisteenusele.

   #. Kogumisteenuse vastuse kontrollimine

      #. :ref:`entity-kehtivuskinnitus`
      #. :ref:`entity-ajatempel`
      #. :ref:`entity-registreerimistoend`
      #. :ref:`entity-registreerimisparing`

      #. :ref:`check-kehtivuskinnitus-correctness`
      #. :ref:`check-valija-sertifikaat-kehtivuskinnitus-consistency`
      #. :ref:`check-ajatempel-correctness`
      #. :ref:`check-ajatempel-consistency-protocol-settings`
      #. :ref:`check-kehtivuskinnitus-ajatempel-order`

      #. :ref:`check-registreerimisparing-correctness`
      #. :ref:`check-registreerimisparing-haale-allkiri-consistency`
      #. :ref:`check-registreerimisparing-registreerimistoend-consistency`
      #. :ref:`check-registreerimistoend-correctness`
      #. :ref:`check-registreerimistoend-haale-allkiri-consistency`

Kontrollid kontrollrakenduses
-----------------------------

Sarnaselt valijarakendusele on kontrollrakenduse rolliks peale hääle
allkirjastamist veenduda, et kogumisteenus käitus häält kvalifitseerivate
elementide võtmisel protokollikohaselt ning et hääl on talletatud selliselt, et
ta saab töötlemisrakenduse poolt arvesse võetud.

Täiendavalt on kontrollrakenduse ülesandeks anda valijale tagasisidet, kas tema
tahteavaldus sai valijarakenduse poolt korrektselt hääleks vormistatud.

Kui hääle kvalifitseerimiseks vajalike elementide kontroll ei õnnestu, siis
teavitab kontrollrakendus sellest kasutajat. Tahteavalduse korrektsuses peab
valija ise veenduma.

Kontrollrakendus tuvastab ise järgmised olemid:

   #. :ref:`entity-tahteavaldus`

Kontrollrakendus saab järgmised olemid teistelt osapooltelt:

   #. :ref:`entity-juhuslikkus`

   #. :ref:`entity-krypteeritud-sedel`

   #. :ref:`entity-valija-sertifikaat`

   #. :ref:`entity-valija-identiteet`

   #. :ref:`entity-haale-signatuur`

   #. :ref:`entity-haale-allkiri`

   #. :ref:`entity-haale-konteiner`

   #. :ref:`entity-kehtivuskinnitus`

   #. :ref:`entity-ajatempel`

   #. :ref:`entity-registreerimistoend`

   #. :ref:`entity-registreerimisparing`

   #. :ref:`entity-ringkonna-valikutenimekiri`

Kontrollrakendus ei puutu vahetult kokku järgmiste olemitega:

   #. :ref:`entity-registreerimisparing-konteiner`

   #. :ref:`entity-nimekirjatunnus`

   #. :ref:`entity-ringkonnatunnus`

   #. :ref:`entity-miksitud-krypteeritud-sedel`

Kontrollrakendus viib läbi järgmised tegevused ja teostab järgmised kontrollid:

   #. :ref:`check-valija-sertifikaat-correctness`
   #. :ref:`check-valija-sertifikaat-consistency-protocol-settings`
   #. :ref:`check-valija-sertifikaat-kehtivuskinnitus-consistency`
   #. :ref:`check-valija-identiteet-correctness`
   #. :ref:`check-haale-signatuur-correctness`
   #. :ref:`check-haale-signatuur-krypteeritud-sedel-valija-sertifikaat-consistency`
   #. :ref:`check-haale-allkiri-correctness`
   #. :ref:`check-haale-allkiri-haale-signatuur-consistency`
   #. :ref:`check-haale-konteiner-correctness`
   #. :ref:`check-haale-konteiner-haale-allkiri-valija-sertifikaat-krypteeritud-sedel-consistency`
   #. :ref:`check-kehtivuskinnitus-correctness`
   #. :ref:`check-ajatempel-correctness`
   #. :ref:`check-ajatempel-consistency-protocol-settings`
   #. :ref:`check-kehtivuskinnitus-ajatempel-order`
   #. :ref:`check-registreerimisparing-correctness`
   #. :ref:`check-registreerimisparing-haale-allkiri-consistency`
   #. :ref:`check-registreerimisparing-registreerimistoend-consistency`
   #. :ref:`check-registreerimistoend-correctness`
   #. :ref:`check-registreerimistoend-haale-allkiri-consistency`
   #. :ref:`check-ringkonna-valikutenimekiri-correctness`
   #. :ref:`check-juhuslikkus-correctness`
   #. :ref:`check-juhuslikkus-public-key-consistency`
   #. :ref:`check-krypteeritud-sedel-correctness`
   #. :ref:`check-krypteeritud-sedel-public-key-consistency`
   #. :ref:`check-krypteeritud-sedel-juhuslikkus-consistency`
   #. :ref:`check-tahteavaldus-correctness`
   #. :ref:`check-tahteavaldus-ringkonna-valikutenimekiri-consistency`

Kontrollid töötlemisrakenduses
------------------------------

Töötlemisrakenduse sisendiks on e-valimiskast ja registreerimisteenuse
väljavõte registreerimispäringutest. Töötlemisrakendus kontrollib mõlema
andmehulga elemente kõigepealt eraldi ning seejärel püüab luua vastavuse nende
vahel.

Töötlemisrakendus otsustab, milline valija häältest oli viimane ning liigub
töötlemise järgmisesse etappi. S.t. üks häält kvalifitseerivatest elementidest
täidab hääle talletamise aja fikseerimise rolli ning selle elemendi põhjal
moodustatakse üksikute häälte ajaline järgnevus. Olenevalt IVXV profiilist võib
see element olla kehtivuskinnituse koosseisus (BDOC-TM), eraldi ajatemplina
(BDOC-TS) või registreerimistõendi koosseisus (BDOC-TS).

Töötlemisrakendusele tehakse kättesaadavaks järgmised olemid:

   #. :ref:`entity-krypteeritud-sedel`

   #. :ref:`entity-valija-sertifikaat`

   #. :ref:`entity-valija-identiteet`

   #. :ref:`entity-haale-signatuur`

   #. :ref:`entity-haale-allkiri`

   #. :ref:`entity-haale-konteiner`

   #. :ref:`entity-kehtivuskinnitus`

   #. :ref:`entity-ajatempel`

   #. :ref:`entity-registreerimisparing`

   #. :ref:`entity-registreerimisparing-konteiner`

   #. :ref:`entity-registreerimistoend`

   #. :ref:`entity-nimekirjatunnus`

   #. :ref:`entity-ringkonna-valikutenimekiri`

Töötlemisrakendus tuvastab/loob järgmised olemid:

   #. :ref:`entity-ringkonnatunnus`


Töötlemisrakendus ei puutu vahetult kokku järgmiste olemitega:

   #. :ref:`entity-miksitud-krypteeritud-sedel`

   #. :ref:`entity-tahteavaldus`

   #. :ref:`entity-juhuslikkus`

Töötlemisrakenduse töö jaguneb neljaks etapiks:

   #. Kontrollimine

   #. Korduvhäälte eemaldamine

   #. Tühistamine/-ennistamine

   #. Anonüümimine


Töötlemisrakendus teostab kontrollimise etapis järgmised kontrollid:

   #. e-valimiskasti elementide kontrollid teostatakse kõigi hääle elementide kohta:

      #. :ref:`check-valija-sertifikaat-correctness`
      #. :ref:`check-valija-sertifikaat-consistency-protocol-settings`
      #. :ref:`check-valija-identiteet-correctness`
      #. :ref:`check-nimekirjatunnus-correctness`
      #. :ref:`check-nimekirjatunnus-consistency-protocol-settings`
      #. :ref:`check-valija-identiteet-nimekirjatunnus-eligibility`
      #. :ref:`check-ringkonnatunnus-correctness`
      #. :ref:`check-haale-signatuur-correctness`
      #. :ref:`check-haale-signatuur-krypteeritud-sedel-valija-sertifikaat-consistency`
      #. :ref:`check-haale-allkiri-correctness`
      #. :ref:`check-haale-allkiri-haale-signatuur-consistency`
      #. :ref:`check-haale-konteiner-correctness`
      #. :ref:`check-haale-konteiner-haale-allkiri-valija-sertifikaat-krypteeritud-sedel-consistency`
      #. :ref:`check-kehtivuskinnitus-correctness`
      #. :ref:`check-valija-sertifikaat-kehtivuskinnitus-consistency`
      #. :ref:`check-ajatempel-correctness`
      #. :ref:`check-ajatempel-consistency-protocol-settings`
      #. :ref:`check-kehtivuskinnitus-ajatempel-order`
      #. :ref:`check-registreerimistoend-correctness`
      #. :ref:`check-registreerimistoend-haale-allkiri-consistency`

   #. registreerimisteenuse väljavõtte kontrollid teostatakse iga
      registreerimispäringu kohta

      #. :ref:`check-registreerimisparing-correctness`
      #. :ref:`check-registreerimisparing-konteiner-correctness`
      #. :ref:`check-registreerimisparing-konteiner-registreerimispäring-consistency`

   #. Töötlemisrakendus loob vastavuse e-valimiskasti ja registreerimisteenuse
      väljavõtte vahel võttes aluseks olemi
      :ref:`entity-registreerimisparing-konteiner` registreerimisteenuse
      väljavõttest ning olemi :ref:`entity-registreerimistoend`
      e-valimiskastist. Ühendavaks lüliks kahe vaate vahel on
      :ref:`entity-registreerimisparing`.

   #. e-valimiskasti ja registreerimisteenuse väljavõtte vastavuskontrollid

      #. :ref:`check-registreerimisparing-haale-allkiri-consistency`
      #. :ref:`check-registreerimisparing-registreerimistoend-consistency`

Töötlemisrakendus teostab korduvhäälte eemaldamise etapis järgmised kontrollid:

   #. Töötlemisrakendus tuvastab iga valija häälte hulgast ajaliselt viimase.

   #. Töötlemisrakendus viib läbi krüpteeritud sedeli korrektsuse kontrollid.

      #. :ref:`check-krypteeritud-sedel-correctness`
      #. :ref:`check-krypteeritud-sedel-public-key-consistency`


Töötlemisrakenduse töö järgmistes etappides täiendavaid kontrolle ei teostata.
Tühistamis-/ennistamisetapis eemaldatakse/taastatakse isikukoodile vastavaid
hääli korduvhäältest puhastatud e-valimiskastist. Anonüümimisetapis
eemaldatakse e-häältelt kvalifitseerivad elemendid ning saadud loend
krüpteeritud sedelitest suunatakse miksimisrakendusse.

Kontrollid miksimisrakenduses
-----------------------------

Miksimisrakendusele tehakse kättesaadavaks järgmised olemid:

   #. :ref:`entity-krypteeritud-sedel`

   #. :ref:`entity-ringkonnatunnus`

Miksimisrakendus loob järgmised olemid:

   #. :ref:`entity-miksitud-krypteeritud-sedel`

Miksimisrakendus ei puutu vahetult kokku järgmiste olemitega:

   #. :ref:`entity-tahteavaldus`

   #. :ref:`entity-juhuslikkus`

   #. :ref:`entity-valija-sertifikaat`

   #. :ref:`entity-valija-identiteet`

   #. :ref:`entity-haale-signatuur`

   #. :ref:`entity-haale-allkiri`

   #. :ref:`entity-haale-konteiner`

   #. :ref:`entity-kehtivuskinnitus`

   #. :ref:`entity-ajatempel`

   #. :ref:`entity-registreerimisparing`

   #. :ref:`entity-registreerimisparing-konteiner`

   #. :ref:`entity-registreerimistoend`

   #. :ref:`entity-nimekirjatunnus`

   #. :ref:`entity-ringkonna-valikutenimekiri`


Miksimisrakendus viib läbi järgmised tegevused ja kontrollid:

   #. :ref:`check-ringkonnatunnus-correctness`
   #. :ref:`check-krypteeritud-sedel-correctness`
   #. :ref:`check-krypteeritud-sedel-public-key-consistency`

   #. Miksimisrakendus grupeerib sisendiks antud krüpteeritud sedelid
      ringkonnatunnuste kaupa ning järjestab krüpteeritud sedelid ringkonnas
      ringi.

   #. Miksimisrakendus arvutab iga ümberjärjestatud krüpteeritud sedeli kohta
      uue miksitud krüpteeritud sedeli.

   #. Miksimisrakendus koostab nullteadmustõestused sedelite korrektse miksimise kohta.

Kontrollid võtmerakenduses
--------------------------

Võtmerakendusele tehakse kättesaadavaks järgmised olemid:

   #. :ref:`entity-miksitud-krypteeritud-sedel`

   #. :ref:`entity-ringkonnatunnus`

   #. :ref:`entity-ringkonna-valikutenimekiri`

Võtmerakendus tuvastab järgmised olemid:

   #. :ref:`entity-tahteavaldus`

Võtmerakendus ei puutu vahetult kokku järgmiste olemitega:

   #. :ref:`entity-juhuslikkus`

   #. :ref:`entity-krypteeritud-sedel`

   #. :ref:`entity-valija-sertifikaat`

   #. :ref:`entity-valija-identiteet`

   #. :ref:`entity-haale-signatuur`

   #. :ref:`entity-haale-allkiri`

   #. :ref:`entity-haale-konteiner`

   #. :ref:`entity-kehtivuskinnitus`

   #. :ref:`entity-ajatempel`

   #. :ref:`entity-registreerimisparing`

   #. :ref:`entity-registreerimisparing-konteiner`

   #. :ref:`entity-registreerimistoend`

   #. :ref:`entity-nimekirjatunnus`


Võtmerakendus viib läbi järgmised tegevused ja kontrollid:

   #. :ref:`check-miksitud-krypteeritud-sedel-correctness`
   #. :ref:`check-miksitud-krypteeritud-sedel-public-key-consistency`
   #. Häälte dekrüpteerimine
   #. :ref:`check-ringkonnatunnus-correctness`
   #. :ref:`check-tahteavaldus-correctness`
   #. :ref:`check-tahteavaldus-ringkonnatunnus-valikutenimekiri-consistency`
