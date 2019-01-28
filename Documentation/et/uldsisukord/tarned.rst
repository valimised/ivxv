..  IVXV dokumentatsiooni üldsisukord

Tarned
======

Muudatused tarne 1.4.1 koosseisus, erinevused võrreldes tarnega 1.4.0
--------------------------------------------------------------------------------

Kogumisteenus
~~~~~~~~~~~~~

Üldised muutused

* Väiksemad muutused/veaparandused vastavalt `changelog` failile

Haldusteenus

* Väiksemad muutused/veaparandused vastavalt `changelog` failile

IVXV mikroteenused

* Väiksemad muutused/veaparandused vastavalt `changelog` failile

Registreerimisteenus
~~~~~~~~~~~~~~~~~~~~

* Muudatusi ei ole

Kontrollrakendused
~~~~~~~~~~~~~~~~~~

Android

* Muudatusi ei ole

iOS

* Muudatusi ei ole

Mixnet
~~~~~~

* Muudatusi ei ole

Töötleja rakendused
~~~~~~~~~~~~~~~~~~~

Üldised

* Muudatusi ei ole

Võtmerakendus

* Tööriista *init* väljundfailide muutus
* Tööriista *init* SN ja CN parameetrite muutus
* Tööriista *testkey* sisendparameetrite muutus
* Tööriista *decrypt* kasutatud kaardinumbrite kuvamine

Töötlemisrakendus

* Muudatusi ei ole

Auditirakendus

* Muudatusi ei ole

Valijarakendused ja seadistusrakendus
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

* MSAA toe täpsustamine

Dokumentatsioon
~~~~~~~~~~~~~~~

* IVXV protokollid - registreerimisprotokolli lisamine
* IVXV seadistuste koostejuhend - läbivad täiendused ning käsiraamatuga
  ühtlustamine.
* IVXV valijarakendus - läbivad täiendused.
* IVXV mixnet - dokument hõlmatud seadistuste koostejuhendisse, eemaldatud.
* IVXV registreerimisteenus - dokument hõlmatud protokollistikku, eemaldatud.

Logimonitor
~~~~~~~~~~~

* Väiksemad muutused/veaparandused vastavalt `changelog` failile

Muudatused tarne 1.4.0 koosseisus, erinevused võrreldes tarnega 1.3.0
--------------------------------------------------------------------------------

Kogumisteenus
~~~~~~~~~~~~~

Üldised muutused

* OCSP ja ajatemplipäringute automaatne kordamine vastavalt
  konfiguratsioonile
* BDOC-TS allkirjakonteinerite toetamine
* Tühjade valijanimekirjade tugi

Haldusteenus

* Väiksemad muutused/veaparandused vastavalt `changelog` failile

IVXV mikroteenused

* Väiksemad muutused/veaparandused vastavalt `changelog` failile

Registreerimisteenus
~~~~~~~~~~~~~~~~~~~~
* Muudatusi ei ole

Kontrollrakendused
~~~~~~~~~~~~~~~~~~

Android

* TLS 1.2 toetamine API-versioonide < 19 korral
* Abiinfo vaate asendamine süsteemse brauseriga
* Kuvast väljuvate tekstide automaatne lühendamine nuppude korral

iOS

* Muudatusi ei ole

Mixnet
~~~~~~

* Muudatusi ei ole

Töötleja rakendused
~~~~~~~~~~~~~~~~~~~

Üldised

* Jaoskonnanumbri unikaalsusnõuete täpsustamine
* Tühja valijatenimekirja tugi
* Kodeeringuvigadega sertifikaatide toetamine rakendustes

Võtmerakendus

* Eemaldatud LOG4 ja LOG5

Töötlemisrakendus

* Eemaldatud PDF vormingus valijate nimekiri faasist *revoke*

Auditirakendus

* Muudatusi ei ole

Valijarakendused ja seadistusrakendus
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

* macOS 10.11 toetamine
* 32bit Linuxi toetamine
* UPX versiooni uuendamine
* Pinpad kaardilugejate töökindluse tõstmine (Win)
* ID-kaardi suhtlusvigade parandamine (Win7/ECC)
* Nii PEM kui DER vormingus andmete import (Seadistaja)
* Lisatud Seadistuste valideerimisvõimalus (Seadistaja)
* Parandatud käitumine liigsuurte seadistuste korral (Seadistaja)

Dokumentatsioon
~~~~~~~~~~~~~~~

* Dokumentatsiooni uuendamine muudatuste kajastamiseks ning DEMO2018
  tagasiside arvestamiseks

Logimonitor
~~~~~~~~~~~

* MTA sõltuvuse lisamine
* CSV logiväljavõtte täpsustamine, algus- ja lõpuaja lisamine
* Logianalüüsi optimeerimine mitmetuumalise riistvara jaoks
* Väiksemad muutused/veaparandused vastavalt `changelog` failile

Muudatused tarne 1.3.0 koosseisus, erinevused võrreldes tarnega 1.2.0
--------------------------------------------------------------------------------

Kogumisteenus
~~~~~~~~~~~~~

Üldised muutused

* Ubuntu 18.04 LTS (Bionic Beaver) kasutuselevõtmine
* Krahhitaaste protseduuride kirjeldamine

Haldusteenus

* Parandatud tööriistad teenuste seisundiinfo saamiseks
* Väiksemad muutused/veaparandused vastavalt `changelog` failile

IVXV mikroteenused

* golang keeleversioon 1.9 kasutuselevõtmine
* Eesti ID-kaardi uuenenud profiili toetamine (PNOEE)
* Väiksemad muutused/veaparandused vastavalt `changelog` failile

Registreerimisteenus
~~~~~~~~~~~~~~~~~~~~
* Muudatusi ei ole

Kontrollrakendused
~~~~~~~~~~~~~~~~~~

Android

* ESTEID2018 sertifikaatide toetamine
* Täpsustatud vigase ASN1-kodeeringuga avalike võtmete käitlemist

iOS

* ESTEID2018 sertifikaatide toetamine
* iPhone 10 X muudatused
* XCode 10 ja iOS 12 SDK kasutamine

Mixnet
~~~~~~

* Verificatumi AGPL versiooni kasutuselevõtmine

Töötleja rakendused
~~~~~~~~~~~~~~~~~~~

Võtmerakendus

* Muudatusi ei ole

Töötlemisrakendus

* Lisatud tööriist StatsTool urnist statistikafaili genereerimiseks
* Lisatud tööriist StatsDiffTool kahe statistikafaili võrdlemiseks
* ESTEID2018 sertifikaatide ja profiili toetamine
* digidoc4j 2.1.0 kasutamine

Auditirakendus

* Muudatusi ei ole

Valijarakendused ja seadistusrakendus
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

* Valijarakenduses kandidaatide otsingu võimaldamine
* Valijarakenduses erakondade ja kandidaatide kaustana kuvamine
* Valijarakenduse ja Seadistusrakenduse üleviimine JSON-vormingus seadistustele
* ESTEID2018 sertifikaatide ja profiili toetamine
* Win: IDEMIA minidraiveri toetamine
* Linux/macOS: IDEMIA PKCS11 draiveri toetamine
* macOS 10.14 toetamine

Dokumentatsioon
~~~~~~~~~~~~~~~

* Dokumentatsiooni uuendamine muudatuste kajastamiseks
* Ingliskeelse arhitekuuridokumendi ja protokollistiku lisamine

Logimonitor
~~~~~~~~~~~

* Ubuntu 18.04 LTS (Bionic Beaver) kasutuselevõtmine
* Võetud kasutusele Grafana 5.3.4
* Parandatud vanusepõhise statistika genereerimine ja vanusegruppidesse
  jaotumine
* CSV väljundi võtmine seanssidest

Muudatused tarne 1.2.0 koosseisus, erinevused võrreldes KOV2017 valimistega.
--------------------------------------------------------------------------------

Kogumisteenus
~~~~~~~~~~~~~

Haldusteenus

* Lisatud tööriist vigaste valijanimekirjade eemaldamiseks.
* Lisatud tööriist jaoskondade/ringkondade nimekirja lisamiseks.
* Lisatud ringkonnapõhine statistika.
* Lisatud varundusteenus.
* Lisatud tööriist varundatud urnide konsolideerimiseks.
* Lisatud tööriist nimekirjade kooskõlalisuse kontrolliks.
* Lisatud võimekus seadistusfailidele valimisspetsiifiliste prefiksite lisamiseks.
* Täiendatud haldusliidese kasutajaliidest abinfoga.
* Eemaldatud aegunud konfiguratsiooniparameeter “stats.*”
* Parandatud sisendfailide vormingu kontrolli ja laadimist.

IVXV mikroteenused

* Uuendatud etcd versioon.
* Lisatud võimekus etcd ajalõppude seadistamiseks keskkonnamuutujate kaudu.
* Lisatud võimekus klastri modiftseerimiseks krahhitaaste eesmärgil.
* Parandatud klastri käitumist liidrivahetuse korral, pooleliolevate talletamiste kordamine.
* TLS šifrid muudetud seadistatavaks.
* Parandatud BDOC profiili identifitseeriva konfiguratsioonivälja nimi.
* Lisatud võimekus seadistada Mobiil-ID autentimist nõudma nii isikukoodi kui telefoninumbrit.
* Lisatud võimekus piirata korduvhääletamise sagedust.
* Lisatud võimekus toetada Windowsi reavahetusi konfifailides.
* Täiustatud BDOC XML kanoniseerimist ja parsimist.
* Karmistatud DDS päringute vormingukontrolle.
* Logimine viidud üle RELP protokollile.
* Muudetud seadistusfailide ülesehitust eristamaks Koguja ja Töötleja vastutusi.

Registreerimisteenus
~~~~~~~~~~~~~~~~~~~~
* Muudatusi ei ole.

Kontrollrakendused
~~~~~~~~~~~~~~~~~~

Android

* Lisatud juhised publitseeritud kontrollrakenduse ja avalikustatud lähtekoodi vastavuse kontrollimiseks.

iOS

* Muudatusi ei ole.

Mixnet
~~~~~~
* Muudatusi ei ole.

Töötleja rakendused
~~~~~~~~~~~~~~~~~~~
Võtmerakendus

* Muudatusi ei ole.

Töötlemisrakendus

* Muudatusi ei ole.

Auditirakendus

* Muudatusi ei ole.

Valijarakendused ja seadistusrakendus
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
* Linux ja Mac – platvormispetsiifilisi muudatusi ei ole.
* Windows – parandatud liidestumist Minidraiveriga, mingw64 kasutuselevõtmine.
* Lisatud ID-kaardi ECC toetamine.
* Seadistatud Mobiil-ID nõudma vajadusel isikukoodi ja telefoninumbrit.
* Täiendatud veakoode.
* Kohandatud seadistusrakendus muudatustega vastavusse.
* Kohandatud BDOC XML templated

Dokumentatsioon
~~~~~~~~~~~~~~~
* Dokumentatsioon läbivalt kaasajastatud seoses muudatustega

Logimonitor
~~~~~~~~~~~
* Loobutud CrateDBst.
* Võetud läbivalt kasutusele PostgreSQL.
* Võetud kasutusele Grafana 5.0.1.
* Seansside valideerimise parandused lähtudes KOV2017 logianalüüsist.
* Lisatud statistika genereerimine ringkondade kaupa.
