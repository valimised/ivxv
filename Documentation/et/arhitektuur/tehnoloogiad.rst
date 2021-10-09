..  IVXV arhitektuur

.. _tehnoloogiad:

Kasutatavad tehnoloogiad
========================

Kogumisteenuse programmeerimiskeel
----------------------------------

Kogumisteenuse tuumikfunktsionaalsus on programmeeritud keeles Go, mis vastab
järgmistele hanke nõuetele:

* Staatiline tüüpimine;

* Automaatne mäluhaldus;

* Kompilaator avatud lähtekoodiga;

* Ribastamine (rööprapse).

Kogumisteenuse haldusteenus on programmeeritud keeles Python.


Rakenduste programmeerimiskeel
------------------------------

Rakendused on programmeeritud keeles Java, mis vastab hanke nõuetele keele laia
leviku ja jätkusuutlikkuse kohta.


Projekti sõltuvused
-------------------

Projektis kasutatavad kolmandate osapoolte komponendid koos nende motiveeritud
kasutamisvajadusega on üles loetletud järgnevates tabelites. Eraldi tabelid on
raamistiku pakendamiseks ja töötamiseks ning raamistiku arenduseks ja
testimiseks.

Kõik IVXV projektis kasutatavad välised teegid asuvad ``ivxv-external.git``
hoidlas või on saadaval platvormil, kus rakendus tööle hakkab.

Kõik kogumisteenuses kasutatavad komponendid on avatud lähtekoodiga.

.. tabularcolumns:: |p{0.4\linewidth}|p{0.1\linewidth}|p{0.25\linewidth}|p{0.25\linewidth}|
.. list-table::
   IVXV raamistiku tööks kasutatavad kolmandate osapoolte komponendid
   :header-rows: 1

   *  - Nimi
      - Versioon
      - Litsents (SPDX)
      - Kasutusvajadus

   *  - `Bootstrap <http://getbootstrap.com>`_
      - 3.4.1
      - MIT
      - Kogumisteenuse haldusteenuse kasutajaliidese kujundus

   *  - Bouncy Castle
      - 1.58
      - MIT
      - ASN1 käsitlemine, andmetüübi BigInteger abifunktsioonid

   *  - `Bottle <https://bottlepy.org/>`_
      - 0.12.15
      - MIT
      - Raamistik kogumisteenuse haldusteenuse veebiliidese teostamiseks

   *  - CAL10N
      - 0.7.7
      - MIT
      - Mitmekeelsuse tugi, tõlkefailide valideerimine

   *  - Digidoc 4j
      - 2.1.0
      - LGPL-2.1-only
      - BDoc konteinerite käsitlemine

   *  - Digidoc 4j DSS
      - 5.2.d4j.3
      - LGPL-2.1-only or EUPL-1.1
      - Digidoc 4j sõltuvus

   *  - Apache Commons (cli 1.4, codec 1.10, collections4 4.1, io 2.5, lang3 3.6, logging 1.2, compress 1.3)
      - -
      - Apache-2.0
      - Digidoc 4j ja PDFBox sõltuvused

   *  - Apache HttpComponents
      - 4.5.3
      - Apache-2.0
      - Digidoc 4j sõltuvus

   *  - Apache Santuario
      - 2.0.9
      - Apache-2.0
      - Digidoc 4j sõltuvus

   *  - Google Guava
      - 20.0
      - Apache-2.0
      - Digidoc 4j sõltuvus

   *  - JDigiDoc
      - 3.12.1
      - LGPL-2.1-only
      - Digidoc 4j sõltuvus

   *  - StaX
      - 1.0-2
      - Apache-2.0
      - Digidoc 4j sõltuvus

   *  - log4j
      - 1.2.6
      - Apache-2.0
      - Digidoc 4j sõltuvus

   *  - Woodstox
      - 4.4.1
      - Apache-2.0
      - Digidoc 4j sõltuvus

   *  - Xalan-Java
      - 2.7.2
      - Apache-2.0
      - Digidoc 4j sõltuvus

   *  - Xml Apis
      - 1.3.04
      - Apache-2.0
      - Digidoc 4j sõltuvus

   *  - `Docopt <http://docopt.org/>`_
      - 0.6.2
      - MIT
      - Kogumisteenuse haldusutiliitide käsurealiidese teostus

   *  - `Fasteners <https://github.com/harlowja/fasteners>`_
      - 0.14.1
      - Apache-2.0
      - Kogumisteenuse haldusteenuse protsesside lukustus

   *  - `etcd <https://coreos.com/etcd>`_
      - 3.2.26
      - Apache-2.0
      - Talletusteenusena kasutatav hajus võti-väärtus andmebaas

   *  - `github.com/golang/protobuf <https://github.com/golang/protobuf>`_
      - 1.3.2
      - BSD-3-Clause
      - etcd klientteegi sõltuvus

   *  - Glassfish JAXB
      - 2.3.3
      - BSD-3-Clause
      - Java XML teek

   *  - `google.golang.org/genproto <https://google.golang.org/genproto>`_
      - 58ce757
      - Apache-2.0
      - etcd klientteegi sõltuvus

   *  - `google.golang.org/grpc <https://google.golang.org/grpc>`_
      - 1.22.1
      - Apache-2.0
      - etcd klientteegi sõltuvus

   *  - `golang.org/x/net <https://golang.org/x/net>`_
      - 74dc4d7
      - BSD-3-Clause
      - etcd klientteegi sõltuvus

   *  - `golang.org/x/sys <https://golang.org/x/sys>`_
      - fc99dfb
      - BSD-3-Clause
      - etcd klientteegi sõltuvus

   *  - `golang.org/x/text <https://golang.org/x/text>`_
      - 0.3.2
      - BSD-3-Clause
      - etcd klientteegi sõltuvus

   *  - Gradle
      - 6.4
      - Apache-2.0
      - Java rakenduste ehitamise raamistik

   *  - `HAProxy <http://www.haproxy.org/>`_
      - 2.0.13
      - GPL-2.0-or-later
      - Vahendusteenusena kasutatav TCP-proksi

   *  - IvyPot
      - 0.12
      - Apache-2.0
      - Gradle ehitusraamistiku laiendus sõltuvuste haldamiseks ja rakenduste
        ehitamiseks vallasrežiimis

   *  - Jackson
      - 2.8.9
      - Apache-2.0
      - JSON vormingus failide lugemine ja kirjutamine

   *  - `jQuery <https://jquery.org/>`_
      - 3.3.1
      - MIT
      - Kogumisteenuse haldusteenuse kasutajaliides

   *  - Logback
      - 1.2.3
      - EPL-1.0 or LGPL-v2.1-only
      - Logimise API SLF4J teostus

   *  - Logback JSON
      - 0.1.5
      - EPL-1.0 or LGPL-v2.1-only
      - Logback logija laiendus JSON vormingus logikirjete koostamiseks
        Jackson teegi abil

   *  - `metisMenu <https://github.com/onokumus/metisMenu>`_
      - 1.1.3
      - MIT
      - Kogumisteenuse haldusteenuse kasutajaliides

   *  - PDFBox
      - 2.0.8
      - Apache-2.0
      - PDF vormingus raportite genereerimise tugi Java rakendustele

   *  - `PyYAML <http://pyyaml.org/>`_
      - 5.3.1
      - MIT
      - Kogumisteenuse seadistusfailide töötlemise tugi haldusteenusele

   *  - `Schematics <https://github.com/schematics/schematics>`_
      - 2.0.1
      - BSD-3-Clause
      - Kogumisteenuse seadistusfailide valideerimise tugi haldusteenusele

   *  - SLF4J
      - 1.7.25
      - MIT
      - Standardne logimise API

   *  - SnakeYAML
      - 1.18
      - Apache-2.0
      - YAML vormingus andmete lugemine

   *  - `SB Admin 2 <https://github.com/BlackrockDigital/startbootstrap-sb-admin-2>`_
      - 3.3.7+1
      - MIT
      - Kogumisteenuse haldusteenuse kasutajaliidese kujundus

.. list-table::
   IVXV raamistiku testide
   kasutatavad kolmandate osapoolte komponendid
   :header-rows: 1

   *  - Nimi
      - Versioon
      - Litsents (SPDX)
      - Kasutusvajadus

   *  - Hamcrest
      - 1.3
      - BSD-3-Clause
      - Loetavam assert-meetodite kasutamine Java üksuste testides

   *  - JUnit
      - 4.12
      - EPL-1.0
      - Java testimisraamistik

   *  - JUnitParams
      - 1.1.0
      - Apache-2.0
      - Testide parametriseerimise tugi

   *  - Mockito
      - 2.10.0
      - MIT
      - Testitava koodi sõltuvuste mockimise tugi

   *  - Byte Buddy
      - 1.9.10
      - Apache-2.0
      - Mockito sõltuvus

   *  - Objenesis
      - 2.6
      - Apache-2.0
      - Mockito sõltuvus

   *  - libdigidocpp-tools
      - 3.14.5.1404
      - LGPL-2.1-or-later
      - Testandmete genereerimine

   *  - PyTest
      - 6.2.3
      - MIT
      - Üksuste testimise tugi Pythonile

   *  - Requests
      - 2.25.1
      - Apache 2.0
      - HTTP päringute moodul Pythoni testidele

.. list-table::
   IVXV raamistiku arendamiseks ja/või testimiseks
   kasutatavad kolmandate osapoolte tööriistad
   :header-rows: 1

   *  - Nimi
      - Versioon
      - Litsents (SPDX)
      - Kasutusvajadus

   *  - `Behave <https://github.com/behave/behave>`_
      - 1.2.6
      - BSD-2-Clause
      - Regressioonitestide käivitaja (*Behavior-driven development*)

   *  - `Docker <http://www.docker.com/>`_
      - 18.06 (või uuem)
      - Apache-2.0
      - Regressioonitestide läbiviimise keskkond - tarkvarakonteinerid

   *  - `Docker Compose <http://www.docker.com/>`_
      - 1.28.6
      - Apache-2.0
      - Regressioonitestide läbiviimise keskkond - tarkvarakonteinerite haldus

   *  - `Sphinx <http://www.sphinx-doc.org/>`_
      - 3.4.3
      - BSD
      - Dokumentatsiooni genereerimine
