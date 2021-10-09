..  IVXV kogumisteenuse haldusjuhend

Kogumisteenuse seadistused
==========================

Logimise seadistused
--------------------

Kogumisteenuse logi hoitakse logi tekkimise asukohas ja dubleeritakse
logiserveritesse. Logide kogumiseks ja edastamiseks kasutatakse vaikimisi
*syslog*-teenust ``rsyslog``.

Kogumisteenus toetab kahte liiki logiservereid, mis
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
   salvestama kohalikku failisüsteemi (:file:`/var/log/ivxv-YYYY-MM-DD.log`);

#. Kõigi teenusmasinate (peale logikogumisteenuse) *syslog*-teenused on
   seadistatud edastama üle võrgu:

   #. Kõiki logikirjeid logikogumisteenusesse (protokoll: RELP);

   #. Kogumisteenuse logikirjeid tegevuslogi monitooringu serverisse
      (protokoll: RELP);


Logikogumisteenuse korraldus
^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Logikogumisteenuse seadistusfail tuleb teenuse tarkvarapakist
(:ref:`ivxv-logcollector.conf <ivxv-logcollector.conf>`).

#. Logikogumisteenus võtab logikirjeid vastu RELP-protokolli kaudu;

#. Kogumisteenuse logikirjeid kirjutatakse JSON-vormingus faili
   :file:`/var/log/ivxv-YYYY-MM-DD.log` (välja arvatud päringu- ja
   silumislogi);

#. Kogumisteenuse päringulogi kirjutatakse rsyslogi standardvormingus faili
   :file:`/var/log/ivxv-request-YYYY-MM-DD.log`;

#. Kogumisteenuse silumislogi ja teiste oluliste teenuste (haproxy, etcd,
   rsyslog, sshd) logi kirjutatakse rsyslogi standardvormingus faili
   :file:`/var/log/ivxv-debug-YYYY-MM-DD.log`.


Talletamisteenuse seadistused
-----------------------------

Hetkel ainus talletamisteenuse teostus kasutab hajusat võti-väärtus andmebaasi
``etcd``. Korraga käivitatakse mitu ``etcd`` isendit, mis saavutavad omavahel
konsensuse talletatud andmete osas.

Talletusteenuse sujuvaks tööks võib olla vajalik osade ``etcd`` parameetrite
häälestamine konkreetse evituskeskkonna jaoks. Selleks tuleb teenuse masinas
luua fail :file:`/etc/default/ivxv` ning sinna lisada järgmistes jaotistes
kirjeldatud read. Pärast faili loomist või selle sisu muutmist tuleb uute
väärtuste rakendamiseks talletusteenus taaskäivitada. Parameetri puudumise
korral kasutatakse vaikeväärtust.

Seadistuste väärtuste valimisel on abiks ``etcd`` dokumentatsioon aadressil
https://coreos.com/etcd/docs/latest/tuning.html.

Ajaparameetrid
^^^^^^^^^^^^^^

``etcd`` klaster valib ühe liikmetest juhiks, mis koordineerib kõiki
andmemuudatusi. Lisaks pingib juht perioodiliselt kõiki ülejäänud klastri
liikmeid aitamaks tuvastada olukorda, kus ühendus juhiga on katkenud: kui mõni
klastri liikmetest pole piisavalt kaua ühtegi pingi saanud, algatab see uue
juhi valimise.

Suurema võrgu- või kettalatentsuse tagajärjel võib juhi ping liialt viibida
ning põhjustada uue juhi valimise. Tõrgete korral on juhivahetus süsteemi
loomulik osa, ent töötava süsteemi puhul tarbetu koormus. Seetõttu tuleks
seadistada juhi pingimise tihedust ``ETCD_HEARTBEAT_INTERVAL`` ning teiste
liikmete ooteaega ``ETCD_ELECTION_TIMEOUT`` vastavalt evituskeskkonna
latentsusele::

   ETCD_HEARTBEAT_INTERVAL=100
   ETCD_ELECTION_TIMEOUT=1000

Mõlemad väärtused on millisekundites ning vaikimisi vastavalt 100ms ja 1000ms.

.. important::

   Ühes klastris peavad kõigil talletamisteenuse isenditel olema samad
   ajaparameetrid. Vastasel korral võib esineda stabiilsusprobleeme erinevate
   pingi ootuste tõttu.


Hetkvõtete parameetrid
^^^^^^^^^^^^^^^^^^^^^^

``etcd`` peab logi kõigist andmemuudatustest. Vältimaks logi liiga suureks
kasvamist tehakse andmebaasi seisust perioodiliselt hetkvõtteid ning eelnev
logi kustutatakse. Kui talletamisteenus kasutab liiga palju mälu või
kettaruumi, siis võib aidata tihedam hetkvõtete tegemine.

Uus hetkvõte tehakse iga ``ETCD_SNAPSHOT_COUNT`` andmemuudatuse järel, seega
madalam väärtus toob kaasa tihedamad hetkvõtted ning väiksema logi suuruse::

   ETCD_SNAPSHOT_COUNT=10000

Vaikimisi tehakse hetkvõte iga 10000 muudatuse järel.
