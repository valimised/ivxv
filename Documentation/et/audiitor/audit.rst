
================================================================================
IVXV juhend audiitorile
================================================================================

Rakenduste kompileerimine
================================================================================


EELDUSED
--------------------------------------------------------------------------------

Kõik tegevused viime läbi tavakasutaja -- näidetes `ubuntu` -- õigustes,
kasutades vajadusel käsklust `sudo`.

Üldjuhul viime kõik tegevused läbi kasutaja kodukaustas::

  cd $HOME

Paigaldame tarkvara, mis on vajalik audititööriistade ehitamiseks::

  sudo apt-get install --no-install-recommends -y autoconf automake build-essential libgmp-dev libtool git openjdk-11-jdk-headless python unzip zip wget make


INTCHECK
--------------------------------------------------------------------------------

Verificatum mixneti tervikluskontroll toimub rakendusega `intcheck`::

  wget https://github.com/vvk-ehk/intcheck/archive/master.zip
  unzip master.zip
  rm master.zip
  mv intcheck-master intcheck
  chmod +x intcheck/src/intcheck.py

Veendume et rakendus on paigaldatud korrektselt::

  ./intcheck/src/intcheck.py -h


JAVA RAKENDUSED
--------------------------------------------------------------------------------

Admeauditi läbiviimiseks läheb vaja rakendust `auditor`, mille lähtekood on
avalikustatud IVXV repositooriumis::

  wget https://github.com/vvk-ehk/ivxv/archive/master.zip
  unzip master.zip
  rm master.zip
  mv ivxv-master ivxv


Paigaldame Java sõltuvuspaketid::

  cd $HOME/ivxv/common/external
  wget -O gradle-6.4.zip https://services.gradle.org/distributions/gradle-6.4-bin.zip
  unzip gradle-6.4.zip
  rm gradle-6.4.zip
  cd $HOME/ivxv/common/java
  make sync

Veendume, et ettevalmistused on tehtud korrektselt::

  cd $HOME/ivxv
  make clean-java

Ehitame Java rakendused::

  make java

RVT'le tarnitavad rakendused::

  $HOME/ivxv/auditor/build/distributions/auditor-1.7.7.zip
  $HOME/ivxv/key/build/distributions/key-1.7.7.zip
  $HOME/ivxv/processor/build/distributions/processor-1.7.7.zip

Käivitatavad failid::

  $HOME/ivxv/auditor/build/install/auditor/bin/auditor
  $HOME/ivxv/key/build/install/key/bin/key
  $HOME/ivxv/processor/build/install/processor/bin/processor


VERIFICATUM
--------------------------------------------------------------------------------

Verificatum mixneti adapter on vajalik miksimistõendi kontrolliks::

  cd $HOME
  wget https://github.com/vvk-ehk/ivxv-mixnet-adapter/archive/master.zip
  unzip master.zip
  rm master.zip
  mv ivxv-mixnet-adapter-master ivxv-verificatum

Verificatum tarkvara allalaadmine::

  git clone https://github.com/verificatum/verificatum-gmpmee gmpmee
  git clone https://github.com/verificatum/vmgj
  git clone https://github.com/verificatum/vcr
  git clone https://github.com/verificatum/vmn

Tarkvara täpse versiooni hankimine ning tervikluse kontroll::

  cd gmpmee
  git checkout 4aafc31
  rm -rf .git/
  cd ..
  ./intcheck/src/intcheck.py verify gmpmee ivxv-verificatum/doc/gmpmee.dirsha256sum

  cd vmgj
  git checkout 8d7d412
  rm -rf .git/
  cd ..
  ./intcheck/src/intcheck.py verify vmgj ivxv-verificatum/doc/vmgj.dirsha256sum

  cd vcr
  git checkout af9fd82
  rm -rf .git/
  cd ..
  ./intcheck/src/intcheck.py verify vcr ivxv-verificatum/doc/vcr.dirsha256sum

  cd vmn
  git checkout bb00543
  rm -rf .git/
  cd ..
  ./intcheck/src/intcheck.py verify vmn ivxv-verificatum/doc/vmn.dirsha256sum

Verificatumi adapteri ehitamine::

  cd $HOME/ivxv-verificatum
  make zipext

Juhuarvugeneraatori initsialiseerimine Verificatumi jaoks::

  cd $HOME
  ./vcr/bin/vog -rndinit RandomDevice /dev/urandom


Auditeerimine
================================================================================

Siit edasi eeldame, et lugeja on tuttav dokumendiga "IVXV seadistuste
koostamise juhend" järgmises ulatuses:

* Ptk. 2, IVXV seadistused valimise korraldamise protsessis
* Ptk. 3, IVXV rakendused
* Ptk. 6, Auditirakendus
* Ptk. 10, E-häälte miksimine

Olgu samuti paigaldatud pakk `audit-examples.tar`, millel on järgmine
struktuur::

   audit-conv
   |-- auditor.yaml -- konfinäide
   |-- inputs
   |   |-- <RVT poolt tarnitavad sisendid>
   |-- process
   |   |-- <Töökataloog koos seadistustega>
   |
   audit-mix
   |-- auditor.yaml -- konfinäide
   |-- inputs
   |   |-- <RVT poolt tarnitavad sisendid>
   |-- process
   |   |-- <Töökataloog koos seadistustega>
   |
   audit-mixver
   |-- inputs
   |   |-- <RVT poolt tarnitavad sisendid>
   |
   audit-pdec
   |-- auditor.yaml -- konfinäide
   |-- inputs
   |   |-- <RVT poolt tarnitavad sisendid>
   |-- process
   |   |-- <Töökataloog koos seadistustega>
   |
   audit-vertally
   |-- inputs
   |   |-- <RVT poolt tarnitavad sisendid>
   |
   processor
   |-- <Töötlemisrakenduse sisendid ja väljundid>

Tegutsemine on üldjuhul järgmine:

* Tutvuge konfinäitega
* Veenduge, et on olemas vajalik RVT sisend
* Tehke kausta `process` konfinäitest lähtuv failistruktuur
* Käivitage kaustas `process` rakendus ja tööriist (eelvalmendatud konf on seal
  juba ees)

Täpsemad juhised järgnevad.

Genereeritud avalike võtmete kooskõlalisuse kontroll
--------------------------------------------------------------------------------

Võtmete genereerimise ajal tekib kaks võtit - tulemusfaili signeerimisvõti ja
häälte salastamise võti.

Tulemusfaili signeerimisvõti on kodeeritud X509 sertifikaadina failis
`RK2051-sign.pem`. Häälte salastamise võti on antud kolmes kodeeringus:

* X509 sertifikaadina failis `RK2051-enc.pem`
* DER-kodeeritud avaliku võtmena failis `RK2051-pub.der`
* PEM-kodeeritud avaliku võtmena failis `RK2051-pub.pem`

On võimalik kontrollida, et sertfikaat, mis sisaldab tulemusfaili
signeerimisvõtit, on korrektselt isesigneeritud. Seda saab teha järgnevalt::

    openssl verify -CAfile RK2051-sign.pem -check_ss_sig RK2051-sign.pem

Korrektse sertifikaadi korral on väljund::

    RK2051-sign.pem: OK

On võimalik kontrollida, et sertifikaat, mis sisaldab häälte salastamise võtit,
on korrektselt signeeritud tulemusfaili signeerimisvõtmega. Seda saab
teha järgnevalt::

    openssl verify -CAfile RK2051-sign.pem -check_ss_sig RK2051-enc.pem

Korrektselt allkirjastatud sertifikaadi korral on väljund::

    RK2051-enc.pem: OK

.. note:: Teadaoleva OpenSSL vea tõttu ei suuda OpenSSL versioonist 1.1.1b
   vanemad versioonid sertifikaadi usaldusahelat kontrollida. Eelneva kontrolli
   õnnestumise jaoks on eelduseks vähemalt OpenSSL versioon 1.1.1b.

Lisaks on võimalik kontrollida, et häälte salastamise võtme eri kodeeringud
vastavad üksteisele. Me kontrollime, et X509 sertifikaadis olev võti vastab
DER-kodeeritud võtmele ning lisaks, et PEM-kodeeritud võti vastab DER-kodeeritud
võtmele. Transitiivsuse tõttu on seega kõik kolm kodeeringut kooskõlalised.

Esiteks tuleb eraldada häälte salastamise võti vastavast sertifikaadist. Kuna
OpenSSL ei toeta kasutatavad ElGamali krüptoskeemi, siis tuleb avaliku võtme
eksportimiseks kasutada OpenSSL `asn1parse` tööriista.

Kõigepealt tuleb leida avaliku võtme nihe sertifikaadis::

    openssl asn1parse -in RK2051-enc.pem

Avalik võti on vastavas `SubjectPublicKeyInfo` väljal::

    156:d=2  hl=4 l= 816 cons: SEQUENCE
    160:d=3  hl=4 l= 415 cons: SEQUENCE
    164:d=4  hl=2 l=   9 prim: OBJECT            :1.3.6.1.4.1.3029.2.1
    175:d=4  hl=4 l= 400 cons: SEQUENCE
    179:d=5  hl=4 l= 385 prim: INTEGER           :FFFFFFFFFFFFFFFFC90FDA
        A22168C234C4C6628B80DC1CD129024E088A67CC74020BBEA63B139B22514A08
        798E3404DDEF9519B3CD3A431B302B0A6DF25F14374FE1356D6D51C245E485B5
        76625E7EC6F44C42E9A637ED6B0BFF5CB6F406B7EDEE386BFB5A899FA5AE9F24
        117C4B1FE649286651ECE45B3DC2007CB8A163BF0598DA48361C55D39A69163F
        A8FD24CF5F83655D23DCA3AD961C62F356208552BB9ED529077096966D670C35
        4E4ABC9804F1746C08CA18217C32905E462E36CE3BE39E772C180E86039B2783
        A2EC07A28FB5C55DF06F4C52C9DE2BCBF6955817183995497CEA956AE515D226
        1898FA051015728E5A8AAAC42DAD33170D04507A33A85521ABDF1CBA64ECFB85
        0458DBEF0A8AEA71575D060C7DB3970F85A6E1E4C7ABF5AE8CDB0933D71E8C94
        E04A25619DCEE3D2261AD2EE6BF12FFA06D98A0864D87602733EC86A64521F2B
        18177B200CBBE117577A615D6C770988C0BAD946E208E24FA074E5AB3143DB5B
        FCE0FD108E4B82D120A93AD2CAFFFFFFFFFFFFFFFF
    568:d=5  hl=2 l=   1 prim: INTEGER           :02
    571:d=5  hl=2 l=   6 prim: GENERALSTRING
    579:d=3  hl=4 l= 393 prim: BIT STRING

Näeme, et `SubjectPublicKeyInfo` välja  nihe on 156 baiti. Eraldame avaliku
võtme ja kontrollime vastavust väljastatud avaliku võtmega::

    openssl asn1parse -in RK2051-enc.pem -strparse 156 -noout -out extracted.der
    diff -s extracted.der RK2051-pub.der

Samaväärsete võtmete korral on väljundiks::

    Files extracted.der and RK2051-pub.der are identical

Teiseks kontrollime DER-kodeeritud võtme vastavust PEM-kodeeritud võtmele.
Selleks teisendame PEM-kodeeritud võtme DER-kodeeringusse ja võrdleme::

    openssl asn1parse -in RK2051-pub.pem -noout -out converted.der
    diff -s converted.der RK2051-pub.der

Samaväärsete võtme korral on väljundiks::

    Files converted.der and RK2051-pub.der are identical

Hääletamistulemuse allkirja verifitseerimine
--------------------------------------------------------------------------------

Nii tavalise dekrüpteerimise kui tõestatava dekrüpteerimise käigus tekib kaks
faili:

* Tulemusfail `RK2051.1.tally`
* Signatuurifail `RK2051.1.tally.signature`

Koos häälte salastamise võtmega genereeritakse tulemusfaili signeerimisvõti ja
vastav sertifikaat (`RK2051-sign.pem`). Dekrüpteeritud tulemusele antakse selle
võtmega signatuur, mida tuleb kontrollida.

Eraldame signeerimisvõtme sertifikaadist avaliku võtme::

  openssl x509 -in RK2051-sign.pem -noout -pubkey > sign.pub

Kasutame avalikku võtit tulemusfaili allkirja kontrollimiseks::

  openssl dgst -sha256 -sigopt rsa_padding_mode:pss -sigopt rsa_pss_saltlen:32 -sigopt rsa_mgf1_md:sha256 -verify sign.pub -signature RK2051.1.tally.signature RK2051.1.tally

NB! Tavalise dekrüpteerimise ja tõestatava dekrüpteerimise käigus tekkivad
tulemusfailid peavad olema identsed. Kontrollimiseks UNIXi tööriist `diff`::

  diff decout/RK2051.1.tally pdecout/RK2051.1.tally

Näitefailid on pakis::

  cd $HOME/audit-examples/audit-vertally


IVXV <-> Verificatum teisenduste korrektsuse kontroll
--------------------------------------------------------------------------------

Teisenduste korrektsuse kontroll toimub tööriistaga `convert`. NB! Kaust
`process` tuleb auditor.yaml põhjal sisenditest ettevalmistada::

  cd $HOME/audit-examples/audit-conv/process
  $HOME/ivxv/auditor/build/install/auditor/bin/auditor convert -c conf.bdoc -p auditor.yaml.bdoc

Miksimistõendi kontroll tööriistaga `auditor`
--------------------------------------------------------------------------------

Miksimistõendi kontroll toimub tööriistaga `mixer`. NB! Kaust `process` tuleb
auditor.yaml põhjal sisenditest ettevalmistada::

  cd $HOME/audit-examples/audit-mix/process
  $HOME/ivxv/auditor/build/install/auditor/bin/auditor mixer -c conf.bdoc -p auditor.yaml.bdoc

Lugemistõendi kontroll
--------------------------------------------------------------------------------

Lugemistõendi kontroll toimub tööriistaga `decrypt`. NB! Kaust `process` tuleb
auditor.yaml põhjal sisenditest ettevalmistada::

  cd $HOME/audit-examples/audit-pdec/process
  $HOME/ivxv/auditor/build/install/auditor/bin/auditor decrypt -c conf.bdoc -p auditor.yaml.bdoc

Miksimistõendi kontroll Verificatumi originaaltööriistaga
--------------------------------------------------------------------------------

Miksimistõendi kontroll Verificatumi abil::

  cd $HOME/audit-examples/audit-mixver
  $HOME/ivxv-verificatum/release/mixer/bin/mix.py verify --proof-zipfile shuffle_proof.zip

Töötlemise audit
--------------------------------------------------------------------------------

Täiendavalt on lisatud pakki kõik töötlemisrakenduse sisendid ja väljundid
lihtsustamaks töötlemisprotsessi auditit.
