..  IVXV kogumisteenuse haldusjuhend

Ülevaade
========

Kogumisteenuse ülevaade
-----------------------

IVXV kogumisteenus on elektroonilise
hääletuse käigus hääletajate teenindamiseks ja häälte kogumiseks mõeldud
tarkvara.

Kogumisteenus koosneb mikroteenustest ja nende haldamiseks mõeldud
haldusteenusest. Haldusteenuse kasutamine on käsureapõhine. Osade funktsioonide
kasutamist on laiendatud veebipõhise liidesega, mida on kirjeldatud dokumendis
``IVXV kogumisteenuse haldusliidese
kasutusjuhend``.

.. attention::

   Kogumisteenus paigaldatakse ja seadistatakse eraldi iga hääletuse
   läbiviimiseks. Ühe kogumisteenusega on korraga võimalik teenindada ainult
   ühte hääletust.


Lisamaterjalid
--------------

Käesolevas dokumendis kasutatakse mõisteid ja definitsioone, mis on kirjeldatud
dokumendis ``IVXV-ÜK-0.95 Elektroonilise hääletamise üldraamistik ja selle
kasutamine Eesti riiklikel valimistel``:

* E-hääletamise etapid;

* Süsteemi osapooled ja komponendid.


Kogumisteenuse kasutajate rollid
--------------------------------

Kogumisteenuses on kasutusel järgnevad rollid:

#. **Kogumisteenuse haldur** tegeleb kogumisteenuse tehnilise haldamisega;

#. **Valimiste haldur** tegeleb valimiste seadistuste kehtestamisega;

#. **Vaataja** pääseb ligi haldusteenuse kaudu väljastatavatele seisundi- ja
   statistikaandmetele;

Rollide täpsem kirjeldus asub dokumendis ``Elektroonilise hääletamise
infosüsteemi IVXV seadistuste koostamise juhend``.


Süsteemi komponendid
--------------------

Kogumisteenus
^^^^^^^^^^^^^

**Haldusteenus** on kogumisteenuse haldamise teenus. Haldusteenuse kaudu
juhitakse ja jälgitakse kogumisteenust alates paigaldusest kuni mahavõtmiseni.
Vaata lähemalt lõigus :ref:`haldusteenus`.

**Logikoguja** on kogumisteenuse sisemine logiserver, mis kogub ja säilitab
kõigi kogumisteenuste alamteenuste logisid. Logikogujasse kogutud logid antakse
valimiste lõppedes üle korraldajale.

**Sisemine varundus** on kogumisteenuse varundusteenus, mis varundab kõigi
alamteenuste andmeid ja teeb need lihtsa liidese (failisüsteemi kataloog) kaudu
kättesaadavaks välisele varundusteenusele.

**Alamteenused** on kogumisteenuse eri lõikude eest vastutavad teenused.


.. _tugiteenused:

Tugiteenused
^^^^^^^^^^^^

**Logiseire** on kogumisteenuse logide analüüsiks ja jälgimiseks mõeldud
seireprogramm.

**Tehniline seire** on kogumisteenuse tehnilise toimimise jälgimiseks mõeldud
seireprogramm.

**Väline varundus** on kogumisteenuse sisemisest varunduse poolt varundatud
andmete säilitamiseks mõeldud väline varundusteenus.


.. _välisteenused:

Välised teenused
^^^^^^^^^^^^^^^^

Välised teenused on läbiviidavatele valimistele kehtestatud nõuetest sõltuvad
teenused, millega kogumisteenus on võimeline liidestuma. Väliste teenuste hulka
kuuluvad Registreerimisteenus, Ajatempliteenus, Mobiil-ID
teenus, OCSP teenus vms.


Ülevaade toimingutest
---------------------

* Hääletamiseelsel etapil:

   * Kirjeldatakse kogumisteenuse poolt kasutatavad :ref:`välised teenused
     <välisteenused>`;

   * Valmistatakse ette kogumisteenuse :ref:`tugiteenused <tugiteenused>`;

   * Koostatakse kogumisteenuse seadistused (usaldusjuur, tehnilised
     seadistused ja valimiste seadistused);

   * Genereeritakse teenuse toimimiseks vajalikud krüptovõtmed ja
     sertifikaadid;

   * Valmistatakse ette kogumisteenuse käitamiseks vajalik taristu;

   * Paigaldatakse haldusteenus;

   * Rakendatakse seadistused haldusteenusele, mille põhjal haldusteenus
     paigaldab ja seadistab kogumisteenuse alamteenused.

* Hääletamisetapil

   * Jälgitakse teenuse toimimist;

   * Luuakse e-valimiskastist varukoopiaid.

* Töötlusetapil

   * Eksporditakse kogumisteenusesse kogutud andmed:

      #. konsolideeritud e-valimiskast kogutud häältega.

* Lugemisetapil

   * Lugemisetapil kogumisteenust ei kasutata;
