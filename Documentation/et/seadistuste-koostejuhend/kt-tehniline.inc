..  IVXV seadistuste koostamise juhend

.. _kt-tehniline-seadistus:

Tehniline seadistus
-------------------

Tehniline seadistus määrab kogumisteenuse tehnilised parameetrid. Sama tehnilist
seadistust peaks olema võimalik kasutada erinevate valimiste seadistustega [#]_.

.. [#] Aga mitte samaaegselt: kogumisteenus toetab ainult ühte valimist.

Seadistusfaili nimi peab alati olema :file:`technical.yaml`.

Tehnilises seadistuses on laienduste seadistamiseks kaks erinevat lähenemist:

1. Kui sama funktsionaalsust saavad korraga pakkuda mitu laiendust, siis igaüks
   saab oma alambloki ning laiendus on seadistatud, kui blokk ei ole tühi. Näiteks
   ``container``, kus ``container.bdoc`` on vaid üks võimalikest allkirjade
   kontrollimise laiendustest ning saab töötada rööbiti teiste ``container.*``
   laiendustega.

2. Kui sama funktsionaalsust saab pakkuda korraga ainult üks laiendus, siis
   laienduse nimi tuuakse eraldi välja ning tema seadistus pannakse alamblokki.
   Näiteks ``storage``, kus saab korraga kasutada ainult ühte talletusprotokolli.
   ``storage.protocol`` määrab laienduse ning seadistus läheb ``storage.conf`` blokki.
   Sellise struktuuriga saab korraga seadistada ainult ühe laienduse.

.. todo::

   Tehnilistesse seadistustesse on vaja lisada:

   * Varunduse ajastused (väike näitekonf on haldusteenuse dokumentatsioonis)

   * Varundamiseks vajalike pääsude kirjeldamine (varundusteenus on
     kogumisteenuse jaoks väline teenus)

     Talletusteenus peab saama varundusteenusesse kirjutada

   * E-urni väljastamiseks vajalike pääsude kirjeldamine

     Haldusteenus peab saama varundusteenusest lugeda

:debug:
        Tõeväärtus, kas logidesse kirjutatakse silumisteateid.

        Määrab kogumisteenuse osutaja.

----

:voterlist:
        Kohustuslik väli.
        Valijate nimekirjade kontrollimise parameetrid.

:voterlist.key:
        Kohustuslik väli.
        RSA-võtmepaari avalik võti valijate nimekirjade allkirja
        kontrollimiseks.

        Määrab Korraldaja.

----

:auth:
        Kohustuslik väli.
        Alamblokk, mis sisaldab valija tuvastamise seadistust.

        Kõik valija tuvastamise parameetrid määrab Korraldaja.

:auth.ticket:
        Alamblokk, mis sisaldab piletipõhise valija tuvastamise seadistust.

        Piletipõhist valija tuvastamist kasutatakse Mobiil-ID puhul, kus
        ``dds`` teenus tuvastab valija ning väljastab talle pileti, millega
        teistele teenustele ennast tuvastada.

        See alamblokk on tühi, aga tema olemasolek või puudumine määrab, kas
        piletipõhine valija tuvastus on lubatud või ei.

:auth.tls:
        Alamblokk, mis sisaldab TLS-põhise valija tuvastamise seadistust.

        TLS-põhist valijatuvastust kasutatakse ID-kaardi puhul.

:auth.tls.roots:
        Kohustuslik väli.
        Valija TLS-klientsertifikaatide usaldusjuured.

:auth.tls.intermediates:
        Valija TLS-klientsertifikaatide vahesertifikaadid. TLS-autentimiseks
        peab nende sertifikaatide abil olema võimalik luua ahel valija
        klientsertifikaadist usaldusjuureni.

:auth.tls.ocsp:
        Alamblokk, mis sisaldab valija TLS-klientsertifikaatide oleku
        kontrollimise seadistust. Selle bloki puudumisel valija sertifikaatide
        kehtivust ei kontrollita välisest kehtivuskinnitusteenusest.

:auth.tls.ocsp.url:
        Kohustuslik väli.
        Valija TLS-klientsertifikaatide kehtivuskinnitusteenuse aadress.

:auth.tls.ocsp.responders:
        Valija TLS-klientsertifikaatide kehtivuskinnitusteenuse responderi
        sertifikaadid. Kui nende hulgast responderi sertifikaati ei leita, siis
        otsitakse vastuses olevate sertifikaatide hulgast selline, mis on antud
        välja sama väljastaja poolt, mis kontrollitav sertifikaat, ning on
        lubatud OCSP vastuste signeerimiseks.

:identity:
        Tuvastatud valija X.500 eraldusnimest unikaalse identifikaatori
        tuletamise meetod. Hetkel toetatud valikud ``commonname`` ja
        ``serialnumber``.

        Eesti elektrooniliste isikut tõendavate dokumentide korral on esimese
        valiku puhul identifikaator kujul "PERENIMI,EESNIMI,ISIKUKOOD" ning
        teise puhul "ISIKUKOOD".

----

:age:
        Alamblokk, mis sisaldab valija vanuse kontrolli seadistust. Kui see
        blokk puudub, siis valija vanust ei kontrollita.

        Kõik valija vanuse kontrollimise parameetrid määrab Korraldaja.

:age.method:
        Kohustuslik väli.
        Valija sünniaja tuvastamiseks kasutatav meetod. Hetkel toetatud ainult
        ``estpic``, mis eeldab, et valija unikaalne identifikaator on Eesti
        isikukood ning eraldab sealt sünniaja.

:age.timezone:
        Kohustuslik väli.
        IANA ajavööndi nimi, milles valija vanus arvutatakse ehk millises
        ajavööndis peab valija olema valimisealine.

:age.limit:
        Kohustuslik väli.
        Valija peab olema vähemalt nii vana, et hääletada. Kui väärtus on 0,
        siis valija vanust ei kontrollita.

----

:vote:
        Kohustuslik väli.
        Alamblokk, mis sisaldab häälte allkirjade kontrollimise seadistust.

:vote.bdoc:
        Alamblokk, mis sisaldab häälte BDOC-allkirjade kontrollimise seadistust.

        Kõik BDOC-allkirjade kontrollimise parameetrid määrab Korraldaja.

:vote.bdoc.bdocsize:
        Kohustuslik väli.
        BDOC konteineri maksimaalne lubatud suurus baitides.

:vote.bdoc.filesize:
        Kohustuslik väli.
        BDOC konteineris olevate failide maksimaalne lubatud hõrendatud suurus
        baitides.

:vote.bdoc.roots:
        Kohustuslik väli.
        Häälte allkirjastajate sertifikaatide usaldusjuured.

:vote.bdoc.intermediates:
        Häälte allkirjastajate sertifikaatide vahesertifikaadid.
        Hääle arvesseminekuks peab nende sertifikaatide abil olema võimalik
        luua ahel allkirjastaja sertifikaadist usaldusjuureni.

:vote.bdoc.checktimemark:
        Kohustuslik väli.
        Tõeväärtus, kas häälte allkirjadel peavad olema BDOC spetsifikatsiooni
        jaotises 6.1 kirjeldatud ajamärgid. Kui tõene, siis ilma ajamärkideta
        hääli vastu ei võeta. Kui väär, siis ajamärgid on lubatud, aga neid ei
        kontrollita.

        See peaks olema väär, kuna kõikide allkirjastamisvahendite puhul ei ole
        sissetuleval häälel ajamärki (nt Eesti ID-kaart). Ajamärgid ei ole siin
        kohustuslikud, kuna kogumisteenus küsib allkirjastaja sertifikaadile ise
        kehtivuskinnituse.

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

:filter.codec:
        Kohustuslik väli.
        Alamblokk, mis sisaldab ühenduste kodekfiltri seadistusi.

:filter.codec.rwtimeout:
        Kohustuslik väli.
        Maksimaalne aeg sekundites valijalt tervikliku päringu lugemiseks.
        Maksimaalne aeg sekundites valijale tervikliku päringu kirjutamiseks.

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

        Kõik logiserverite parameetrid määrab Korraldaja.

:logging.address:
        Kohustuslik väli.
        Logiserveri aadress (ip-aadress või hostinimi).

:logging.port:
        Logiserveri port (täisarv, vaikimisi 514).

:logging.protocol:
        Kohustuslik väli.
        Logiserverisse logimise protokoll (*udp* või *tcp*, vaikimisi *udp*).

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

----

:dds:
        Alamblokk, mis sisaldab Mobiil-ID teenusepakkuja seadistust.

        Kõik Mobiil-ID teenusepakkuja parameetrid määrab Korraldaja.

:dds.url:
        Kohustuslik väli.
        Mobiil-ID teenusepakkuja asukoht.

:dds.language:
        Kohustuslik väli.
        Mobiil-ID kasutajale kuvatavate sõnumite keel. Võimalikud väärtused
        ``EST``, ``ENG``, ``RUS`` ja ``LIT``.

:dds.servicename:
        Kohustuslik väli.
        Mobiil-ID teenusepakkujaga kokkulepitud teenusenimi.

:dds.authmessage:
        Kohustuslik väli.
        Sõnum, mida Mobiil-ID kasutajale kuvada autentimise käigus.

:dds.signmessage:
        Kohustuslik väli.
        Sõnum, mida Mobiil-ID kasutajale kuvada allkirjastamise käigus.

:dds.roots:
        Kohustuslik väli.
        Mobiil-ID sertifikaatide usaldusjuured.

:dds.intermediates:
        Mobiil-ID sertifikaatide vahesertifikaadid. Mobiil-ID autentimiseks
        peab nende sertifikaatide abil olema võimalik luua ahel Mobiil-ID
        sertifikaadist usaldusjuureni.

:dds.ocsp.responders:
        Mobiil-ID sertifikaatide OCSP responderi sertifikaadid. Kui nende
        hulgast responderi sertifikaati ei leita, siis otsitakse vastuses
        olevate sertifikaatide hulgast selline, mis on antud välja sama
        väljastaja poolt, mis kontrollitav sertifikaat, ning on lubatud OCSP
        vastuste signeerimiseks.

----

:qualification:
        Loetelu välistest kvalifitseerivatest päringutest, mis tehakse iga
        hääle kohta, koos seadistustega.

        Siin on kasutatud loetelu protokoll ja seadistus blokkidest selle
        asemel, et anda igale protokollile oma blokk, kuna kvalifitseerivate
        päringute järjekord on oluline ning seadistatav.

        Kõik kvalifitseerivate päringute parameetrid määrab Korraldaja.

:qualification.*.protocol:
        Kohustuslik väli.
        Kvalifitseeriva päringu protokoll. Hetkel toetatud ``ocsp``,
        ``ocsptm``, ``tsp`` ja ``tspreg``.

:qualification.*.conf:
        Kohustuslik väli.
        Kvalifitseeriva päringu protokolli seadistus. Sisu sõltub
        ``qualification.*.protocol`` parameetri väärtusest.

:qualification.*.conf.url:
        Kohustuslik väli.
        Aadress, kuhu kvalifitseeriv päring tehakse.

:qualification.*.conf.responders:
        Kasutatakse ainult juhul kui ``qualification.*.protocol`` on ``ocsp``
        või ``ocsptm``.

        OCSP reponderi sertifikaadid. Kui nende hulgast responderi sertifikaati
        ei leita, siis otsitakse vastuses olevate sertifikaatide hulgast
        selline, mis on antud välja sama väljastaja poolt, mis kontrollitav
        sertifikaat, ning on lubatud OCSP vastuste signeerimiseks.

:qualification.*.conf.signers:
        Kohustuslik väli.
        Kasutatakse ainult juhul kui ``qualification.*.protocol`` on ``tsp``
        või ``tspreg``.

        Ajatempliteenuseteenuse vastuse allkirjastamise sertifikaadid.

:qualification.*.conf.delaytime:
        Kohustuslik väli.
        Kasutatakse ainult juhul kui ``qualification.*.protocol`` on ``tsp``
        või ``tspreg``.

        Maksimaalne ajanihe ajatempli loomise ja allkirjastamise vahel
        sekundites.

----

:stats:
        E-hääletuse statistika laadimise parameetrid.

        Kõik statistika laadimise parameetrid määrab Korraldaja.

:stats.url:
        Kohustuslik väli.
        E-hääletuse statistika laadimise URL.

:stats.tlscert:
        Kohustuslik väli.
        E-hääletuse statistika laadimise serveri TLS-sertifikaat.

:file:`technical.yaml`:

.. literalinclude:: config-examples/technical.yaml
   :language: yaml
   :linenos:

.. vim: sts=3 sw=3 et syntax=rst:
