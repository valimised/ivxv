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
kontrollrakenduses. Tahteavaldus sisaldab nii valiku koodi ringkonnas, ringkonna
EHAK-koodi kui ka valiku nimekirja nime ning konkreetse valiku nime nimekirjas.

.. code-block:: bnf

    choice-name = 1*100UTF-8-CHAR
    choicelist-name = 1*100UTF-8-CHAR

    ballot = district-choice '\x1F' choicelist-name '\x1F' choice-name


Krüpteeritud sedel
==================

Valija tahteavaldus avakujul :token:`ballot` krüpteeritakse valijarakenduse
poolt valimise korraldaja genereeritud avaliku võtmega. IVXV vajab
krüpteerimiseks mitte-deterministlikku, homomorfset avaliku võtme
krüptosüsteemi. Selliseks süsteemiks sobib ElGamal krüptosüsteem, mida täna
rakendatakse IVXV kontekstis jäägiklassi rühmal.

ElGamal avalik võti kodeeritakse koos ElGamal krüptosüsteemi parameetritega ning
konkreetset valimist iseloomustava identifikaatoriga. Krüptosüsteemi parameetrid
on osaks algoritmi identifikaatori struktuurist, avalik võti on kodeeritud
:token:`SubjectPublicKeyInfo` struktuuri.

::

    elGamalEncryption OBJECT IDENTIFIER ::= {
        {iso(1) org(3) dod(6) internet(1) private(4) enterprise(1) dds(3029) asymmetric-encryption(2) 1}
    }

    elGamal-Params-IVXV ::= SEQUENCE {
        p                   INTEGER,
        g                   INTEGER,
        election-identifier GeneralString
    }

    elGamalPublicKey ::= SEQUENCE {
        y           INTEGER,
    }

    SubjectPublicKeyInfo ::= SEQUENCE {
        algorithm   AlgorithmIdentifier,
        subjectPublicKey    BIT STRING
    }


Valija tahteavalduse krüpteerimiseks võetakse UTF-8 kodeeringus struktuur
:token:`ballot` ning teisendatakse see ElGamal parameetrite poolt kirjeldatud
rühma elemendiks. Eeldame, et parameeter :token:`p` on 256 baiti. Sellisel juhul
võib struktuuri :token:`ballot` pikkus olla 253 baiti. Avakujul tahteavaldus
pikendatakse parameetri :token:`p` pikkuseni.

.. code-block:: bnf

    padded-ballot = ballot '\x00' '\x01' *'\xff' '\x00'


Pikendatud tahteavaldust interpreteeritakse kui täisarvu, mis kodeeritakse
ruutjäägina parameetri :token:`p` poolt kirjeldatud rühmas. Kodeerimine on
üksühene ning oluline krüptogrammi edasise miksimise jaoks.

Tahteavaldus krüpteeritakse vastavalt ElGamal meetodile avaliku võtmega.

::

    elGamalEncryptedMessage ::= SEQUENCE {
        a           INTEGER,
        b           INTEGER
    }

    encryptedBallot ::= SEQUENCE {
        algorithm   AlgorithmIdentifier,
        cipher  ANY
    }


Andmestruktuuri :token:`encryptedBallot` DER-kodeering on krüpteeritud sedel ehk
sisemine ümbrik topeltümbriku skeemis.

Tahteavalduse krüpteerimise käigus genereeritakse valijarakenduses juhuarv, mida
ElGamal krüpteerimisel kasutab. Sama juhuarv avalikustatakse hiljem
kontrollrakendusele. Tulenevalt ElGamal krüptosüsteemi eripärast funktsioneerib
see juhuarv nö. teise võtmena ning võimaldab krüptogrammi dekodeerimist
kontrollrakenduses.


Valija poolt allkirjastatud hääl
================================

Krüpteeritud sedel tuleb enne kogumisteenusesse talletamisele saatmist
digitaalselt allkirjastada, milleks on võimalik kasutada kõiki Eesti Vabariigis
kehtivaid digitaalallkirjavahendeid – ID-kaart, Digi-ID, Mobiil-ID.

Käesolev spetsifikatsioon näeb ette Eesti Vabariigi Standardikavandis [BDOC2.1]
defineeritud BDOC allkirjavormingu kasutamise. BDOC allkirjavorming koosneb ETSI
standardi TS 101 903 (XadES) profiilist ning OpenDocument konteineri vormingust.
IVXV protokollistik võimaldab ka alternatiivsete allkirja- ning
konteinervormingute kasutamist.

Olenevalt käimasoleval valimisel esitatud küsimuste arvust võib digitaalselt
allkirjastatud hääl sisaldada ühte või mitut andmefaili MIME-tüübiga
``application/octet-stream``. Iga andmefaili sisuks on krüpteeritud sedel.
Andmefaili ja teiste signeeritavate andmeobjektide räsimiseks enne
allkirjastamist kasutatakse räsifunktsiooni SHA-256. Andmefaili nimi
moodustatakse laiendist '``ballot``' ning valimise ja küsimuse
identifikaatorist. Kõik viidatud andmefailid peavad sisalduma
allkirjakonteineris. Digitaalselt allkirjastatud hääl ei tohi sisaldada muid
andmefaile kui neid, mis sisaldavad hääli mõne käimasoleva valimise kontekstis.
Seadistusele mittevastavate häälte vastuvõtmisest, talletamisest ja
töötlemisest peab kogumisteenus keelduma.


