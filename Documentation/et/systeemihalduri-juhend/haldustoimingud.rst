..  IVXV kogumisteenuse haldusjuhend

Süsteemi haldustoimingud
========================

Kogumisteenuse oleku jälgimine
------------------------------

Kogumisteenuse olekuandmed registreeritakse haldusteenuse anbmebaasis. Oleku
kuvamiseks on utiliit :ref:`ivxv-status`.

Olekus kuvatakse järgmisi andmeid:

* Valimise ID, faas, algus- ja lõpuaeg;

* Laaditud seadistuspakkide versioonid;

* Laaditud valimisnimekirjade versioonid;

* Teenuste nimekiri koos rakendatud seadistuste versioonidega, teenuse seisundi
  ja selle viimase tuvastamise ajaga;

* Väliste teenuste seisundid;

* Haldusteenuse andmehoidla statistika.

Sõltuvalt kogumisteenuse seisundis võib oleku kuvamise utiliit jätta mõned
andmeblokid kuvamata. Täieliku andmestiku väljastamiseks vaata utiliidi
:ref:`ivxv-status` abiteavet.

Alamteenuste oleku jälgimise ja haldusteenuse andmebaasis registreerimisega
tegeleb haldusteenuse agentdeemon.


.. _korralduste-rakendamine:

Korralduste rakendamine
-----------------------

Kogumisteenuse korraldused koostatakse signeeritud korralduspakkidena, millega
kirjeldatakse kasutaja identifikaator (*Common Name* ehk CN väli ID-kaardilt)
ja rollide nimekiri.

Korralduste valideerimine ja rakendamine toimub käsuga :ref:`ivxv-cmd-load`.

Valimise seadistuse valideerimise näide::

   $ ivxv-cmd-load --validate-only election valimise-seadistus-TEST2017.bdoc

Valikute nimekirja korralduse rakendamise näide::

   $ ivxv-cmd-load choices valikute-nimekiri-TEST2017.bdoc

.. seealso::

   * Käsu :ref:`ivxv-cmd-load` abiteave;

   * Korralduste rollide kirjeldus ja korralduste koostamise juhend asuvad
     dokumendis ``IVXV seadistuste
     koostamise juhend``.


Kasutajate haldus
-----------------

Kasutajate algsed kirjeldused määratakse usaldusjuure seadistuses, hilisem
haldus toimub vastavate korralduste abil.

Kasutajate halduse korraldused rakendatakse käsuga :ref:`ivxv-cmd-load` (vaata
:ref:`korralduste-rakendamine`).

Kasutajaõiguste määramise korralduse rakendamise näide::

   $ ivxv-cmd-load user ORAV,IVAN,30809010001-admin.bdoc

.. attention::

   Juba lisatud kasutajate eemaldamine süsteemist pole võimalik. Kasutaja
   eemaldamise asemel tuleb kasutaja rolliks määrata "õigusteta kasutaja".

.. seealso::

   * Kasutajate rollide kirjeldus ja volituste korralduste koostamise juhend
     asuvad dokumendis ``IVXV
     seadistuste koostamise juhend``.

   * Korralduste rakendamine on kirjeldatud lõigus
     :ref:`korralduste-rakendamine`.


Varukoopiate tegemine
---------------------

.. todo:: Varukoopiate tegemine


Varukoopiast taastamine
-----------------------

.. todo:: Varukoopiast taastamine

.. vim: sts=3 sw=3 et:
