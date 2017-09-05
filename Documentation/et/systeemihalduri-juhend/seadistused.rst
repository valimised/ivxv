..  IVXV kogumisteenuse haldusjuhend

Kogumisteenuse seadistused
==========================

Logimise seadistused
--------------------

Kogumisteenuse logi hoitakse logi tekkimise asukohas ja dubleeritakse
logiserveritesse. Logide kogumiseks ja edastamiseks kasutatakse vaikimisi
*syslog*-teenust ``rsyslog``.

Kogumisteenus toetab kahte liiki logservereid, mis
kirjeldatakse kogumisteenuse tehnilises seadistuses.

#. Kogumisteenuse logikogumisteenus on kogumisteenuse sisemine teenus ja seda
   võib süsteemis olla mitu isendit.

#. Tegevusmonitooringu server on kogumisteenuse jaoks
   väline teenus ja seda võib olla ainult üks isend.

Kogumisteenusele tehniliste seadistuse rakendamisel paigaldab haldusteenus
logikogumisteenuse(d) enne teise teenuseid, et teenuste poolt toodetav logi
saaks võimalikult varakult ka logikogumisteenusesse kogutud.

.. note::

   Kogumisteenuse logiteated tekivad pärast valimiste seadistuse esmakordset
   laadimist, kuna teenused käivitatakse selle seadistuse laadimise järel.

Logi tootva teenuse logimise korraldus
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Logi tootva teenuse logimise seadistuse genereerib haldusteenus vastavalt
tehnilistele seadistustele.

#. Iga teenus logib kohalikku *syslog*-teenusesse;

#. Kõigi teenusmasinate *syslog*-teenused on seadistatud kogumisteenuse logi
   salvestama kohalikku failisüsteemi (:file:`/var/log/ivxv.log`);

#. Kõigi teenusmasinate (peale logikogumisteenuse) *syslog*-teenused on
   seadistatud edastama üle võrgu:

   #. Kõiki logikirjeid logikogumisteenusesse (protokoll: RELP);

   #. Kogumisteenuse logikirjeid tegevuslogi monitooringu serverisse
      (protokoll: UDP);


Logikogumisteenuse korraldus
^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Logikogumisteenuse seadistusfail tuleb teenuse tarkvarapakist
(:ref:`ivxv-logcollector.conf <ivxv-logcollector.conf>`).

#. Logikogumisteenus võtab logikirjeid vastu RELP-protokolli kaudu;

#. Kogumisteenuse logkikirjed salvestatakse JSON-vormingus faili
   :file:`/var/log/ivxv.log` (välja arvatud silumislogi);

#. Kogumisteenuse silumislogi ja teiste oluliste teenuste (haproxy, rsyslog,
   sshd) logi kirjutatakse rsyslogi standardvormingus faili
   :file:`/var/log/ivxv-debug.log`.

.. vim: sts=3 sw=3 et:
