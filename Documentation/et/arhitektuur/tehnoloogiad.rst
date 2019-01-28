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

Kõik kogumisteenuses kasutatavad komponendid on avatud lähtekoodiga.

.. tabularcolumns:: |p{0.4\linewidth}|p{0.1\linewidth}|p{0.25\linewidth}|p{0.25\linewidth}|
.. list-table::
   IVXV raamistiku tööks kasutatavad kolmandate osapoolte komponendid
   :header-rows: 1

   *  - Nimi
      - Versioon
      - Litsents
      - Kasutusvajadus

   *  - `Bootstrap <http://getbootstrap.com>`_
      - 3.3.7
      - MIT
      - Kogumisteenuse haldusteenuse kasutajaliidese kujundus

   *  - Bouncy Castle
      - 1.58
      - MIT
      - ASN1 käsitlemine, andmetüübi BigInteger abifunktsioonid

   *  - `Bottle <https://bottlepy.org/>`_
      - 0.12.13
      - MIT
      - Raamistik kogumisteenuse haldusteenuse veebiliidese teostamiseks

   *  - CAL10N
      - 0.7.7
      - MIT
      - Mitmekeelsuse tugi, tõlkefailide valideerimine

   *  - Digidoc 4j
      - 2.1.0
      - LGPL
      - BDoc konteinerite käsitlemine

   *  - Digidoc 4j DSS
      - 5.2.d4j.3
      - LGPL
      - Digidoc 4j sõltuvus

   *  - Apache Commons (cli 1.4, codec 1.10, collections 4.1, io 2.5, lang 3.6, logging 1.2, compress 1.3)
      - -
      - Apache License v2.0
      - Digidoc 4j ja PDFBox sõltuvused

   *  - Apache HttpComponents
      - 4.5.3
      - Apache License v2.0
      - Digidoc 4j sõltuvus

   *  - Apache Santuario
      - 2.0.9
      - Apache License v2.0
      - Digidoc 4j sõltuvus

   *  - Google Guava
      - 20.0
      - Apache License v2.0
      - Digidoc 4j sõltuvus

   *  - JDigiDoc
      - 3.12.1
      - LGPL
      - Digidoc 4j sõltuvus

   *  - StaX
      - 1.0-2
      - Apache License v2.0
      - Digidoc 4j sõltuvus

   *  - log4j
      - 1.2.6
      - Apache License 2.0
      - Digidoc 4j sõltuvus

   *  - Woodstox
      - 4.4.1
      - Apache License 2.0
      - Digidoc 4j sõltuvus

   *  - Xalan-Java
      - 2.7.2
      - Apache License 2.0
      - Digidoc 4j sõltuvus

   *  - Xml Apis
      - 1.3.04
      - Apache License 2.0
      - Digidoc 4j sõltuvus

   *  - `Docopt <http://docopt.org/>`_
      - 0.6.2
      - MIT
      - Kogumisteenuse haldusutiliitide käsurealiidese teostus

   *  - `etcd <https://coreos.com/etcd>`_
      - 3.2.17
      - Apache License v2.0
      - Talletusteenusena kasutatav hajus võti-väärtus andmebaas

   *  - `github.com/ghodss/yaml <https://github.com/ghodss/yaml>`_
      - 73d445a
      - MIT
      - etcd klientteegi sõltuvus

   *  - `gopkg.in/yaml.v2 <https://gopkg.in/yaml.v2>`_
      - 4c78c97
      - Apache License v2.0
      - github.com/ghodss/yaml sõltuvus

   *  - `github.com/golang/protobuf <https://github.com/golang/protobuf>`_
      - 224aaba
      - BSD 2.0
      - etcd klientteegi sõltuvus

   *  - `github.com/grpc-ecosystem/go-grpc-prometheus <https://github.com/grpc-ecosystem/go-grpc-prometheus>`_
      - 6b7015e
      - Apache License v2.0
      - etcd klientteegi sõltuvus

   *  - `github.com/grpc-ecosystem/grpc-gateway <https://github.com/grpc-ecosystem/grpc-gateway>`_
      - 6863684
      - BSD 2.0
      - etcd klientteegi sõltuvus

   *  - `google.golang.org/grpc <https://google.golang.org/grpc>`_
      - 1.0.4
      - Apache License v2.0
      - etcd klientteegi sõltuvus

   *  - `golang.org/x/net <https://golang.org/x/net>`_
      - f249948
      - BSD 2.0
      - etcd klientteegi sõltuvus

   *  - `Prometheuse klientteek <https://prometheus.io>`_
      - 0.8
      - Apache License v2.0
      - etcd klientteegi sõltuvus

   *  - `github.com/beorn7/perks/quantile <https://github.com/beorn7/perks>`_
      - 4c0e845
      - MIT
      - Prometheuse klientteegi sõltuvus

   *  - `github.com/matttproud/golang_protobuf_extensions <https://github.com/matttproud/golang_protobuf_extensions>`_
      - 1.0.0
      - Apache License v2.0
      - Prometheuse klientteegi sõltuvus

   *  - Gradle
      - 3.0
      - Apache License v2.0
      - Java rakenduste ehitamise raamistik

   *  - `HAProxy <http://www.haproxy.org/>`_
      - 1.8.8
      - GPL v2
      - Vahendusteenusena kasutatav TCP-proksi

   *  - IvyPot
      - 0.4
      - Apache License v2.0
      - Gradle ehitusraamistiku laiendus sõltuvuste haldamiseks ja rakenduste
        ehitamiseks vallasrežiimis

   *  - Jackson
      - 2.8.9
      - Apache License v2.0
      - JSON vormingus failide lugemine ja kirjutamine

   *  - `jQuery <https://jquery.org/>`_
      - 3.1.0
      - MIT
      - Kogumisteenuse haldusteenuse kasutajaliides

   *  - Logback
      - 1.2.3
      - Eclipse Public License v1.0 või LGPL v2.1
      - Logimise API SLF4J realisatsioon

   *  - Logback JSON
      - 0.1.5
      - Eclipse Public License v1.0 või LGPL v2.1
      - Logback logija laiendus JSON vormingus logikirjete koostamiseks
        Jackson teegi abil

   *  - `metisMenu <https://github.com/onokumus/metisMenu>`_
      - 1.1.3
      - MIT
      - Kogumisteenuse haldusteenuse kasutajaliides

   *  - PDFBox
      - 2.0.8
      - Apache License v2.0
      - PDF vormingus raportite genereerimise tugi Java rakendustele

   *  - `PyYAML <http://pyyaml.org/>`_
      - 3.12
      - MIT
      - Kogumisteenuse seadistusfailide töötlemise tugi haldusteenusele

   *  - `Schematics <https://github.com/schematics/schematics>`_
      - 2.0.1
      - BSD
      - Kogumisteenuse seadistusfailide valideerimise tugi haldusteenusele

   *  - SLF4J
      - 1.7.25
      - MIT
      - Standardne logimise API

   *  - SnakeYAML
      - 1.18
      - Apache License v2.0
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
      - Litsents
      - Kasutusvajadus

   *  - Hamcrest
      - 1.3
      - BSD
      - Loetavam assert-meetodite kasutamine Java ühiktestides

   *  - JUnit
      - 4.12
      - Eclipse Public License v1.0
      - Java testimisraamistik

   *  - JUnitParams
      - 1.1.0
      - Apache License v2.0
      - Testide parametriseerimise tugi

   *  - Mockito
      - 2.10.0
      - MIT
      - Testitava koodi sõltuvuste mockimise tugi

   *  - Byte Buddy
      - 1.6.14
      - Apache License v2.0
      - Mockito sõltuvus

   *  - Objenesis
      - 2.5
      - Apache License v2.0
      - Mockito sõltuvus

   *  - libdigidoc2
      - 3.10.4.1218
      - LGPL
      - Testandmete genereerimine

   *  - libdigidocpp-tools
      - 3.13.6.1371
      - LGPL
      - Testandmete genereerimine

.. list-table::
   IVXV raamistiku arendamiseks ja/või testimiseks
   kasutatavad kolmandate osapoolte tööriistad
   :header-rows: 1

   *  - Nimi
      - Versioon
      - Litsents
      - Kasutusvajadus

   *  - `Behave <https://github.com/behave/behave>`_
      - 1.2.6
      - BSD
      - Regressioonitestide käivitaja (*Behavior-driven development*)

   *  - `Docker <http://www.docker.com/>`_
      - 18.06 (või uuem)
      - Apache License 2.0
      - Regressioonitestide läbiviimise keskkond - tarkvarakonteinerid

   *  - `Docker Compose <http://www.docker.com/>`_
      - 1.22.0
      - Apache License 2.0
      - Regressioonitestide läbiviimise keskkond - tarkvarakonteinerite haldus

   *  - `Sphinx <http://www.sphinx-doc.org/>`_
      - 1.8.2
      - BSD
      - Dokumentatsiooni genereerimine
