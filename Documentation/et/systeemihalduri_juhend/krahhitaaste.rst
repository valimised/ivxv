..  IVXV kogumisteenuse haldusjuhend

Krahhitaaste
============

Kogumisteenus on projekteeritud nii, et teenuse või selle osade krahhimise
tagajärjel ei tekiks andmekadu või oleks see minimaalne.


Eeldused edukaks krahhitaasteks
-------------------------------

Kõrgkäideldav seadistus
^^^^^^^^^^^^^^^^^^^^^^^

Peamine eeldus edukaks krahhitaasteks on kogumisteenuse paigaldamine
kõrgkäideldava seadistusega, mis määrab vähemalt kolme talletusteenuse isendi
kasutamise. Lisaks on krahhiolukorra kiiremaks lahendamiseks kasulik eraldada
mikroteenustele ühe lisaisendite komplekti paigalduseks vajalik taristu.


Logikoguja kasutamine
^^^^^^^^^^^^^^^^^^^^^

Kogumisteenuse seadistus peab kirjeldama logikogumisteenuse, et mikroteenuste
poolt toodetavad logisid oleks võimalik lihtsal moel kokku koguda. Soovitav on
kasutada mitut logikogujat erinevas füüsilises lokatsioonis, et minimeerida
logikirjete kaotsimineku võimalust.


Varundusteenuse kasutamine
^^^^^^^^^^^^^^^^^^^^^^^^^^

Kogumisteenuse seadistus peab kirjeldama varundusteenuse ning automaatse
varundamise ajad piisava sagedusega. Samuti on soovitav teha varukoopiad ka
varundusteenusest.

Automaatne varundamine tagab e-urni koopia säilimise :ref:`talletusteenuse
täieliku krahhi <talletusteenuste-täielik-krahh>` korral.

.. note::

   Varundusteenus on soovitav paigaldada teistest kogumisteenuse isenditest
   füüsiliselt eraldi, et võimalikud eriolukorrad (näiteks tulekahju) ei
   mõjutaks korraga nii varundusteenust kui teisi teenuseid.

Varundusteenus on projekteeritud kogumisteenuse andmetest automaatsete
varukoopiate loomiseks ühte kohta ning nende kättesaadavaks tegemiseks
operatsioonidele, mis varukoopiaid kasutavad (näiteks häälte kokkulugemine).

.. note::

   Kogumisteenuse osutaja peaks kaaluma võimalust teha varundusteenusest
   täiendavaid varukoopiaid, et tagada varundatud andmete säilimine ka
   varundusteenuse krahhi korral.

Valmisolek krahhiks
^^^^^^^^^^^^^^^^^^^

Kogumisteenuse krahh mõjutab kõiki e-hääletamise komponente, erilist
tähelepanu tuleb pöörata valijarakenduste ja kontrollrakenduste
nimelahendusele ning TLS ühenduste usaldamiseks vajalikele
sertifikaatidele.

Hääletamise edukaks läbiviimiseks tuleb tagada, et nimeserverid
sisaldaks kogu hääletusperioodi vältel ajakohast infot
hääletamissüsteemi sisendpunktide kohta - siis suudavad
valijarakendused ja kontrollrakendused vastavalt muutuvatele oludele
nimesid korrektselt lahendada.

#. Krahhimise tuvastamisel tuleb esimeste tegevuste hulgas eemaldada
   nimelahendusest krahhinud teenus, et rakendused enam selle poole
   pöörduda ei saaks.
#. Kui teenus(ed) pärast krahhi uuesti töökorda saadakse, tuleb
   viimase sammuna nimelahenduses panna uute teenuste aadressid
   lahenduma vastavalt rakendustes defineeritule.

Kui kogumisteenusesse lisatakse uusi mikroteenuseid (eeldatavalt
pärast krahhimist), siis on tarvis tagada lisatud teenuste
usaldusväärsus rakendustes.

Teenuse plaanimisel tuleb luua serdid/võtmed ka võimalike
asendusteenuste jaoks (dds, choices, voting). Need võtmed tuleb
pakendada valijarakendusse, et pärast krahhi poleks tarvis hakata uut
rakendust levitama. Kui sertifikaadid luuakse ühe CA alt, siis piisab
valijarakendusse vastava CA sertifikaadi pakendamisest.
Kontrollrakenduste jaoks tuleb seadistustes alati näidata konkreetsed
teenussertifikaadid, kuid kontrollrakenduste seadistuste muutmine ei
eelda kontrollrakenduste uuesti levitamist.

Teenuste taastamine krahhist
----------------------------


Mikroteenuse isendi krahh ilma andmekaota
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Mikroteenuse isendi krahh ilma andmekaota võib esineda teenuste puhul, mis ei
tegele andmete säilitamisega (nimekirjateenus, hääletusteenus, kontrolliteenus
või mobiil-id tugiteenus). Sellises olukorras piisab teenuse isendi
taastamiseks kas teenuse taaskäivitamisest (kui see on võimalik) või teenuse
isendi asendamisest uuega.

.. seealso::

   * :ref:`teenuse-taaskäivitamine`

   * :ref:`teenuse-asendamine`


Logikogumisteenuse isendi krahh
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Logikogumisteenuse krahh võib esineda nii logiandmete riknemisega kui ka ilma.

Ilma logiandmete riknemiseta krahh tähendab olukorda, kus rsyslog teenus seisab
ja ei võta seetõtte teenustelt logikirjeid vastu ning salvestatud logifailid ei
ole rikutud. Sellises olukorras piisab teenuse isendi töökorda seadmiseks selle
taaskäivitamisest.

Logikogumisteenuse krahh koos logiandmete riknemisega nõuab teenuse isendi
asendamist uuega.

Kui logiandmete riknemisega kaasneb alati logiandmete kadu, siis ilma
riknemiseta krahhi puhul tuleb samuti selle võimalusega arvestada. Logisid
edastatakse üle RELP-protokolli, mis on küllalt töökindel, kuid vaatamata
sellele võib logiedastus katkeda olukorras, kus logi genereeriva teenuse hostil
on rsyslogi isendit taaskäivitatud ajal, mil logikoguja rsyslog isend ei
töötanud.

.. seealso::

   * :ref:`teenuse-taaskäivitamine`

   * :ref:`teenuse-asendamine`

   * `RELP - The Reliable Event Logging Protocol
     <https://www.rsyslog.com/doc/relp.html>`_


Varundusteenuse isendi krahh
^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Varundusteenuse isendi krahh tähendab varundusteenusesse varundatud andmete
riknemist. Teenuse taastamiseks tuleb varundusteenus uuesti paigaldada ja
varundatud andmed taastada. Andmete taastamine varundusserverisse võib toimuda
ka pärast häälte kogumise lõppemist, kuid enne häälte kokkulugemist.

.. note::

   Varundusprotseduuride käivitamist juhitakse haldusteenusest ja seetõttu pole
   varundusteenust võimalik käivitada ega seisma jätta.


Talletusteenuse isendi krahh
^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Talletusteenuse ühe isendi krahhimisel piisab isendi asendamisest uuega.

Talletusteenuseid saab lisada ja eemaldada ainult siis, kui klastris on
vähemalt kvoorumi jagu töökorras talletusteenuse isendeid. Kvoorumi suurus on
N/2+1 ümardatud alla, kus N on seadistatud isendite arv (näiteks kolme
seadistatud isendi korral on kvoorumi suurus kaks).

Kui talletusteenuse isendeid jääb alles vähem kui kvoorumi jagu, siis tuleb
teha kõigile isenditele uus paigaldus (vt.
:ref:`talletusteenuste-täielik-krahh`).

Talletusteenuse kvoorumist tingitud piirangud:

#. Talletusteenuse isendite arvu ei ole kunagi võimalik vähendada ühele;

#. Talletusteenuste isendite eemaldamisel peab arvestama kvoorumi säilimisega.

   Näide: kui on seadistatud 6 talletusteenuse isendit (kvoorum=4), siis sealt
   ei saa korraga eemaldada kolme isendit (jääks järgi kolm isendit,
   kvoorum=2), kuna seadistatud isendite hulk oleks siis väiksem kui algne
   kvoorum. Kõigepealt tuleb eemaldada üks (jääb järgi 5 isendit, kvoorum=3)
   isend ja alles pärast seda saab eemaldada ülejäänud kaks.

.. seealso::

   * :ref:`teenuse-taaskäivitamine`

   * :ref:`teenuse-asendamine`


.. _talletusteenuste-täielik-krahh:

Talletusteenuste täielik krahh
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Talletusteenuste täielikul asendamisel tuleb koostada uus tehniline seadistus,
mis vastab järgmistele tingimustele:

* ei sisalda ühtegi vana talletusteenust;

* kõik uued talletusteenused on loetletud parameetri ``storage.conf.bootstrap``
  nimekirjas.

.. important::

   Talletusteenuste täielikul asendamisel tuleb arvestada järgnevada:

   * enne asendamist kogutud hääled säilivad varundusserveritesse tehtud
     varukoopiates;

   * varukoopia loomise ja krahhi vahel kogutud hääled lähevad kaotsi;

   * valikute ja valijate nimekirjad tuleb teenustele uuesti rakendada.


Kogumisteenuse täielik asendamine
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Kui tekib vajadus kogumisteenuse täielikuks asendamiseks, siis tuleb
kogumisteenusele teha uus paigaldus ilma andmete taastamiseta, mis on kiireim
meetod teenuse uuesti töökorda seadmiseks.

Varasemalt kogutud häälte kaasamiseks häälte kokkulugemisele tuleb
varundusserverisse taastada eelnevalt loodud varukoopiad.

.. important::

   Kogumisteenuse täielikul asendamisel tuleb arvestada, et enne asendamist
   kogutud hääled säilivad varundusserveritesse tehtud varukoopiates. Pärast
   varukoopia loomist kogutud hääled lähevad kaotsi.
