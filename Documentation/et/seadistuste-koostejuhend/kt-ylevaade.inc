..  IVXV kogumisteenuse haldusteenuse kirjeldus

Ülevaade
--------

Kogumisteenuse juhtimine toimub signeeritud korralduste abil. Korraldused
koostatakse tekstiredaktori abil ja signeeritakse ID-kaardiga.

Nimekiri kogumisteenuse haldamise korraldustest:

#. Usaldusjuure seadistus;

#. Tehniline seadistus;

#. Valimiste seadistus;

#. Valikute nimekiri;

#. Kasutajate volituste määramine;

#. Valijate nimekiri.

Lisaks korraldustele tuleb kogumisteenusele genereerida ka komplekt
krüptovõtmeid.


Korralduste vorming
*******************

Kogumisteenuse korralduste vorming on enamasti YAML või JSON, valijate
nimekirja puhul kasutatakse spetsiifilist vormingut.

YAML-vormingus korraldused:

#. Usaldusjuure seadistus;

#. Tehniline seadistus;

#. Valimiste seadistus.

JSON-vormingus korraldused:

#. Valikute nimekiri;

#. Kasutajate volituste määramine.

Kohandatud vormingus korraldused:

#. Valijate nimekiri.


YAML-vormingus korraldused
~~~~~~~~~~~~~~~~~~~~~~~~~~

YAML-vormingus seadistustesse on võimalik kaasata väliseid faile. Selleks
kasutatakse silti ``!container``.

Näide:

.. code-block:: yaml

   # välja "ext_file" väärtus loetakse failist "certificate-file.pem"
   ext_file: !container certificate-file.pem

.. caution::

   Väliste failide kaasamisel tuleb arvestada sellega, et seadistused
   laaditakse süsteemi BDOC-vormingus konteinerites (vt. lõiku
   :ref:`korralduspaki-vorming`). See seab väliste failide kaasamisele
   järgmised nõuded:

   * Kasutatavad välised failid peavad olema pakendatud seadistusfailiga
     samasse konteinerisse;

   * Välised failid peavad asuma seadistusfailiga samas kataloogis.


.. _korralduspaki-vorming:

Korralduspaki vorming
*********************

Korralduspakk on BDOC-vormingus konteiner, milles on korraldusfail ja
signatuurid. Üks korralduspakk tohib sisaldada ainult ühte korraldust.
YAML-vormingus seadistuse (usaldusjuure seadistus, tehniline seadistus ja
valimiste seadistus) korral võivad failis olla ka seadistusfaili pool
kasutatavad välised failid.

.. vim: sts=3 sw=3 et syntax=rst:
