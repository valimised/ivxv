
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

  sudo apt-get install --no-install-recommends -y autoconf automake build-essential libgmp-dev libtool git openjdk-8-jdk-headless python unzip zip wget make gradle

Kuna Java rakendused eeldavad Java versiooni 8, siis seadistame vaikeversiooni::

  sudo update-alternatives --config java

Paigaldame tekstiredaktori::

  sudo apt-get install vim


INTCHECK
--------------------------------------------------------------------------------

Verificatum mixneti tervikluskontroll toimub rakendusega `intcheck`::

  wget https://github.com/vvk-ehk/intcheck/archive/master.zip
  mv intcheck-master intcheck
  rm master.zip
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
  wget -O gradle-3.0.zip https://services.gradle.org/distributions/gradle-3.0-bin.zip
  unzip gradle-3.0.zip
  rm gradle-3.0.zip
  cd $HOME/ivxv/common/java
  $HOME/ivxv/common/external/gradle-3.0/bin/gradle syncRemoteRepositories

Kaks pakki tuleb paigaldada käsitsi::

  cd $HOME/ivxv/common/external/commons-codec/commons-codec/1.10
  wget -O commons-codec-1.10.jar https://search.maven.org/remotecontent?filepath=commons-codec/commons-codec/1.10/commons-codec-1.10.jar
  cd $HOME/ivxv/common/external/xml-apis/xml-apis/1.3.04
  wget -O xml-apis-1.3.04.jar https://search.maven.org/remotecontent?filepath=xml-apis/xml-apis/1.3.04/xml-apis-1.3.04.jar

Veendume, et ettevalmistused on tehtud korrektselt::

  make clean-java

Ehitame Java rakendused::

  make java

RVT'le tarnitavad rakendused::

  $HOME/ivxv/auditor/build/distributions/auditor-1.5.0.zip
  $HOME/ivxv/key/build/distributions/key-1.5.0.zip
  $HOME/ivxv/processor/build/distributions/processor-1.5.0.zip

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
  git checkout d781e4a
  rm -rf .git/
  cd ..
  ./intcheck/src/intcheck.py verify gmpmee ivxv-verificatum/doc/gmpmee.dirsha256sum

  cd vmgj
  git checkout 82b57dd
  rm -rf .git/
  cd ..
  ./intcheck/src/intcheck.py verify vmgj ivxv-verificatum/doc/vmgj.dirsha256sum

  cd vcr
  git checkout 6dba049
  rm -rf .git/
  cd ..
  ./intcheck/src/intcheck.py verify vcr ivxv-verificatum/doc/vcr.dirsha256sum

  cd vmn
  git checkout 2a0719e
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
   |   |-- RVT sisendid
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
