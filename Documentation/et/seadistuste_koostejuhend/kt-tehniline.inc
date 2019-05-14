..  IVXV seadistuste koostamise juhend

.. _kt-technical:

Kogumisteenuse tehnilise seadistususe koostamine
------------------------------------------------

Tehniline seadistus määrab kogumisteenuse tehnilised parameetrid. Sama tehnilist
seadistust peaks olema võimalik kasutada erinevate valimiste seadistustega [#]_.

.. [#] Aga mitte samaaegselt: kogumisteenus toetab ainult ühte valimist.

Tehnilise seadistuse koostab kogumisteenuse osutaja.

Seadistusfaili nimi peab alati lõppema stringiga :file:`technical.yaml`.
Failinime võimalik eesliide peab alati lõppema punktiga.

:debug:
        Tõeväärtus, kas logidesse kirjutatakse silumisteateid.

----

:filter:
        Kohustuslik väli.
        Alamblokk, mis sisaldab ühenduste filtrite seadistusi.

:filter.tls:
        Kohustuslik väli.
        Alamblokk, mis sisaldab ühenduste TLS-filtri seadistusi.

:filter.tls.handshaketimeout:
        Kohustuslik väli.
        Maksimaalne aeg sekundites TLS-kätluse läbiviimiseks.

:filter.tls.ciphersuites:
        Kohustuslik väli.
        TLS-protokollis rakendamiseks lubatud šifrikomplektid. Hetkel toetatud
        valikud on::

            TLS_RSA_WITH_AES_128_CBC_SHA
            TLS_RSA_WITH_AES_256_CBC_SHA
            TLS_RSA_WITH_AES_128_GCM_SHA256
            TLS_RSA_WITH_AES_256_GCM_SHA384
            TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA
            TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA
            TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA
            TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA
            TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
            TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256
            TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
            TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384

:filter.codec:
        Kohustuslik väli.
        Alamblokk, mis sisaldab ühenduste kodekfiltri seadistusi.

:filter.codec.rwtimeout:
        Kohustuslik väli.
        Maksimaalne aeg sekundites valijalt tervikliku päringu lugemiseks.
        Maksimaalne aeg sekundites valijale tervikliku päringu kirjutamiseks.

:filter.codec.requestsize:
        Päringu maksimaalne suurus baitides. Välja puudumise või väärtuse 0
        korral päringu suurust ei piirata.

:filter.codec.logrequests:
        Tõeväärtus, kas logidesse kirjutatakse kõik sissetulnud päringud
        algkujul. Päringu logi talletatakse tavalogist eraldi.

----

:network:
        Kohustuslik väli.
        Loetelu kogumisteenuse võrgusegmentidest.

        Kõik kogumisteenuse võrgusegmentide parameetrid määrab kogumisteenuse osutaja.

:network.*.id:
        Kohustuslik väli.
        Võrgusegmendi identifikaator.

:network.*.services:
        Kohustuslik väli.
        Alamblokk, mis sisaldab kogumisteenuse selle võrgusegmendi
        mikroteenuste seadistust.

        .. todo:: Kuigi kõik selle bloki parameetrid määrab Koguja, kas
                Korraldaja peaks vähemalt vahendusteenuste väliseid aadresse
                kinnitama, kuna need on otspunktid, kuhu valija- ja
                kontrollrakendus peavad ühendama hakkama?

:network.*.services.proxy:
        Loetelu, mis sisaldab vahendusteenuste isendite seadistust.

:network.*.services.dds:
        Loetelu, mis sisaldab Mobiil-ID toeteenuste isendite seadistust.

:network.*.services.choices:
        Loetelu, mis sisaldab nimekirjateenuste isendite seadistust.

:network.*.services.voting:
        Loetelu, mis sisaldab hääletamisteenuste isendite seadistust.

:network.*.services.verification:
        Loetelu, mis sisaldab kontrollteenuste isendite seadistust.

:network.*.services.storage:
        Loetelu, mis sisaldab talletusteenuste isendite seadistust.

:network.*.services.log:
        Loetelu, mis sisaldab logikogumisteenuste isendite seadistust.

:network.*.services.backup:
        Varundusteenuse isendi seadistus.

:network.*.services.*.id:
        Kohustuslik väli.
        Mikroteenuse isendi identifikaator.

:network.*.services.*.address:
        Kohustuslik väli.
        Mikroteenuse isendi võrguaadress ja -port.

        .. todo:: Sama parameetrit kasutatakse kirjeldamaks seda, millisel
                aadressil teenus peab kuulama hakkama ning milliselt aadressilt
                saavad haldusteenus ja vahendusteenus teenusele ligi. Kui
                sisemine ja välimine aadress hakkavad erinema, siis tuleb see
                parameeter lüüa kaheks.

:network.*.services.*.peeraddress:
        Mikroteenuse isenditevahelise suhtluse võrguaadress ja -port.
        Kasutatakse ainult juhul, kui sama teenust pakkuvad isendid peavad
        omavahel suhtlema (nt talletusteenus).

        .. todo:: Sama parameetrit kasutatakse kirjeldamaks seda, millisel
                aadressil teenus peab kuulama hakkama ning milliselt aadressilt
                saavad teised isendid teenusele ligi. Kui sisemine ja välimine
                aadress hakkavad erinema, siis tuleb see parameeter lüüa
                kaheks.

----

:logging:
        Alamblokk, mis sisaldab loetelu mikroteenuste kauglogimise serveritest.
        Siin blokis kirjeldatakse:

        * Logiseire teenus. Alati loetelus esimene. Logiseirele saadetakse IVXV
          logi alates tasemest INFO;

        * Vajadusel ka täiendavad välised logikogujad, kuhu saadetakse täielik
          logi alates tasemest DEBUG.

        Logimine toimub üle RELP protokolli.

        Kõik logiserverite parameetrid määrab kogumisteenuse osutaja.

:logging.address:
        Kohustuslik väli.
        Logiserveri aadress (IP-aadress või hostinimi).

:logging.port:
        Logiserveri port (täisarv, vaikimisi *20514*).

----

:storage:
        Kohustuslik väli.
        Alamblokk, mis sisaldab talletusprotokolli seadistust.

        Kõik talletusprotokolli parameetrid määrab kogumisteenuse osutaja.

:storage.protocol:
        Kohustuslik väli.
        Kogumisteenuse talletusprotokoll. Hetkel toetatud ainult ``etcd``.

:storage.conf:
        Kohustuslik väli.
        Talletusprotokolli seadistus. Sisu sõltub ``storage.protocol`` parameetri
        väärtusest.

:storage.conf.ca:
        Kohustuslik väli.
        Kasutatakse ainult juhul kui ``storage.protocol`` on ``etcd``.

        Talletusteenuste TLS sertifikaatide väljastaja sertifikaat.

:storage.conf.conntimeout:
        Kohustuslik väli.
        Kasutatakse ainult juhul kui ``storage.protocol`` on ``etcd``.

        Maksimaalne aeg sekundites etcd serveriga ühenduse loomiseks.

:storage.conf.optimeout:
        Kohustuslik väli.
        Kasutatakse ainult juhul kui ``storage.protocol`` on ``etcd``.

        Maksimaalne aeg sekundites ühe talletusoperatsiooni teostamiseks.

:storage.conf.bootstrap:
        Kohustuslik väli.
        Kasutatakse ainult juhul kui ``storage.protocol`` on ``etcd``.

        Loetelu nende talletusteenuste identifikaatoritest, mis on osa algsest
        etcd klastrist. Vajalik, et talletusteenuse isend teaks esmasel
        käivitumisel, kas see loob uut klastrit või liitub olemasolevaga.

        Esmases tehnilises seadistuses peab loetelu kattuma ``network`` blokis
        loetletud talletusteenuste identifikaatoritega. Hiljem teenuseid
        lisades või eemaldades ei tohi ``storage.conf.bootstrap`` väärtust
        uuendada.

----

:backup:
        Varunduse parameetrid.

        Loetelu varundmise kellaaegadest vormingus HH:MM.

:file:`example.technical.yaml`:

.. literalinclude:: config-examples/example.technical.yaml
   :language: yaml
   :linenos:

.. vim:syntax=rst:
