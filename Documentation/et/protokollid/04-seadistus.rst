..  IVXV protokollid

================================================================================
Valimise definitsioon
================================================================================

Valimise defineerib valimise korraldaja. Eesti riiklikel valimistel jagunevad
kõik hääleõiguslikud isikud ühte või mitmesse valimisringkonda. Valijal on
võimalik hääletamisel valida ainult selle ringkonna kandidaatide vahel, kuhu ta
kuulub.

Valimise defineerimiseks tuleb määratleda vähemalt

#. valimise unikaalne identifikaator ning küsimuste unikaalsed identifikaatorid;

#. täielik loend valimisringkondadest ja -jaoskondadest;

#. hääleõiguslike isikute nimekiri ja jagunemine valimisringkondadesse;

#. kandidaatide nimekiri ja jagunemine valimisringkondadesse.

Valimise ja küsimuste identifikaatorid
================================================================================

Üht valimist puudutav andmestik on seotud unikaalse valimise identifikaatori
abil. Tüüpiliselt hääletatakse ühel konkreetsel valimisel täpselt ühes
küsimuses. Siiski eksisteerib võimalus, et valimisel esitatakse mitu küsimust.
Kõik küsimused on samamoodi eristatud unikaalse identifikaatori abil.

Identifikaatorite pikkus on piiratud 28 tähemärgiga ASCII kooditabelist.
Konkreetse valimise tarbeks kasutatavad identifikaatorid spetsifitseeritakse
igakordselt valimise seadistustes. Spetsifikatsiooni realiseerivad rakendused
peavad keelduma töötlemast andmeid, mis identifitseerivad valimise/küsimuse, mis
ei kuulu rakenduse jaoks seadistatud valimiste/küsimuste nimekirja.

.. code-block:: bnf

    election-identifier := 1*28CHAR

    question-identifier := 1*28CHAR


Valimisjaoskondade ja -ringkondade nimekiri
================================================================================

Kandidaate on võimalik valimisele üles seada ainult konkreetses
valimisringkonnas. Ringkondade abil antakse valijatele hääletamise valikud:

#. Ringkonnad jagunevad jaoskondadeks;

#. Iga valija kuulub talle määratud jaoskonda ja selle kaudu ka ringkonda;

#. Kõigis ühe ringkonna jaoskondades saavad valijad
   teha valiku vaid selle ringkonna valikute vahel;

Eesti riiklikel valimistel eristatakse Kohalike Omavalitsuste Volikogude (KOV)
valimisi, Riigikogu valimisi, Euroopa Parlamendi valimisi ning rahvahääletusi.

KOV korraldatakse valimised vastavalt seadusele "Kohaliku omavalitsuse volikogu
valimise seadus" [KOVVS]. Valimine toimub kohaliku omavalitsuse tasandil, igal
omavalitsusel on oma hääletamistulemus. Valimisringkonnad moodustatakse
omavalitsuse tasemel vastavalt seaduses kirjeldatud reeglitele.

Riigikogu valimised korraldatakse vastavalt seadusele "Riigikogu valimise seadus"
[RKVS]. Valimine toimub riigi tasandil, hääletamistulemus on kõigile kohalikele
omavalitsustele ühine. Riik jaguneb 12 valimisringkonnaks.

Europarlamendi valimised korraldatakse vastavalt seadusele "Euroopa Parlamendi
valimise seadus" [EPVS]. Valimine toimub riigi tasandil, hääletamistulemus on
kõigile kohalikele omavalitsustele ühine. Terve riik on üks suur
valimisringkond.

Rahvahääletused korraldatakse vastavalt seadusele "Rahvahääletuse seadus" [RHS].
Valimine toimub riigi tasandil, hääletamistulemus on kõigile kohalikele
omavalitsustele ühine. Terve riik on üks suur valimisringkond.

Erinevad valimised ei erine elektroonilise hääletamise andmevormingute ja
protseduuride poolest. Erinevad ringkondade jaotused hallatakse Valimiste
Infosüsteemi poolt.

Kandidaate on võimalik valimisele üles seada ainult konkreetses
valimisringkonnas. Ringkonnad jagunevad jaoskondadeks ning valijad jaotatakse
jaoskondade vahel. Kõigis ühe ringkonna jaoskondades saavad konkreetse jaoskonna
alla kuuluvad valijad hääletada selle ringkonna kandidaatide poolt. Valija
jaoskonnakuuluvuse kaudu on määratud ka tema ringkonnakuuluvus. Valija saab teha
valiku ainult tema ringkonnas kandideerivate valikute vahel.

Kuna kohaliku omavalitsuse volikogude valimisel toimub valimine Eesti
omavalitsuste (vallad, linnad) tasemel, siis kasutatakse elektroonilise
hääletamise protokollistikus valimisringkondade, ja -jaoskondade kirjeldamisel
ning valijate ja valikute ringkonnakuuluvuse näitamisel `Eesti haldus- ja
asustusjaotuse klassifikaatorit EHAK <http://ads.maaamet.ee/>`_

Näiteks:

* Tallinna linna Pirita linnaosa EHAK kood on 0596;

* Aegviidu valla EHAK kood on 0112.

Riigi tasemel toimuvatel valimistel pannakse ringkonna EHAK koodiks
kokkuleppeliselt 0. Valimisjaoskondade EHAK koodiks pannakse selle omavalitsuse
kood, mille koosseisus konkreetne jaoskond on moodustatud.

Riigikogu ja europarlamendi valimistel ning rahvahääletusel moodustatakse
valimisjaoskondade ja –ringkondade nimekirjas igasse ringkonda fiktiivne
jaoskond välismaal hääletajate tarbeks. Välismaalaste puhul valimisjaoskonna
number on 0 ning vastav EHAK kood on samuti 0.

.. code-block:: bnf

   ehak-code = 1*10DIGIT

   ehak-district = ehak-code
   no-district = 1*10DIGIT

   ehak-station = ehak-code
   no-station = 1*10 DIGIT

   district = ehak-district '.' no-district
   district-legacy = ehak-district TAB no-district

   station = ehak-station '.' no-station
   station-legacy = ehak-station TAB no-station TAB district-legacy


Ringkondade nimekirja JSON-vorming on defineeritud järgnevalt. Objekti
:token:`region_dict` elemente indekseeritakse elemendiga tüüpi
:token:`ehak-code`. Objekti :token:`district_dict` elemente indekseeritakse
elemendiga tüüpi :token:`district`. Massiivi :token:`stations` elemendid on
tüüpi :token:`station`.

.. literalinclude:: schema/ivxv.districts.schema
   :language: json
   :linenos:

Näide:

.. literalinclude:: schema/ivxv.districts.schema.example
   :language: json

Ringkondade nimekiri saadakse Valimiste Infosüsteemist ning JSON-vormingus faili
vahendatakse elektroonilise hääletamise süsteemile BDOC-vormingus
digitaalallkirjastatud faili koosseisus.


Valijate nimekiri
================================================================================

Valijate nimekiri sisaldab valijate nimesid, isikukoode, valimisjaoskonda ning
rea numbrit valimisjaoskonna valijate nimekirjas, milles valija hääletab.
Valijate nimekirja vorming on järgmine:

.. code-block:: bnf

    voter-personalcode = 11DIGIT
    voter-name = 1*100UTF-8-CHAR
    action = "lisamine" | "kustutamine"
    line-no = "" | 1*11DIGIT
    reason = "" | "tokend" | "jaoskonna vahetus" | "muu"

    voter = voter-personalcode TAB voter-name TAB action TAB station-legacy TAB line-no TAB reason LF


    version-no = "1"
    list-type = "algne" | "muudatused"
    voter-list = version-no LF election-identifier LF list-type LF *voter

Pärandsüsteemide andmestruktuurid sisaldavad välja versiooninumber, mille pikkus
on piiratud 2 tähemärgiga. Välja väärtus on 1.

Valijate nimekiri võib olla kas algne nimekiri või muudatusnimekiri. Algne
nimekiri lubab ainult valijate lisamisi, muudatusnimekirja korral on võimalik ka
valijate eemaldamine nimekirjast. Valijakirje võib sisaldada täiendavat
informatsiooni -- valijakirje reanumbrit jaoskonna nimekirjas ning põhjust
konkreetses muudatusnimekirjas esinemiseks.


Andmete sisu on järgmine.

#. Tüüp (``list-type``) "``algne``" tähistab esialgset suurt nimekirja, mis
   laaditakse süsteemi enne e-hääletamise algust ja "``muudatused``" hilisemaid
   kumulatiivseid uuendusi.

#. Tegevus (``action``) "``lisamine``" tähendab uue valija lisamist nimetatud
   valimisjaoskonda ja "``kustutamine``" eemaldamist. Kui valija liigub ühest
   jaoskonnast teise, siis kantakse valijate nimekirja muudatuste hulka üks
   kustutamise kirje, millega valija oma eelmisest valimisjaoskonnast
   kustutatakse ja üks lisamise kirje, millega valija uues valimisjaoskonnas
   valijate nimekirja kantakse. Algses nimekirjas on kõik kirjed "lisamine"
   tüüpi.

#. Jaoskond (``station-legacy``) identifitseerib jaoskonna ja ringkonna (ning
   nende kaudu ka omavalitsuse), kus valija hääletab.

#. Reanumber (``line-no``) on valija rea number valimisjaoskonna nimekirjas.
   Täidetud ainult algse nimekirja puhul, muudatuste korral on see väli tühi.

#. Põhjust (``reason``) kasutatakse kustutamiskirjete juures märkimaks
   kustutamise põhjust.  Lisamiskirjete korral peab põhjus tühi olema.  Põhjus
   ``tokend`` tähendab, et muudatuse rakendumisest alates ei tohi vastava
   isikukoodiga valija enam hääletada. Põhjus ``jaoskonna vahetus`` tähendab,
   et valija kustutatakse ühest jaoskonnast, kuna ta lisatakse teise jaoskonda.
   Sellisel juhul peab kaasnema kustutamiskirjega ka lisamiskirje (seda
   kontrollitakse). Kui kasutaja eemaldatakse nimekirjast mingil muul põhjusel
   (surm, mujale (piirkonda, mis ei osale valimistel) elama kolimine), peab
   põhjuseks olema ``muu`` või võib põhjus ka tühi olla.  Väli on
   informatiivne.


Valijate nimekirja signatuur
----------------------------


Valijate nimekiri saadakse SMIT-i poolt hallatavast Rahvastikuregistrist.
Pärandvormingus tekstifailile kaasatakse allkirjafail, mille moodustab
Rahvastikuregister, arvutades algsest valijate nimekirjast SHA256-räsi ning
allkirjastades selle räsi 2048 bitise RSA-võtmega. Rahvastikuregistri poolt
genereeritud avalik võti tehakse kättesaadavaks elektroonilise hääletamise
infosüsteemile ning selle võtme alusel kontrollitakse valijate nimekirjade
terviklust. Skeem on kasutusel alates 2015. aasta Riigikogu valimistest.

..
   Võtme genereerimine:
   openssl genrsa -out voterlist-signer.key 2048
   Avaliku võtme eraldamine:
   openssl rsa -in voterlist-signer.key -pubout -out voterlist-signer.pem
   Allkirjastamine:
   openssl dgst -sha256 -sign voterlist-signer.key -out voters.signature voters
   Verifitseerimine:
   openssl dgst -sha256 -verify voterlist-signer.pem -signature voters.signature voters

Rakendatud nimekirja versioon
-----------------------------

Rakendatud valijate nimekiri mingis ajahetkes sõltub algnimekirjast, sellele
rakendatud muudatustest ja muudatuste rakendamise järjekorrast. Selle seisu
ühtlaseks tuvastamiseks tuleb arvutada nimekirja versioon.

.. attention::

   Nimekirja versioon ei ole seotud nimekirja failis sisalduva
   versiooninumbriga, mis määrab nimekirja formaadi versiooni.

Versiooni arvutamine toimub järgmiselt::

    v_0 = ""
    v_n = base64(sha256(v_{n-1} | base64(sha256(nk_n))))

kus ``nk_n`` on ``n``-is laaditud nimekiri (lugemine algab ühest ehk
algnimekiri on ``nk_1``), ``v_n`` on valijate nimekirja versioon pärast selle
laadimist, ``""`` on tühi sõne ja ``|`` on sõnede sidurdamine operatsioon.

Rakendatud nimekirja versiooni üle peavad arvet kogumisteenus ja
töötlemisrakendus, mis tagavad, et iga hääl läheks arvesse õiges ringkonnas.

Valikute nimekiri
================================================================================

Valikute nimekiri sisaldab andmeid kandidaatide (valimistel) või
vastusevariantide (rahvahääletusel) kohta. Valimiste korral on lisaks kandidaadi
andmetele nimekirjas ka tema valimisnimekirja nimi.

Valijale elektroonilise hääletamise käigus nähtavaid valimiste vahelisi
süsteemseid erinevusi on kaks:

#. Rahvahääletusel ei valita erakondadesse kuuluvate kandidaatide vahel vaid
   vastatakse "JAH"/"EI" konkreetsetele küsimustele;

#. Riigikogu, KOV ja Euroopa Parlamendi valimistel antakse hääl ühele
   kandidaadile, kes võib, aga ei pruugi kuuluda suuremasse erakonda/nimekirja.

Protokollistik kodeerib valija võimalikud valikud ringkonnas kuni 11-kohalise
arvväärtusena, mis valikute nimekirjas kodeeritakse koos ringkonna EHAK-koodiga.
Valijale tohivad kättesaadavad olla ainult tema ringkonnakohased valikud.
Valijarakendus peab seda omadust tagama ning hääletamistulemust arvutav rakendus
kontrollima.

.. code-block:: bnf

    choice-no = 1*11DIGIT
    district-choice = ehak-district '.' choice-no


Valikute nimekirja JSON-vorming on defineeritud järgnevalt: Objekti
:token:`district_dict` elemente indekseeritakse elemendiga tüüpi
:token:`district`.
Objekti :token:`list-choices` elemente indekseeritakse elemendiga tüüpi
:token:`district-choice`.

.. literalinclude:: schema/ivxv.choices.schema
   :language: json
   :linenos:

Näide:

.. literalinclude:: schema/ivxv.choices.schema.example
   :language: json
