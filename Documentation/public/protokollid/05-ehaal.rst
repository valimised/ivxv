..  IVXV protokollid

===================
Elektrooniline hääl
===================

IVXV hääletamisprotokoll põhineb topeltümbrikuskeemil, mis tähendab, et valija
avakujul tahteavaldus krüpteeritakse valimise korraldaja poolt levitatud
avaliku võtmega. Krüpteeritud tahteavaldus allkirjastatakse digitaalselt valija
käsutuses oleva allkirjastamisvahendiga ning edastatakse kogumisteenusesse
mingis kokkulepitud konteinervormingus. Kogumisteenus võib valija poolt
allkirjastatud häält täiendavalt kvalifitseerida, veendudes näiteks
allkirjastamissertifikaadi kehtivuses. IVXV protokollistik näeb mh. ette
kogumisteenuse poolt vastuvõetud häälte registreerimise välises
registreerimisteenuses.

Kogumisteenuse poolt talletamisele võetud hääl koos kvalifitseerivate
elementidega tehakse kättesaadavaks nii valijarakendusele kui
kontrollrakendusele, mis teostavad üksiku hääle peal samad kontrollid, mida
hilisem valimise korraldaja töötlemisrakendus teostab kõigi häälte peal.
Kvalifitseerivate elementide kontrollimise võimalus annab valijale kindluse, et
tema häält on hilisemates protsessides korrektselt menetletud.


Valija tahteavaldus avakujul
============================

Valija tahteavaldus avakujul eksisteerib valijarakenduses ning hiljem ka
kontrollrakenduses ja häälte kokkulugemisel võtmerakenduses. Tahteavaldus
sisaldab valiku koodi ringkonnas ja ringkonna EHAK-koodi. Eraldajana
kasutatakse ASCII kooditabeli sümbolit `0x2E` ehk ".".
Vormingus spetsifitseerimata sümbolite kasutamine ei ole lubatud, vormingule
mittevastavad sedelid loetakse kehtetuks.

.. code-block:: bnf

   <digit> ::= "0" | "1" | "2" | "3" | "4" | "5" | "6" | "7" | "8" | "9"

   <ehak-kood> ::= <digit> <digit> <digit> <digit>

   <kandidaadi-number> ::= <digit> <digit> <digit>
                         | <digit> <digit> <digit> <digit>

   <dot> ::= 0x2E

   <ballot> ::= <ehak-kood> <dot> <kandidaadi-number>


Järgmised tahteavaldused on vormistatud korrektselt:

.. code-block:: bnf

   0000.123
   0321.1234
   0000.1234

Järgmised tahteavaldused ei ole vormistatud korrektselt:

.. code-block:: bnf

   0000 . 123      (*  vorming ei näe ette tühikute kasutamist     *)
   0000-123        (*  vorming lubab eraldajana vaid sümbolit "."  *)
   123             (*  mõlemad vormingukomponendid tuleb esitada   *)
   321.1234        (*  EHAK kood on 4-kohaline identifikaator      *)


Krüpteeritud sedel
==================

Valija tahteavaldus avakujul :token:`ballot` krüpteeritakse valijarakenduse
poolt valimise korraldaja genereeritud avaliku võtmega. IVXV vajab
krüpteerimiseks mitte-deterministlikku, homomorfset avaliku võtme
krüptosüsteemi. Selliseks süsteemiks sobib ElGamal krüptosüsteem, mida täna
rakendatakse IVXV kontekstis nii jäägiklassiringi `Zp` multiplikatiivsel rühmal
`Zp*` (MODP tüüpi rühmad) kui ka elliptkõveratel (ECC tüüpi rühmad).

ElGamal avalik võti kodeeritakse koos ElGamal krüptosüsteemi parameetritega ning
konkreetset valimist iseloomustava identifikaatoriga. Krüptosüsteemi parameetrid
on osaks algoritmi identifikaatori struktuurist, avalik võti on kodeeritud X509
standardis kirjeldatud :token:`SubjectPublicKeyInfo` struktuuri
:token:`subjectPublicKey` välja.

::

    SubjectPublicKeyInfo ::= SEQUENCE {
        algorithm   AlgorithmIdentifier,
        subjectPublicKey    BIT STRING
    }

MODP tüüpi rühmade korral identifitseerib algoritmi identifikaator
:token:`id-ivxv-modp-elgamal`. Parameetrid vastavad struktuurile
:token:`IVXVModPElGamalParameters` ning avalik võti struktuurile
:token:`IVXVModPElGamalPublicKey` (:ref:`asn-modp`).

ECC tüüpi rühmade korral identifitseerib algoritmi identifikaator
:token:`id-ivxv-ecc-elgamal`. Parameetrid vastavad struktuurile
:token:`IVXVECCElGamalParameters` ning avalik võti struktuurile
:token:`IVXVECCElGamalPublicKey` (:ref:`asn-ecc`).

Valija tahteavalduse krüpteerimiseks võetakse UTF-8 kodeeringus struktuur
:token:`ballot` ning teisendatakse see ElGamal parameetrite poolt kirjeldatud
rühma elemendiks vastavalt MODP või ECC algoritmidele. Rühma elemendiks
teisendatud tahteavaldus krüpteeritakse ElGamal algoritmiga ning kodeeritakse
MODP korral struktuuri :token:`IVXVModPElGamalCiphertext` (:ref:`asn-modp`) ja
ECC korral struktuuri :token:`IVXVECCElGamalCiphertext` (:ref:`asn-ecc`).
Lõplik krüpteeritud tahteavaldus esitatakse struktuuris
:token:`IVXVElGamalCiphertext` (:ref:`asn-general`), kus väli
:token:`algorithm` viitab rühma tüübile ning väli :token:`ciphertext`
rühmaspetsiifilisele krüptogrammile. Andmestruktuuri
:token:`IVXVElGamalCiphertext` DER-kodeering on krüpteeritud sedel ehk sisemine
ümbrik topeltümbriku skeemis.

..
  NB! Nende algoritmide spetsifikatsioonidele tuleb viidata.

Tahteavalduse krüpteerimise käigus genereeritakse valijarakenduses juhuarv, mida
ElGamal krüpteerimisel kasutab. Sama juhuarv avalikustatakse hiljem
kontrollrakendusele. Tulenevalt ElGamal krüptosüsteemi eripärast funktsioneerib
see juhuarv nö. teise võtmena ning võimaldab krüptogrammi dekodeerimist
kontrollrakenduses.

.. _signed-vote:

Valija poolt allkirjastatud hääl
================================

Krüpteeritud sedel tuleb enne kogumisteenusesse talletamisele saatmist
digitaalselt allkirjastada, milleks on võimalik kasutada kõiki Eesti Vabariigis
kehtivaid digitaalallkirjavahendeid – ID-kaart, Digi-ID, Mobiil-ID, Smart-ID.
ID-kaarti saab kasutada nii TLS-CCA kui Web-eID protokolliga.

.. attention::

   Loend digitaalallkirjavahenditest ja nendega seotud protokollidest on
   tehniliselt korrektne, kuid valimiste korraldajal on võimalus kasutada vaid
   mõnda alamhulka nimetatud vahenditest.

Käesolev spetsifikatsioon näeb ette Eesti Vabariigi Standardikavandis [BDOC2.1]
defineeritud BDOC allkirjavormingu kasutamise. BDOC allkirjavorming koosneb ETSI
standardi TS 101 903 (XadES) profiilist ning OpenDocument konteineri vormingust.

Olenevalt käimasoleval valimisel esitatud küsimuste arvust võib digitaalselt
allkirjastatud hääl sisaldada ühte või mitut andmefaili MIME-tüübiga
``application/octet-stream``. Iga andmefaili sisuks on krüpteeritud sedel.
Andmefaili ja teiste signeeritavate andmeobjektide räsimiseks enne
allkirjastamist kasutatakse räsifunktsiooni SHA-256. Andmefaili nimi
moodustatakse laiendist '``ballot``' ning valimise identifikaatorist ja
küsimuse identifikaatorist. Kõik viidatud andmefailid peavad sisalduma
allkirjakonteineris. Digitaalselt allkirjastatud hääl ei tohi sisaldada muid
andmefaile kui neid, mis sisaldavad hääli mõne käimasoleva valimise kontekstis.
Seadistusele mittevastavate häälte vastuvõtmisest, talletamisest ja
töötlemisest peab kogumisteenus keelduma.

.. attention::

   Valimise identifikaator ja küsimuse identifikaator on defineeritud valimise
   korraldaja poolt loodud seadistustes ning töötlevad rakendused peavad
   lähtuma neist seadistustest otsustamaks, millised vormingule vastavad
   identifikaatorid on konkreetse sündmuse kontekstis lubatud ja millised
   mitte.


.. code-block:: bnf

   <ascii-char> ::= [ASCII]

   <election-identifier> ::= 1..28 <ascii-char>

   <question-identifier> ::= 1..28 <ascii-char>

   <extension> ::= "ballot"

   <dot> ::= 0x2E

   <encrypted-ballot-name> ::= <election-identifier> <dot> <question-identifier> <dot> <extension>


Valija poolt valijarakenduses allkirjastatud hääl moodustatakse nii, et on
võimalik selle edasine kvalifitseerimine kogumisteenuses. Käesolev
spetsifikatsioon näeb ette hääle kvalifitseerimiseks nii PKIX ajatempli kui
OCSP kehtivuskinnituse võtmise. Sellisena on lõplik kvalifitseeritud hääl
vastav BDOC-TS profiilile.

Kui hääl allkirjastatakse ID-kaardi või Digi-ID'ga, siis toimub algse
allkirjastatud konteineri moodustamine valijarakenduses. Kui hääl
allkirjastatakse Mobiil-ID või Smart-ID'ga, siis toimub konteineri moodustamine
valijarakenduse ning kogumisteenuse poolt vahendatava Mobiil-ID/Smart-ID
teenuse koostöös. Mobiil-ID/Smart-ID juhtumil kasutab kogumisteenus
Mobiil-ID/Smart-ID teenust ainult signatuuri saamiseks krüpteeritud sedelile.
Kõik hääle kvalifitseerimiseks vajalikud elemendid hangitakse vastavatelt
teenustelt alles siis kui valijarakendus on saatnud signeeritud hääle
talletamiseks. Kvalifitseeritud hääl esitatakse kogumisteenuse poolt
valijarakendusele verifitseerimiseks, ainult kvalifitseeritud hääl peab vastama
BDOC 2.1 standardi tingimustele -- valijarakenduse poolt moodustatud hääl on
vaheetapp kvalifitseeritud hääleni jõudmiseks.

Valijarakenduses signeeritud häälel peab olema üks ja ainult üks allkiri, mida
hoitakse signatuurifailis :file:`META-INF/signature0.xml`. Häält ja allkirja
sisaldav konteiner (edaspidi viidatud kui ``SignedVote``) moodustatakse BDOC 2.1
standardis kirjeldatud meetodit kasutades.

Spetsifitseerime valijarakenduses allkirjastatud hääle vormingu ühe
küsimuse korral.

Räsialgoritmina ``DIGEST_ALG`` on kasutusel SHA-256
(http://www.w3.org/2001/04/xmlenc#sha256). XML kanoniseerimiseks
(``CANON_ALG``) kasutatakse meetodit ``c14n11``
(http://www.w3.org/2006/12/xml-c14n11).

ECC võtmete korral on allkirjastamismeetodiks
http://www.w3.org/2001/04/xmldsig-more#ecdsa-sha256. RSA võtmeid enam ei
kasutata.

Identifikaatorite `VOTE_REF`, `SP_REF`, `SP_URI` ning `SV_URI` täpne väärtus ei
ole fikseeritud.

Element `SignedProperties`
---------------------------

Element ``SignedProperties`` moodustatakse kooskõlas BDOC 2.1 standardiga. Kui
kvalifitseerimisel kasutatakse ajatemplit, siis elementi
``SignaturePolicyIdentifier`` ei kasutata. Ühtegi mitte-kohustuslikku elementi
ei kasutata. Allkirjastamise kellaaja fikseerib andmestruktuuri täitev arvuti
ning valija X509-sertifikaat saadakse kas ID-kaardilt, Mobiil-ID, Smart-ID
või Web eID teenuse vahendusel.


.. literalinclude:: ../../common/xmltemplates/sp.template
   :language: xml
   :linenos:


Element `SignedInfo`
---------------------

Element ``SignedInfo`` moodustatakse kooskõlas BDOC 2.1 standardiga, viidates
nii krüpteeritud sedelile (``VOTE_DIGEST``) kui elemendile ``SignedProperties``
(``SP_DIGEST``).

.. literalinclude:: ../../common/xmltemplates/si.template
   :language: xml
   :linenos:


Element `SignatureValue`
------------------------

Element ``SignatureValue`` moodustatakse kooskõlas BDOC 2.1 standardiga.
Kanoniseeritud elemendist ``SignedInfo`` arvutatakse räsi, mis allkirjastatakse
PKCS1-meetodiga.

.. literalinclude:: ../../common/xmltemplates/sv.template
   :language: xml
   :linenos:


Element `XAdESSignatures`
-------------------------

Element ``XAdESSignatures`` sisaldab ühte ``Signature`` elementi, mis on
koostatud lähtudes kõigist eelmistest elementidest ning valija X509
sertifikaadist. Elementi ``UnsignedProperties`` ei kasutata.

.. literalinclude:: ../../common/xmltemplates/sig.template
   :language: xml
   :linenos:
