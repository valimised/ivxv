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

.. tabularcolumns:: |p{0.2\linewidth}|p{0.1\linewidth}|p{0.15\linewidth}|p{0.55\linewidth}|
.. list-table::
   IVXV raamistiku tööks kasutatavad kolmandate osapoolte komponendid
   :header-rows: 1

   *  - Nimi
      - Versioon
      - Litsents (SPDX)
      - Kasutusvajadus

   *  - `Bootstrap <http://getbootstrap.com>`_
      - 3.4.1, JavaScript
      - MIT
      - Kogumisteenuse haldusteenuse kasutajaliidese kujundus

   *  - Bouncy Castle
      - 1.78.1, Java
      - MIT
      - ASN1 käsitlemine, andmetüübi BigInteger abifunktsioonid

   *  - `Bottle <https://bottlepy.org/>`_
      - 0.13.2, Python
      - MIT
      - Raamistik kogumisteenuse haldusteenuse veebiliidese teostamiseks

   *  - CAL10N
      - 0.8.1, Java
      - MIT
      - Mitmekeelsuse tugi, tõlkefailide valideerimine

   *  - Digidoc 4j
      - 5.3.1, Java
      - LGPL-2.1-only
      - BDoc konteinerite käsitlemine

   *  - Apache Commons (collections4 4.4)
      - Java
      - Apache-2.0
      - Digidoc 4j ja PDFBox sõltuvused

   *  - `Docopt <http://docopt.org/>`_
      - 0.6.2, Python
      - MIT
      - Kogumisteenuse haldusutiliitide käsurealiidese teostus

   *  - `Fasteners <https://github.com/harlowja/fasteners>`_
      - 0.19, Python
      - Apache-2.0
      - Kogumisteenuse haldusteenuse protsesside lukustus

   *  - `gin-gonic <https://github.com/gin-gonic>`_
      - 1.9.1, Go
      - MIT
      - Veebiraamistik x-tee liidese jaoks

   *  - `etcd <https://coreos.com/etcd>`_
      - 3.5.9, Go
      - Apache-2.0
      - Talletusteenusena kasutatav hajus võti-väärtus andmebaas

   *  - Glassfish JAXB
      - 4.0.5, Java
      - BSD-3-Clause
      - Java XML teek

   *  - Gradle
      - 8.11, Java
      - Apache-2.0
      - Java rakenduste ehitamise raamistik

   *  - `HAProxy <http://www.haproxy.org/>`_
      - 2.4.24
      - GPL-2.0-or-later
      - Vahendusteenusena kasutatav TCP-proksi

   *  - Jackson
      - 2.18.1, Java
      - Apache-2.0
      - JSON vormingus failide lugemine ja kirjutamine

   *  - Jinja2
      - 3.1.4, Python
      - BSD
      - Jinja mallide kasutamine haldusteenuses

   *  - `jQuery <https://jquery.org/>`_
      - 3.7.1, JavaScript
      - MIT
      - Kogumisteenuse haldusteenuse kasutajaliides

   *  - jsonschema
      - 4.23.0, Python
      - MIT
      - JSON valideerimine haldusteenuses

   *  - Logback
      - 1.5.12, Java
      - EPL-1.0 or LGPL-v2.1-only
      - Logimise API teostus

   *  - Logback JSON
      - 0.1.5, Java
      - EPL-1.0 or LGPL-v2.1-only
      - Logback logija laiendus JSON vormingus logikirjete koostamiseks
        Jackson teegi abil

   *  - `Logrus <https://github.com/sirupsen/logrus>`_
      - 1.9.3, Go
      - MIT
      - Logimisraamistik x-tee liidese jaoks

   *  - `metisMenu <https://github.com/onokumus/metisMenu>`_
      - 1.1.3, JavaScript
      - MIT
      - Kogumisteenuse haldusteenuse kasutajaliides

   *  - `FontAwesome <https://github.com/FortAwesome/Font-Awesome>`_
      - 6.7.2, JavaScript
      - MIT
      - Kogumisteenuse haldusteenuse kasutajaliides

   *  - `DataTables <https://github.com/DataTables/DataTablesSrc>`_
      - 2.3.2, JavaScript
      - MIT
      - Kogumisteenuse haldusteenuse kasutajaliides

   *  - PDFBox
      - 3.0.3, Java
      - Apache-2.0
      - PDF vormingus raportite genereerimise tugi Java rakendustele

   *  - `PyYAML <http://pyyaml.org/>`_
      - 6.0.2, Python
      - MIT
      - Kogumisteenuse seadistusfailide töötlemise tugi haldusteenusele

   *  - python-crontab
      - 3.3.0, Python
      - LGPLv3
      - Crontab haldusteenuses

   *  - python-dateutil
      - 2.9.0, Python
      - BSD
      - Kuupäevad ja kellaajad haldusteenuses

   *  - python-debian
      - 0.1.49, Python
      - GPLv2
      - Debian pakkide lugemine haldusteenuses

   *  - pyopenssl
      - 24.2.1, Python
      - Apache
      - OpenSSL kasutus haldusteenuses

   *  - `Schematics <https://github.com/schematics/schematics>`_
      - 2.1.1, Python
      - BSD-3-Clause
      - Kogumisteenuse seadistusfailide valideerimise tugi haldusteenusele

   *  - SnakeYAML
      - 2.3, Java
      - Apache-2.0
      - YAML vormingus andmete lugemine

   *  - `SB Admin 2 <https://github.com/BlackrockDigital/startbootstrap-sb-admin-2>`_
      - 3.3.7+1, JavaScript
      - MIT
      - Kogumisteenuse haldusteenuse kasutajaliidese kujundus

.. tabularcolumns:: |p{0.2\linewidth}|p{0.1\linewidth}|p{0.15\linewidth}|p{0.55\linewidth}|
.. list-table::
   IVXV raamistiku testide
   kasutatavad kolmandate osapoolte komponendid
   :header-rows: 1

   *  - Nimi
      - Versioon
      - Litsents (SPDX)
      - Kasutusvajadus

   *  - Hamcrest
      - 3.0, Java
      - BSD-3-Clause
      - Loetavam assert-meetodite kasutamine Java üksuste testides

   *  - JUnit
      - 5.10.0, Java
      - EPL-1.0
      - Java testimisraamistik

   *  - JUnitParams
      - 1.1.1, Java
      - Apache-2.0
      - Testide parametriseerimise tugi

   *  - Mockito
      - 5.14.2, Java
      - MIT
      - Testitava koodi sõltuvuste mockimise tugi

   *  - libdigidocpp-tools
      - 3.14.5 .1404
      - LGPL-2.1-or-later
      - Testandmete genereerimine

   *  - PyTest
      - 7.4.2, Python
      - MIT
      - Üksuste testimise tugi Pythonile

   *  - Requests
      - 2.32.3, Python
      - Apache 2.0
      - HTTP päringute moodul Pythoni testidele

.. tabularcolumns:: |p{0.2\linewidth}|p{0.1\linewidth}|p{0.15\linewidth}|p{0.55\linewidth}|
.. list-table::
   IVXV raamistiku arendamiseks ja/või testimiseks
   kasutatavad kolmandate osapoolte tööriistad
   :header-rows: 1

   *  - Nimi
      - Versioon
      - Litsents (SPDX)
      - Kasutusvajadus

   *  - `Behave <https://github.com/behave/behave>`_
      - 1.2.6, Python
      - BSD-2-Clause
      - Regressioonitestide käivitaja (*Behavior-driven development*)

   *  - `Docker <http://www.docker.com/>`_
      - 18.06 (või uuem)
      - Apache-2.0
      - Regressioonitestide läbiviimise keskkond - tarkvarakonteinerid

   *  - `Sphinx <http://www.sphinx-doc.org/>`_
      - 7.2.5, Python
      - BSD
      - Dokumentatsiooni genereerimine