.. code-block:: bnf

    extension = "ballot"

    encrypted-ballot-name = election-identifier '.' question-identifier '.' extension


Valija poolt valijarakenduses allkirjastatud hääl moodustatakse nii, et on
võimalik selle edasine kvalifitseerimine kogumisteenuses. Käesolev
septsifikatsioon näeb ette hääle kvalifitseerimiseks nii OCSP kehtivuskinnituse
kui PKIX ajatempli võtmise. Sellisena on lõplik, kvalfitseeritud hääl, BDOC-TS
vormingus.

Kui hääl allkirjastatakse ID-kaardi või Digi-ID'ga, siis toimub algse
allkirjastatud konteineri moodustamine valijarakenduses. Kui hääl
allkirjastatakse Mobiil-ID'ga, siis toimub konteineri moodustamine
valijarakenduse ning kogumisteenuse poolt vahendatava DigiDoc-teenuse koostöös.
Mobiil-ID juhtumil kasutab kogumisteenus DigiDoc-teenust ainult signatuuri
saamiseks krüpteeritud sedelile. Kõik hääle kvalifitseerimiseks vajalikud
elemendid hangitakse vastavatelt teenustelt alles siis kui valijarakendus on
saatnud signeeritud hääle talletamiseks. Kvalifitseeritud hääl esitatakse
kogumisteenuse poolt valijarakendusele verifitseerimiseks, ainult
kvalifitseeritud hääl peab vastama BDOC 2.1 standardi tingimustele --
valijarakenduse poolt moodustatud hääl on vaheetapp kvalifitseeritud hääleni
jõudmiseks.

Valijarakenduses signeeritud häälel peab olema üks ja ainult üks allkiri, mida
hoitakse signatuurifailis :file:`META-INF/signature0.xml`. Häält ja allkirja
sisaldav konteiner moodustatakse BDOC 2.1 standardis kirjeldatud meetodit
kasutades.

Spetsifitseerime valijarakenduses allkirjastatud hääle vormingu ühe
küsimuse korral.

Räsialgoritmina ``DIGEST_ALG`` on kasutusel SHA-256
(http://www.w3.org/2001/04/xmlenc#sha256). XML kanoniseerimiseks
(``CANON_ALG``) kasutatakse meetodit c14n11
(http://www.w3.org/2006/12/xml-c14n11).

RSA võtmete korral (ID-kaart, Digi-ID) on allkirjastamismeetodiks
http://www.w3.org/2001/04/xmldsig-more#rsa-sha256. ECC võtmete korral
(ID-kaart, Mobiil-ID) http://www.w3.org/2001/04/xmldsig-more#ecdsa-sha256.

Identifikaatorite `VOTE_REF`, `SP_REF`, `SP_URI` ning `SV_URI` täpne väärtus ei
ole oluline.

Element `SignedProperties`
---------------------------

Element ``SignedProperties`` moodustatakse kooskõlas BDOC 2.1 standardiga. Kui
kvalifitseerimisel kasutatakse ajatemplit, siis elementi
``SignaturePolicyIdentifier`` ei kasutata. Ühtegi mitte-kohustuslikku elementi
ei kasutata. Allkirjastamise kellaaja fikseerib andmestruktuuri täitev arvuti
ning valija X509-sertifikaat saadakse kas ID-kaardilt või DigiDoc-teenuse
vahendusel.


.. literalinclude:: xmltemplates/sp.template
   :language: xml
   :linenos:


Element `SignedInfo`
---------------------

Element ``SignedInfo`` moodustatakse kooskõlas BDOC 2.1 standardiga, viidates
nii krüpteeritud sedelile (``VOTE_DIGEST``) kui elemendile ``SignedProperties``
(``SP_DIGEST``).

.. literalinclude:: xmltemplates/si.template
   :language: xml
   :linenos:


Element `SignatureValue`
------------------------

Element ``SignatureValue`` moodustatakse kooskõlas BDOC 2.1 standardiga.
Kanoniseeritud elemendist ``SignedInfo`` arvutatakse räsi, mis allkirjastatakse
PKCS1-meetodiga.

.. literalinclude:: xmltemplates/sv.template
   :language: xml
   :linenos:


Element `XAdESSignatures`
-------------------------

Element ``XAdESSignatures`` sisaldab ühte ``Signature`` elementi, mis on
koostatud lähtudes kõigist eelmistest elementidest ning valija X509
sertifikaadist. Elementi ``UnsignedProperties`` ei kasutata.

.. literalinclude:: xmltemplates/sig.template
   :language: xml
   :linenos:
