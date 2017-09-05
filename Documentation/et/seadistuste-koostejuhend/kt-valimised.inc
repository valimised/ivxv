..  IVXV seadistuste koostamise juhend

Valimise seadistus
------------------

Valimise seadistus määrab ühe valimise seadistuse.

Valimise seadistuse koostab valimiste korraldaja.
Seadistusfaili nimi peab alati olema :file:`election.yaml`.

:identifier:

        Kohustuslik väli.
        Valimise unikaalne identifikaator. 

:questions:

        Loetelu, mis sisaldab ühe või enama valimise küsimuse unikaalset
        identifikaatorit. Unikaalsus peab olema tagatud ainult konkreetse
        valimise küsimuste hulgas. Kohustuslik väli.

----

:period:

        Kohustuslik väli.
        E-hääletuse perioodi andmete alamblokk.

:period.servicestart:

        Kohustuslik väli.
        Kogumisteenuses häälte vastuvõtmise algusaeg. Sellest hetkest alates
        hakkab kogumisteenus ühendusi teenindama. See aeg peab eelnema valimise
        algusajale ning on mõeldud enne valimise algust proovihääle andmiseks.

        Enne ``electionstart`` parameetriga määratud aega vastu võetud häälte
        puhul tagastatakse valijarakendusele hääle esitamise lõpus vastav
        veateade (hääl jõudis kohale enne valimise algust). Sellised hääled
        tühistatakse automaatselt töötlemise käigus.

:period.electionstart:

        Kohustuslik väli.
        E-hääletuse algusaeg. Sellest hetkest alates antud hääled lähevad
        häälte lugemisel arvesse.

:period.electionstop:

        Kohustuslik väli.
        E-hääletuse lõpuaeg. Sellest hetkest lõpetatakse valikute nimekirjade
        väljastamine.

:period.servicestop:

        Kohustuslik väli.
        E-hääletuse lõppemisaeg. Sellest hetkest lõpetatakse häälte
        vastuvõtmine ning teenused lõpetavad töö.

----

:verification:

        Kohustuslik väli.
        Hääle kontrollimise parameetrite alamblokk.

:verification.count:

        Kohustuslik väli.
        Suurim ühe hääle lubatud kontrollimiste arv.

:verification.minutes:

        Kohustuslik väli.
        Hääle kontrollimise perioodi kestus minutites. Pärast hääle
        andmist on selle perioodi vältel võimalik häält kontrollida.

----

:ignorevoterlist:

        Ringkonna identifikaator, mille valikud esitada kõigile valijatele.
        Kui see väärtus ei ole tühi, siis kogumisteenus ei kasuta valijate
        nimekirja ning esitab kõigile valijatele määratud ringkonna valikud ja
        lubab kõigil, kellel õnnestub isikutuvastus ning hääle allkirja
        kontrollimine, hääletada.

Näide
*****

:file:`election.yaml`:

.. literalinclude:: config-examples/election.yaml
   :language: yaml
   :linenos:

.. vim: sts=3 sw=3 et syntax=rst:
