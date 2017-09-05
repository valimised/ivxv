..  IVXV kogumisteenuse haldusteenuse kirjeldus

.. _usaldusjuur:

Usaldusjuur
===========


.. _app-trust:

Rakenduste usaldusjuure kirjeldamine
------------------------------------

Rakenduste usaldusjuur sisaldab andmeid seadistuste (kaasa arvatud
usaldusjuure enda) allkirjade kontrollimiseks.

Usaldusjuure seadistuse koostab valimiste korraldaja.

:ca:

    Komadega eraldatud loetelu konteineris sisalduvatest CA sertifikaatidest ja
    vahesertifikaatidest.

:ocsp:

    Komadega eraldatud loetelu konteineris sisalduvatest OCSP sertifikaatidest.

:tsa:

    Komadega eraldatud loetelu konteineris sisalduvatest ATO sertifikaatidest.

Kõik sertifikaadid antakse PEM vormingus.

Rakendusele esitatakse usaldusjuur BDOC konteineris, kus usaldusjuure
spetsifikatsioon on kirjeldatud failis `ivxv.properties` ning kõik juure
elemendid on konteinerisse laetud.


Näide
*****

:file:`ivxv.properties`:

.. literalinclude:: config-examples/ivxv.properties
   :linenos:


:file:`ivxv.properties`:

.. literalinclude:: config-examples/ivxv.properties.real
   :linenos:

Kogumisteenuse usaldusjuur
--------------------------

Kogumisteenuse usaldusjuur sisaldab andmeid seadistuste (kaasa arvatud
usaldusjuure enda) allkirjade kontrollimiseks ja nimekirja süsteemi esmastest
volitustest.

Usaldusjuure seadistuse koostab valimiste korraldaja.
Seadistusfaili nimi peab alati olema :file:`trust.yaml`.

.. attention::

   Usaldusjuure seadistuste laadimine lähtestab kogumisteenuse. Seetõttu pole
   juba seadistatud kogumisteenuse usaldusahela muutmine võimalik. Volituste
   muutmine on võimalik vastavate korralduste abil.

:container:

        Kohustuslik väli.
        Alamblokk, mis sisaldab seadistusfailide allkirjade kontrollimise
        seadistust.

:container.bdoc:

        Alamblokk, mis sisaldab seadistusfailide BDOC-allkirjade kontrollimise
        seadistust.

:container.bdoc.bdocsize:
        Kohustuslik väli.
        BDOC konteineri maksimaalne lubatud suurus baitides.

        Määrab Korraldaja.

:container.bdoc.filesize:
        Kohustuslik väli.
        BDOC konteineris olevate failide maksimaalne lubatud hõrendatud suuru
        baitides.

        Määrab Korraldaja.

:container.bdoc.roots:

        Kohustuslik väli.
        Seadistuste allkirjastajate sertifikaatide usaldusjuured.

:container.bdoc.intermediates:

        Seadistuste allkirjastajate sertifikaatide vahesertifikaadid.
        Usalduse saavutamiseks peab nende sertifikaatide abil olema võimalik
        luua ahel allkirjastaja sertifikaadist usaldusjuureni.

:container.bdoc.checktimemark:

        Kohustuslik väli.
        Tõeväärtus, kas seadistuste allkirjadel peavad olema BDOC
        spetsifikatsiooni jaotises 6.1 kirjeldatud ajamärgid. Kui tõene, siis
        ilma ajamärkideta allkirju ei usaldata. Kui väär, siis ajamärgid on
        lubatud, aga neid ei kontrollita.

:authorizations:

        Kohustuslik väli.
        Esmane nimekiri halduri volitustega isikutest, mis rakendatakse
        süsteemile usaldusjuure laadimisel. Iga isiku kohta on kirje tema
        ID-kaardi välja ``Common Name`` (CN) väärtusega. Minimaalselt peab
        sisaldama usaldusjuure signeerinud isiku andmeid.


Näide
*****

:file:`trust.yaml`:

.. literalinclude:: config-examples/trust.yaml
   :language: yaml
   :linenos:


.. vim: sts=3 sw=3 et:
