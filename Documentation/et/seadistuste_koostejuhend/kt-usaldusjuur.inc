..  IVXV kogumisteenuse haldusteenuse kirjeldus

.. _kt-trust:

Kogumisteenuse usaldusjuure seadistamine
----------------------------------------

Kogumisteenuse usaldusjuur sisaldab andmeid seadistuste (kaasa arvatud
usaldusjuure enda) allkirjade kontrollimiseks ja nimekirja süsteemi esmastest
volitustest.

Usaldusjuure seadistuse koostab valimiste korraldaja.
Seadistusfaili nimi peab alati lõppema stringiga :file:`trust.yaml`.
Failinime võimalik eesliide peab alati lõppema punktiga.

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

:container.bdoc.filesize:
        Kohustuslik väli.
        BDOC konteineris olevate failide maksimaalne lubatud hõrendatud suuru
        baitides.

:container.bdoc.roots:

        Kohustuslik väli.
        Seadistuste allkirjastajate sertifikaatide usaldusjuured.

:container.bdoc.intermediates:

        Seadistuste allkirjastajate sertifikaatide vahesertifikaadid.
        Usalduse saavutamiseks peab nende sertifikaatide abil olema võimalik
        luua ahel allkirjastaja sertifikaadist usaldusjuureni.

:container.bdoc.profile:

        Kohustuslik väli.
        Seadistuste allkirjadelt nõutav BDOC profiil. Toetatud valikud on
        ``BES`` (põhiprofiil kirjeldatud BDOC spetsifikatsiooni jaotises 5),
        ``TM`` (ajamärkidega profiil kirjeldatud BDOC spetsifikatsiooni
        jaotises 6.1) ja ``TS`` (ajatemplitega profiil kirjeldatud BDOC
        spetsifikatsiooni jaotises 6.2).

:container.bdoc.ocsp.responders:

        Kasutatakse ainult juhul kui ``container.bdoc.profile`` on ``TM`` või
        ``TS``.

        Kehtivuskinnitusi väljastanud OCSP responderi sertifikaadid. Kui nende
        hulgast responderi sertifikaati ei leita, siis otsitakse OCSP vastuses
        olevate sertifikaatide hulgast selline, mis on antud välja sama
        väljastaja poolt, mis seadistuste allkirjastaja sertifikaat, ning on
        lubatud OCSP vastuste signeerimiseks. AIA loogika kasutamise
        korral võib väli olla tühi.

:container.bdoc.tsp.signers:

        Kohustuslik väli.
        Kasutatakse ainult juhul kui ``container.bdoc.profile`` on ``TS``.

        Ajatempliteenuseteenuse vastuse allkirjastamise sertifikaadid.

:container.bdoc.tsp.delaytime:

        Kohustuslik väli.
        Kasutatakse ainult juhul kui ``container.bdoc.profile`` on ``TS``.

        Maksimaalne ajanihe ajatempli loomise ja allkirjastamise vahel
        sekundites.

:container.bdoc.tsdelaytime:

        Kasutatakse ainult juhul kui ``container.bdoc.profile`` on ``TS``.

        Maksimaalne ajanihe ajatempli ja kehtivuskinnituse loomise vahel
        sekundites. Välja puudumise või väärtuse 0 korral peavad mõlemad olema
        loodud samal sekundil.

:authorizations:

        Kohustuslik väli.
        Esmane nimekiri kogumisteenuse halduri volitustega isikutest (vt.
        :ref:`rollid`), mis rakendatakse süsteemile usaldusjuure laadimisel.
        Iga isiku kohta on kirje tema ID-kaardi välja ``Common Name`` (CN)
        väärtusega. Minimaalselt peab sisaldama usaldusjuure signeerinud isiku
        andmeid.


Näide
*****

:file:`example.trust.yaml`:

.. literalinclude:: config-examples/example.trust.yaml
   :language: yaml
   :linenos:
