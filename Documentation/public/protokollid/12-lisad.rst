..  IVXV protokollid

=====
Lisad
=====

Andmestruktuuride ASN.1 spetsifikatsioon
========================================

.. literalinclude:: ivxv-elgamal-general.asn1
   :language: asn1
   :name: asn-general
   :linenos:
   :caption: IVXV ElGamal üldised andmestruktuurid

.. literalinclude:: ivxv-elgamal-modp.asn1
   :language: asn1
   :name: asn-modp
   :linenos:
   :caption: IVXV ElGamal ModP spetsiifilised andmestruktuurid

.. literalinclude:: ivxv-elgamal-ecc.asn1
   :language: asn1
   :name: asn-ecc
   :linenos:
   :caption: IVXV ElGamal ECC spetsiifilised andmestruktuurid



Elektroonilise hääle vormingu spetsifikatsioon
==============================================

Käesoleva protokolliversiooniga kooskõlaline ettepanek VVK otsuseks
hääletamissedeli ja elektroonilise hääle vormi kehtestamiseks.

.. literalinclude:: paragraph.txt


Vead töötlemisprotsessis
========================

Anname ülevaate veakoodidest töötlemisprotsessis. Veakoodid kirjeldavad vigu
üksiku elemendi verifitseerimisel - hääl, registreerimistõend,
registreerimispäring - ning häälte, registreerimistõendite ja
registreerimispäringute vahelise vastavuse loomisel.

Enamus veaolukordi on pigem hüpoteetilised (nt. `REG_NO_NONCE`), siiski on
tegemist olukordadega, mis programselt võivad esineda ning seetõttu tuleb neid
ka käsitleda.

Veaolukord tähendab, et konkreetne element ei liigu edasi töötlemise järgmisesse faasi.

.. list-table:: Töötlemisprotsessi veakoodid
   :widths: 40 60
   :header-rows: 1

   * - Veakood
     - Selgitus
   * - ``INVALID_FILE_NAME``
     - Failinimi ei vasta oodatud mustrile. Viitab tundmatule failile urnis.
   * - ``MISSING_FILE``
     - Mõni hääle koosseisuks vajalik fail puudub. Nt. ajatempel.
   * - ``REPEATED_FILE``
     - Mõnda hääle koosseisuks vajalikku faili on mitu. Nt. kehtivuskinnitus.
   * - ``UNKNOWN_FILE_TYPE``
     - Faili tüüp on tundmatu või mittetoetatud.
   * - ``INVALID_FILE_SIZE``
     - Allkirjastatud hääle faili suurus ei vasta seadistustes nõutud kriteeriumitele.
   * - ``INVALID_BALLOT_SIGNATURE``
     - Hääl on vigase allkirjaga.
   * - ``MISSING_VOTER_SIGNATURE``
     - Hääl ei sisalda valija allkirja.
   * - ``VOTER_NOT_FOUND``
     - Valija ei olnud hääletamise hetkel valijate nimekirjas.
   * - ``VOTERLIST_NOT_FOUND``
     - Versioonile vastavat valijate nimekirja ei leitud.
   * - ``TIME_BEFORE_START``
     - Hääl on antud enne hääletamisperioodi algust. Viitab testhäälele.
   * - ``REG_RESP_INVALID``
     - Registreerimistõend/ajatempel on vigane. Viga ATOs või töötlejas.
   * - ``REG_REQ_INVALID``
     - Registreerimispäring on vigane. Viga kogumisteenuses või töötlejas.
   * - ``REG_RESP_NOT_UNIQUE``
     - Registreerimistõend ei ole unikaalne. Viga kogumisteenuses või töötlejas.
   * - ``REG_REQ_NOT_UNIQUE``
     - Registreerimispäring ei ole unikaalne. Sama räsiga häält on esitatud korduvalt.
   * - ``REG_NO_NONCE``
     - Registreerimistõendis puudub nonss.
   * - ``REG_NONCE_NOT_SIG``
     - Esitatud nonss ei ole IVXV protokolli kohaselt allkirjastatud.
   * - ``REG_NONCE_ALG_MISMATCH``
     - Nonssi allkirjastamisel kasutatud algoritm ei vasta oodatule.
   * - ``REG_NONCE_SIG_INVALID``
     - Nonssi allkiri on vigane.
   * - ``UNKNOWN_FILE_IN_VOTE_CONTAINER``
     - Hääle konteineris on tundmatu fail.
   * - ``TECHNICAL_ERROR``
     - Töötlemise ajal tekkis tehniline viga.
   * - ``REG_RESP_REQ_UNMATCH``
     - Registreerimistõendi andmed ei vasta registreerimispäringu andmetele.
   * - ``REG_REQ_WITHOUT_BALLOT``
     - Esitatud on registreerimispäring, kuid hääl puudub urnist. Kogumisteenuse viga.
   * - ``BALLOT_WITHOUT_REG_REQ``
     - Hääl on esitatud ilma vastava registreerimispäringuta. ATO viga.
   * - ``SAME_TIME_AS_LATEST``
     - Kaks häält samalt valijalt võivad olla arvesse võetavad viimasena.
   * - ``INVALID_SIGNATURE_PROFILE``
     - Hääle allkirjaprofiil on vigane.


Töötlemise käigus tuvastatakse kehtetud sedelid sellisel määral, mil seda
võimaldavad kontrollid avaliku võtmega. Kuigi kogumisteenus ei luba kehtetuid
sedeleid talletada, peab töötlemisrakendus kontrolle siiski kordama tagamaks
mh. võimalike tarkvaravigade vältimise.


.. list-table:: Krüptogrammide kehtivuse kontrolli veakoodid
   :widths: 30 70
   :header-rows: 1

   * - Veakood
     - Selgitus
   * - ``INVALID_BYTES``
     - Baidijada ei ole dekodeeritav ElGamalCiphertext'na
   * - ``INVALID_GROUP``
     - Väärtus ei ole oodatud rühma element
   * - ``INVALID_RANGE``
     - Väärtus on lubatud vahemikust väljas.
   * - ``INVALID_QR``
     - Väärtus ei ole ruutjääk (MODP).
   * - ``INVALID_POINT``
     - Väärtus ei ole kõvera punkt (ECC).
   * - ``INVALID``
     - Vigane šifreeritud tekst.

