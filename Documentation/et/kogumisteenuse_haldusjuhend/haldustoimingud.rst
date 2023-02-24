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

* Haldusteenusesse laaditud konfiguratsioon:

   * Seadistuspakkide versioonid;

   * Valikute nimekirja ja ringkondade nimekirja versioonid;

   * Valijate nimekirjade versioonid ja olekud;

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

Valijate muudatusnimekirjade hankimine Valimiste Infosüsteemist toimub
utiliidiga :ref:`ivxv-voter-list-download`, mis käivitatakse teenuse `cron`
poolt veerandtunnise intervalliga.

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

   $ ivxv-config-validate --election=valimise-seadistus-TEST2017.asice


Korralduste kooskõlalisuse valideerimine
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Kooskõla valideerimine viiakse läbi kahel juhul:

#. Kui valideerimise käsule antakse korraga valideerimiseks mitu korraldust;

#. Korralduse laadimisel juhul, kui laaditava korraldusega kooskõla nõudev
   oluline korraldus juba haldusteenusesse laaditud.

Korralduste kooskõla valideerimise kontrollid:

#. Korraga valimiste seadistust ja/või nimekirju (valikute, ringkondade või
   valijate) valideerides kontrollitakse valimiste identifikaatori kooskõla.

#. Korraga mitut valijate nimekirja valideerides viiakse läbi järgnevad
   kontrollid:

   * Nimekirjade korrektne järjestus;

   *  Muudatusnimekirjas kontrollitakse:

      #. Valija topeltlisamist ja -eemaldamist;

      #. Valija eemaldamist pärast tema lisamist sama muudatusnimekirjaga.

      #. Valija eemaldamist ringkonnast, kuhu teda pole lisatud;

#. Korraga ringkondade ja valikute nimekirja valideerides viiakse läbi
   järgnevad kontrollid:

   * Igas valimisringkonnas peab olema kirjeldatud vähemalt üks valik;

   * Iga valik peab olema seotud olemasoleva valimisringkonnaga.

#. Korraga ringkondade nimekirja ja valijate nimekirju valideerides viiakse
   läbi järgnevad kontrollid:

   * Igas valimisjaoskonnas peab olema vähemalt üks valija;

   * Iga valimisnimekirja kantud isik peab olema seotud olemasoleva
     ringkonnaga;

#. Valijate nimekirja(de) ja ringkondade nimekirja valideerimisel
   kontrollitakse valijale määratud ringkonna olemasolu ringkondade nimekirjas.

Kui valimiste seadistuses on määratud välisriigis asuvale valijale määratav
ringkonna haldusüksuse EHAK-kood (parameeter ``voterforeignehak``), peab
valikute, valijate ja ringkondade nimekirjade vastavuse valideerimisel olema
kaasatud ka valimiste seadistus ja seadistuste valideerimisel tehakse täiendavad
kontrollid:

#. Valijate nimekirja ja ringkondade nimekirja kooskõla valideerides
   kontrollitakse, et ringkondade nimekirjas on parameetriga määratud
   haldusüksuses olemas valijale määratud ringkond.


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

      $ apt purge ivxv-mid

* Smart-ID tugiteenuse paki eemaldamine:

   .. code-block:: shell-session

      $ apt purge ivxv-smartid

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

#. Kogumisteenuse e-valimiskast;

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


Kogumisteenuse e-valimiskasti varundamine
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Kogumisteenuse e-valimiskasti varundamine toimub talletusteenuses kogutud häältest
e-valimiskasti loomisega ja selle kopeerimisega varundusteenusesse. Varundatud andmete
taastamine toimub hääletuse järel e-valimiskasti väljastamise käigus, kus
talletusteenuses olevatest häältest ja varukoopiatesse salvestatud häältest
pannakse kokku töötlemisele minev e-valimiskast.

E-valimiskasti varundamist viib läbi haldusteenus. Varukoopia loomine toimub
talletusteenuses ja see kopeeritakse varundusserverisse.

Varukoopia on sama vorminguga, nagu kogumisteenuse poolt väljastatav e-valimiskast
(ZIP64).

Varundamise andmemahtu saab arvutada järgmise meetodiga:
``häälte arv * 12,1 kB * pakkimistegur``.

Näiteks saja tuhande hääle suurus, kus pakkimistegur on 0,4 = 472 MB.

E-valimiskasti varukoopia loomise näide::

   $ ivxv-backup ballot-box


Logide varundamine
^^^^^^^^^^^^^^^^^^

Logikogumisteenustes kogutud logifailide varundamine toimub logifailide
:file:`/var/log/ivxv-YYYY-MM-DD.log` kopeerimisega varundusserverisse. Logide
varundamist viib läbi haldusteenus.

Logide varukoopiast taastamise protseduuri pole kogumisteenuses ette nähtud.

Logikogumisteenusesse kogutud logist varukoopia loomise näide::

   $ ivxv-backup log


.. _konsolideeritud-e-valimiskasti-koostamine:

Konsolideeritud e-valimiskasti koostamine
-----------------------------------------

Konsolideeritud e-valimiskast koostatakse talletusteenusesse kogutud häältest ja
varundusteenusesse varundatud e-valimiskastidest. Konsolideerimise protsess koosneb
järgmistest sammudest:

#. Talletusteenusesse kogutud hääled varundatakse varundusteenusesse. Selle
   tulemusena on varundusteenusesse salvestatud kõik kogutud e-valimiskastid;

#. Varundusteenuses koostatakse konsolideeritud e-valimiskast;

#. Konsolideeritud e-valimiskast kopeeritakse haldusteenusesse.

Konsolideeritud e-valimiskasti koostamise näide:

.. include:: genereeritud-failid/e-valimiskasti_koostamine.inc


Töötlemisrakenduse sisendi aluse koostamine
-------------------------------------------

Töötlemisrakenduse sisendi alus on häälte töötlemiseks vajalike sisendfailide
komplekt, mis genereeritakse kogumisteenuses salvestatud andmete põhjal.
Komplekti koosseis on järgmine:

#. Ringkondade nimekiri;

#. Valijate nimekirjad;

#. E-valimiskast kogutud häältega;

#. Häälte registreerimispäringute valideerimisandmed;

#. Töötlemisrakenduse seadistused.

Väljund on ZIP-konteiner, mis sisaldab järgmisi faile:

#. Ringkondade nimekiri digitaalselt signeerituna
   :file:`<election-id>.districts.json.asice`;

#. Valijate nimekirjade signeerimisvõtme avalik võti
   :file:`voterfile.pub.key`;

#. Valijate nimekirjad
   :file:`<changeset_no>.<election-id>.voters.utf`;

#. Valijate nimekirjade signatuurid
   :file:`<changeset_no>.<election-id>.voters.sig`;

#. Valijate nimekirja vahelejätmise korraldused
   :file:`<changeset_no>.<election-id>.voters-skip.yaml.asice`;

#. Registreerimispäringute verifitseerimise avalik võti
   :file:`ts.key`;

#. Töötlemisrakenduse seadistuste mall e-valimiskasti verifitseerimiseks
   :file:`<election-id>.processor.yaml`.

Töötlemisrakenduse sisendi alus koostatakse utiliidi
:ref:`ivxv-generate-processor-input` abil. Näide:

.. include:: genereeritud-failid/töötlemisrakenduse_sisendi_koostamine.inc

Hääletamise statistika eksportimine
-----------------------------------

Häälestamise statistika koostatake hääletusteenuses ja see koosneb kahest
osast: üldstatistika (hääletajate koguarv) ja detailstatistika. Üldstatistika
kopeeritakse haldusteenusesse ja eksporditakse Valimiste Infosüsteemi 15
minutilise intervalliga. Detailstatistika koostatakse ja eksporditakse
Valimiste Infosüsteemi käsitsi.

Häälestamise statistika importimine ja eksportimine toimub haldusteenuse
masinas utiliidi :ref:`ivxv-voterstats` abil. Üldstatistika importimise ja
eksportimise automaatika on teostatud cron-teenuse abil ja kirjeldatud failis
:file:`/etc/cron.d/ivxv-admin`.


Hääletamise seansside väljavõtte koostamine
-------------------------------------------

Hääletamise ja hääle kontrollimise seansside väljavõte on CSV-vormingus ja see
koostatakse logiseire teenuses.

Väljavõtet on võimalik koostada anonüümistatud kujul, kus kasutajate
isikukoodid ja IP-aadressid on asendatud anonüümsete väärtustega.

Võimalik on valida, kas väljastada kõik hääletamise seansid või ainult hääle
kontrollimisega seansid.

:ref:`ivxv-voting-sessions`
