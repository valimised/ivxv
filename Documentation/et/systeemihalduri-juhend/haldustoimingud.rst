..  IVXV kogumisteenuse haldusjuhend

Süsteemi haldustoimingud
========================

.. _kogumisteenuse-oleku-jälgimine:

Kogumisteenuse oleku jälgimine
------------------------------

Kogumisteenuse olekuandmed registreeritakse haldusteenuse andmebaasis. Oleku
kuvamiseks on utiliit :ref:`ivxv-status`.

Olekus kuvatakse järgmisi andmeid:

* Valimise ID, faas, algus- ja lõpuaeg;

* Laaditud seadistuspakkide versioonid;

* Laaditud valimisnimekirjade versioonid;

* Teenuste nimekiri koos rakendatud seadistuste versioonidega, teenuse seisundi
  ja selle viimase tuvastamise ajaga;

* Väliste teenuste seisundid;

* Haldusteenuse andmehoidla statistika.

Sõltuvalt kogumisteenuse seisundist võib oleku kuvamise utiliit jätta mõned
andmeblokid kuvamata (kui need pole jooksva seisundi puhul olulised). Täieliku
andmestiku väljastamiseks vaata utiliidi :ref:`ivxv-status` abiteavet.

Mikroteenuste oleku jälgimise ning oleku ja võimaliku veainfo registreerimisega
haldusteenuse andmebaasis tegeleb haldusteenuse :ref:`agentdeemon
<ivxv-agent-daemon>`.

Kogumisteenuse haldusteenuse sündmuste logi kuvamiseks on utiliit
:ref:`ivxv-eventlog-dump`.

.. important::

   Haldusteenus tagab kogumisteenuse alamteenuste olekuandmetes vajaliku teabe
   teenuse töökorda seadmiseks. See võib olla järgmine:

   #. Teave puuduvate seadistuste kohta (seadistusfailid, võtmed jms). Seda
      kuvatakse kuni teenus on varustatud kõigi käivitamiseks vajalike
      seadistustega.

   #. Veateade - alamteenuse haldusvahendite (seadistuste kontrollivahend,
      teenuse haldusvahend) veaväljund mittetöötava teenuse kohta.


.. _korralduste-valideerimine:

Korralduste valideerimine
-------------------------

Korraldusfailide valideerimine võimaldab veenduda korralduste vastavuses
vormistusnõuetele ning tuvastada vigased või mittekooskõlalised korraldused.

Valideerimine toimub käsuga :ref:`ivxv-config-validate`.

Valimise seadistuse valideerimise näide::

   $ ivxv-config-validate --election=valimise-seadistus-TEST2017.bdoc


Korralduste kooskõlalisuse kontrollid
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Kooskõla kontrollid viiakse läbi kahel juhul:

#. Kui valideerimise käsule antakse korraga valideerimiseks mitu korraldust;

#. Mõni kooskõla seisukohalt oluline korraldus juba haldusteenusesse laaditud.

Korraga ringkondade ja valikute nimekirja kontrollides viiakse läbi järgnevad
kooskõla kontrollid:

* Igas valimisringkonnas peab olema kirjeldatud vähemalt üks valik;

* Iga valik paeb olema seotud valimisringkonnaga.

Korraga ringkondade ja valijate nimekirja kontrollides viiakse läbi järgnevad
kooskõla kontrollid:

* Igas valimisjaoskonnas peab olema vähemalt üks valija;

* Iga valimisnimekirja kantud isik peab olema seotud olemasoleva ringkonnaga.


.. _korralduste-laadimine-rakendamine:

Korralduste laadimine ja rakendamine
------------------------------------

Kogumisteenuse korraldused koostatakse signeeritud korralduspakkidena, millega
kirjeldatakse kasutaja identifikaator (*Common Name* ehk CN väli ID-kaardilt)
ja rollide nimekiri.

Sõltuvalt korraldusest tuleb rakendamiseks kasutada ühte või kahte käsku.
Haldusteenust puudutavad korralduste rakendamiseks piisab nende laadimisest
haldusteenusesse. Alamteenuseid puudutavad korraldused (näiteks seadistuspakid)
tuleb pärast haldusteenusesse laadimist rakendada ka hallatavatele teenustele.

Korralduste laadimine haldusteenusesse toimub käsuga :ref:`ivxv-cmd-load`.
Laadimise käigus viiakse läbi ka :ref:`korralduse valideerimine
<korralduste-valideerimine>`, vigane või mittekooskõlaline korraldus jäetakse
laadimata.

Valikute nimekirja korralduse rakendamise näide:

.. include:: genereeritud-failid/haldusteenus-laadi_valikute_nimekiri.inc

.. seealso::

   * Käsu :ref:`ivxv-cmd-load` abiteave;

   * Korralduste rollide kirjeldus ja korralduste koostamise juhend asuvad
     dokumendis ``IVXV seadistuste
     koostamise juhend``.


Korralduste eemaldamine
-----------------------

Vigase korralduse eemaldamine on võimalik ainult valijate nimekirjade korral,
mida pole veel hallatavatele teenustele rakendatud. Eemaldamine toimub käsuga
:ref:`ivxv-cmd-remove`.

Valijate nimekirja eemaldamise näide:

.. include:: genereeritud-failid/haldusteenus-eemalda-valijate-nimekiri.inc

.. seealso::

   * Käsu :ref:`ivxv-cmd-remove` abiteave.


Teenuse isendi seisundi tuvastamine
-----------------------------------

Mikroteenuse isendi seisundi tuvastamiseks on utiliit :ref:`ivxv-service`,
millega on võimalik teenuse seisundit vahetult küsida (utiliit
:ref:`ivxv-status` kuvab andmebaasis puhverdatavat seisundit).

Teenuse seisundi päringu näide:

.. include:: genereeritud-failid/mikroteenuse_seisundi_tuvastamine.inc


.. _teenuse-taaskäivitamine:

Teenuse (taas)käivitamine
-------------------------

Mikroteenuste käivitamiseks ja taaskäivitamiseks on utiliit
:ref:`ivxv-service`.

Teenuse taaskäivitamise näide:

.. include:: genereeritud-failid/mikroteenuse_kaivitamine.inc

.. note::

   Protseduuri nimetamine käivitamiseks või taaskäivitamiseks sõltub teenuse
   protsessi seisundist. Tehniliselt on tegemist sarnaste protseduuridega, kus
   esmalt veendutakse, et teenus seisab (vajadusel jäetakse see seisma) ja siis
   püütakse käivitatada hetkel kehtivate seadistustega.


.. _teenuse-seiskamine:

Teenuse seiskamine
------------------

Mikroteenuste seiskamiseks on utiliit :ref:`ivxv-service`.

Teenuse seiskamise näide:

.. include:: genereeritud-failid/mikroteenuse_seiskamine.inc


.. _teenuse-asendamine:

Teenuse isendi asendamine
-------------------------

Teenuse isendi asendamine koosneb ühe mikroteenuse isendi eemaldamisest (vt.
:ref:`teenuse-eemaldamine`) ja teise sama funktsiooniga mikroteenuse isendi
lisamisest (vt. :ref:`teenuse-lisamine`).


.. _teenuse-lisamine:

Teenuse isendi lisamine
-----------------------

Teenuse isendi lisamiseks tuleb vajadusel teenust hostiv server ette valmistada
(vt.  :ref:`taristu-paigaldamine`) ning rakendada uus tehniline seadistus, mis
sisaldab lisatavat teenuse isendit.

.. important::

   Lisatava isendi identifikaator ei tohi kattuda ühegi teise, ka minevikus
   eemaldatud isendi identifikaatoriga.


.. _teenuse-eemaldamine:

Teenuse isendi eemaldamine
--------------------------

Teenuse isendi eemaldamiseks tuleb:

#. Teenuse isend seisma jätta (vt. :ref:`teenuse-seiskamine`);

#. Keelata teenuse isendi uuesti käivitamine (vt. allpool);

#. Rakendada uus tehniline seadistus, mis eemaldatavat isendit enam ei sisalda.

.. important::

   Teenuse isendi eemaldamisel kogumisteenuse koosseisust on oluline
   eemaldatava isendi täielik elimineerimine.

   Teenuste isendid kasutavad üksteisele usalduse tõestamiseks kindla
   sertifitseerimiskeskuse (CA) poolt välja antud sertifikaate, kuid ei kasuta
   sama meetodit eemaldatud isendi usalduse tühistamiseks (vastava protseduuri
   rakendamise liigse keerukuse tõttu).

   Seetõttu on oluline veenduda, et kogumisteenusest eemaldatud teenuse isend
   on enne uue seadistuse rakendamist täielikult süsteemist eemaldatud.
   Vastasel juhul tekib oht, et eemaldatav isend jätkab tegutsemist ja häirib
   kogumisteenuse tööd.

Teenuse isendi käivitamise keelamiseks teenuse eemaldamisel tuleb eemaldada
vastava teenuse tarkvarapakk teenuse hostist:

* Nimekirjateenuse paki eemaldamine:

   .. code-block:: text

      $ apt purge ivxv-choices

* Mobiil-ID tugiteenuse paki eemaldamine:

   .. code-block:: shell-session

      $ apt purge ivxv-dds

* Vahendusteenuse paki eemaldamine:

   .. code-block:: shell-session

      $ apt purge haproxy

* Talletusteenuse paki eemaldamine:

   .. code-block:: shell-session

      $ apt purge etcd-server

* Kontrolliteenuse paki eemaldamine:

   .. code-block:: shell-session

      $ apt purge ivxv-verification

* Hääletamisteenuse paki eemaldamine:

   .. code-block:: shell-session

      $ apt purge ivxv-voting


Kasutajate haldus
-----------------

Kasutajate algsed kirjeldused määratakse usaldusjuure seadistuses, hilisem
haldus toimub vastavate korralduste abil.

Kasutajate halduse korraldused rakendatakse käsuga :ref:`ivxv-cmd-load` (vaata
:ref:`korralduste-laadimine-rakendamine`).

Kasutajaõiguste määramise korralduse rakendamise näide:

.. include:: genereeritud-failid/kasutaja_lisamine.inc

.. attention::

   Juba lisatud kasutajate eemaldamine süsteemist pole võimalik. Kasutaja
   eemaldamise asemel tuleb kasutaja rolliks määrata "õigusteta kasutaja".

.. seealso::

   * Kasutajate rollide kirjeldus ja volituste korralduste koostamise juhend
     asuvad dokumendis ``IVXV seadistuste koostamise juhend``.

   * Korralduste rakendamine on kirjeldatud lõigus
     :ref:`korralduste-laadimine-rakendamine`.


Tarkvarauuenduste rakendamine
-----------------------------

Tarkvarauuendused jagunevad kogumisteenuse vaatepunktist kaheks:
operatsioonisüsteemi uuendused ja kogumisteenuse uuendused.

Operatsioonisüsteemi tarkvarapakkide uute versioonide paigaldamine pole
kogumisteenuse dokumentatsioonis käsitletud. Süsteemiülem peab tagama
ajakohaste turvauuenduste rakendamise kogumisteenuses kasutatavate
operatsioonisüsteemidele;

Kogumisteenuse tarkvarapakkide uute versioonide paigaldamine toimub järgnevalt:

#. Uuenenud tarkvarapakid kopeeritakse haldusteenuse kataloogi
   :file:`/etc/ivxv/debs` (soovitavalt juurkasutaja õigustes);

#. Haldusteenuse tarkvara uuendatakse juurkasutaja õigustes käsuga
   :command:`dpkg -i /etc/ivxv/debs/ivxv-common_1.0_all.deb
   /etv/ivxv/debs/ivxv-admin_1.0_amd64.deb` (tegelik versiooninumber erineb
   käesolevas näites kasutatud versioonist);

#. Hallatavate teenuste tarkvara uuendamine toimub haldusteenuse kasutaja
   ``ivxv-admin`` õigustes käsuga :ref:`ivxv-update-packages`.


Varundamine
-----------

Varundamine hõlmab kolme liiki andmeid:

#. Haldusteenuse seadistused;

#. Kogumisteenuse e-urn;

#. Kogutud logid.

Varukoopia loomine toimub haldusteenuse masinas utiliidi
:ref:`ivxv-backup` abil, varukoopiad talletatakse varundusserveri
kataloogis :file:`/var/backups/ivxv`.


Haldusteenus seadistuste varundamine
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Haldusteenuse seadistustest varundatakse järgmised andmed:

#. :file:`etc/` - haldusteenusesse laadiditud tarkvarapakid ja hetkel kehtivad
   seadistusfailid;

#. :file:`admin-ui-permissions/` - haldusteenuse kasutajaliidese pääsuõigused;

#. :file:`commands/` - kõik haldusteenusese laaditud korraldusfailid.

Haldusteenuse varundamist viiakse läbi haldusteenuses, varundatavad andmed
kopeeritakse varundusserverisse.

.. hint::

   Haldusteenuse andmete tõhusamaks varundamiseks ja taasteks on soovitav
   kasutada haldusteenuse virtuaalmasina dünaamilist tõmmist (*snapshot dump*).

Haldusteenuse seadistuste varukoopiast taastamise protseduuri pole
kogumisteenuses ette nähtud.

Haldusteenuse seadistusete varukoopia loomise näide::

   $ ivxv-backup management-conf


Kogumisteenuse e-urni varundamine
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Kogumisteenuse e-urni varundamine toimub talletusteenuses kogutud häältest
e-urni loomisega ja selle kopeerimisega varundusteenusesse. Varundatud andmete
taastamine toimub hääletuse järel e-urni väljastamise käigus, kus
talletusteenuses olevatest häältest ja varukoopiatesse salvestatud häältest
pannakse kokku töötlemisele minev e-urn.

E-urni varundamist viib läbi haldusteenus. Varukoopia loomine toimub
talletusteenuses ja see kopeeritakse varundusserverisse.

Varukoopia on sama vorminguga, nagu kogumisteenuse poolt väljastatav e-urn
(ZIP64).

Varundamise andmemahtu saab arvutada järgmise meetodiga:
``häälte arv * 12,1 kB * pakkimistegur``.

Näiteks saja tuhande hääle suurus, kus pakkimistegur on 0,4 = 472 MB.

E-urni varukoopia loomise näide::

   $ ivxv-backup ballot-box


Logide varundamine
^^^^^^^^^^^^^^^^^^

Logikogumisteenustes kogutud logifailide varundamine toimub logifaili
:file:`/var/log/ivxv.log` kopeerimisega varundusserverisse. Logide varundamist
viib läbi haldusteenus.

Logide varukoopiast taastamise protseduuri pole kogumisteenuses ette nähtud.

Logikogumisteenusesse kogutud logist varukoopia loomise näide::

   $ ivxv-backup log


Konsolideeritud e-urni koostamine
---------------------------------

Konsolideeritud e-urn koostatakse talletusteenusesse kogutud häältest ja
varundusteenusesse varundatud e-urnidest. Konsolideerimise protsess koosneb
järgmistest sammudest:

#. Talletusteenusesse kogutud hääled varundatakse varundusteenusesse. Selle
   tulemusena on varundusteenusesse salvestatud kõik kogutud e-urnid;

#. Varundusteenuses koostatakse konsolideeritud e-urn;

#. Konsolideeritud e-urn kopeeritakse haldusteenusesse.

Konsolideeritud e-urni koostamise näide:

.. include:: genereeritud-failid/e-urni_koostamine.inc
