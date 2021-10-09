..  IVXV seadistuste koostamise juhend

.. _kt-election:

Kogumisteenusele valimise seadistuse koostamine
-----------------------------------------------

Valimise seadistus määrab ühe valimise seadistuse.

Valimise seadistuse koostab valimiste korraldaja.
Seadistusfaili nimi peab alati lõppema stringiga :file:`election.yaml`.
Failinime võimalik eesliide peab alati lõppema punktiga.

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

:voting:

        Hääle esitamise parameetrite alamblokk.

:voting.ratelimitstart:

        Valija poolt esitatud häälte kogus, mille järel rakendub talle
        hääletamissageduse piirang. Arvesse lähevad ka vigased hääled, kuna
        muidu saaks nendega süsteemi koormata piirangust hoolimata. Välja
        puudumise või väärtuse 0 korral rakendub piirang alates esimesest
        häälest.

:voting.ratelimitminutes:

        Aeg minutites, mis peab jääma kahe hääle esitamise vahele, kui valijale
        on rakendatud hääletamissageduse piirang. Välja puudumise või väärtuse
        0 korral on hääletamissageduse piirangud välja lülitatud.

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

:verification.latestonly:

        Tõeväärtus, kas kontrollida saab ainult valija viimati antud häält. Kui
        väärtus on väär või puudu, siis saab kontrollida kõiki valija hääli
        (teiste piirangute raames).

----

:voterforeignehak:

        Alaliselt välisriigis elavate valijate ringkonnakuuluvuse tuvastamiseks
        kasutatav EHAK-kood.

        Kui parameeter on määratud, siis alaliselt välisriigis elavad valijad
        kuuluvad ringkondadesse, kuhu kuulub ka parameetris määratud
        EHAK-koodile vastav haldusüksus. Täiendavalt peab ringkondade nimekiri
        sisaldama sellise EHAK-koodiga haldusüksust.

        Kui parameeter on määramata, siis kasutatakse EHAK-koodi "0000". Kui
        parameeter on määramata ning ringkondade nimekiri ei sisalda
        vaikekoodile vastavat haldusüksust, siis valijate nimekiri ei tohi
        sisaldada alaliselt välisriigis elavaid valijaid.

:ignorevoterlist:

        Ringkonna identifikaator, mille valikud esitada kõigile valijatele.
        Kui see väärtus ei ole tühi, siis kogumisteenus ei kasuta valijate
        nimekirja ning esitab kõigile valijatele väärtusega määratud ringkonna
        valikud ja lubab hääletada kõigil, kellel õnnestub isikutuvastus ning
        hääle allkirja kontrollimine.

----

:voterlist:

        Kohustuslik väli.
        Valijate nimekirjade kontrollimise parameetrid.

:voterlist.key:

        Kohustuslik väli.
        ECDSA-võtmepaari avalik võti valijate nimekirjade allkirja
        kontrollimiseks.

:vis:
        Alamblokk, mis sisaldab Valimiste Infosüsteemi seadistust.

:vis.url:

        Kohustuslik väli.
        Valimiste Infosüsteemi URL.

:vis.ca:

        Valimiste Infosüsteemi TLS-sertifikaadi usaldusahel.

----

:auth:

        Kohustuslik väli.
        Alamblokk, mis sisaldab valija tuvastamise seadistust.

:auth.ticket:

        Alamblokk, mis sisaldab piletipõhise valija tuvastamise seadistust.

        Piletipõhist valija tuvastamist kasutatakse Mobiil-ID puhul, kus
        ``mid`` teenus tuvastab valija ning väljastab talle pileti, millega
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

:auth.tls.ocsp.retry:

        Valija TLS-klientsertifikaatide oleku kontrolli korduvkatsete arv.
        Juhul kui sertifikaadi oleku kontroll ebaõnnestub võrgu- või serverivea
        tõttu, saab seda automaatselt korrata. Välja väärtus määrab katsete
        arvu, mis tehakse lisaks algsele päringule. Seega kui väärtus on 1,
        siis tehakse kokku maksimaalselt kaks päringut. Välja puudumise või
        väärtuse 0 korral automaatseid korduvkatseid ei sooritata.

----

:identity:

        Tuvastatud valija X.500 eraldusnimest unikaalse identifikaatori
        tuletamise meetod. Hetkel toetatud valikud ``commonname``,
        ``serialnumber`` ja ``pnoee``.

        Eesti elektrooniliste isikut tõendavate dokumentide korral on
        ``commonname`` puhul identifikaator kujul "PERENIMI,EESNIMI,ISIKUKOOD"
        ning teiste valikute teise puhul "ISIKUKOOD".

        Kui ``serialnumber`` tagastab eraldusnimest ``serialNumber`` välja
        muutmata kujul, siis ``pnoee`` eemaldab sellelt enne mittekohustusliku
        "PNOEE-" eesliite. Viimane on vastavuses standardi ETSI EN 319 412-1
        jaotisega 5.1.3 Eesti isikukoodide jaoks, kuid lubab ka standardile
        mittevastavaid seerianumbreid.


----

:age:

        Alamblokk, mis sisaldab valija vanuse kontrolli seadistust. Kui see
        blokk puudub, siis valija vanust ei kontrollita.

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

:vote.bdoc.profile:

        Kohustuslik väli.
        Häälte allkirjadelt nõutav BDOC profiil. Toetatud valikud on ``BES``
        (põhiprofiil kirjeldatud BDOC spetsifikatsiooni jaotises 5), ``TM``
        (ajamärkidega profiil kirjeldatud BDOC spetsifikatsiooni jaotises 6.1)
        ja ``TS`` (ajatemplitega profiil kirjeldatud BDOC spetsifikatsiooni
        jaotises 6.2).

        See peaks olema ``BES``, kuna kõikide allkirjastamisvahendite puhul ei
        ole sissetulev hääl kvalifitseeritud (nt Eesti ID-kaart). Kogumisteenus
        kvalifitseerib häältel olevad allkirjad ise (vt ``qualification``).

----

:mid:

        Alamblokk, mis sisaldab Mobiil-ID teenusepakkuja seadistust.

:mid.url:

        Kohustuslik väli.
        Mobiil-ID teenusepakkuja asukoht.

:mid.relyingpartyuuid:

        Kohustuslik väli.
        Mobiil-ID teenusepakkujaga kokkulepitud kliendi identifikaator.

:mid.relyingpartyname:

        Kohustuslik väli.
        Mobiil-ID teenusepakkujaga kokkulepitud kliendi nimi.

:mid.language:

        Kohustuslik väli.
        Mobiil-ID kasutajale kuvatavate sõnumite keel. Võimalikud väärtused
        ``EST``, ``ENG``, ``RUS`` ja ``LIT``.

:mid.authmessage:

        Kohustuslik väli.
        Sõnum, mida Mobiil-ID kasutajale kuvada autentimise käigus.

:mid.signmessage:

        Kohustuslik väli.
        Sõnum, mida Mobiil-ID kasutajale kuvada allkirjastamise käigus.

:mid.messageformat:

        Autentimise ning allkirjastamise käigus kasutajale kuvatava sõnumi
        vorming. Võimalikud väärtused ``GSM-7`` (Mobiil-ID poolt kasutatav
        vaikeväärtus) ja ``UCS-2``.

:mid.authchallengesize:

        Autentimise käigus Mobiil-ID teenusele saadetava pretensiooni pikkus.
        Võimalikud väärtused ``32`` (vaikeväärtus), ``48`` ja ``64``.

:mid.statustimeoutms:

        Parameeter, mis edastatakse autentimise ja allkirjastamise staatuse
        päringu korral Mobiil-ID teenusele ning millega saab kontrollida, kui
        kaua ootab Mobiil-ID teenus kasutaja poolse tegevuse lõpptulemust, enne
        kui vastab, et tegevus on pooleli. Väärtuse puudumisel oodatakse
        võimalikult vähe: täpne aeg sõltub Mobiil-ID teenusest.

        Antud parameetri abil saab vähendada ühe autentimise või allkirjastamise
        käigus Mobiil-ID teenusele saadetavate päringute arvu.

:mid.roots:

        Kohustuslik väli.
        Mobiil-ID sertifikaatide usaldusjuured.

:mid.intermediates:

        Mobiil-ID sertifikaatide vahesertifikaadid. Mobiil-ID autentimiseks
        peab nende sertifikaatide abil olema võimalik luua ahel Mobiil-ID
        sertifikaadist usaldusjuureni.

:mid.ocsp:

        Alamblokk, mis sisaldab valija Mobiil-ID sertifikaatide oleku
        kontrollimise seadistust.

:mid.ocsp.url:

        Kohustuslik väli.
        Valija Mobiil-ID sertifikaatide kehtivuskinnitusteenuse aadress.

:mid.ocsp.responders:

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

:qualification.*.protocol:

        Kohustuslik väli.
        Kvalifitseeriva päringu protokoll. Hetkel toetatud ``ocsp``
        (harilik OCSP), ``ocsptm`` (OCSP ajamärgendina), ``tsp`` (PKIX
        ajatempel) ja ``tspreg`` (PKIX ajatempel
        registreerimistõendina).

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
        sertifikaat, ning on lubatud OCSP vastuste signeerimiseks. AIA
        loogika kasutamise korral võib väli jääda tühjaks.

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

:qualification.*.conf.retry:

        Kvalifitseeriva päringu korduvkatsete arv. Juhul kui päring ebaõnnestub
        võrgu- või serverivea tõttu, saab seda automaatselt korrata. Välja
        väärtus määrab katsete arvu, mis tehakse lisaks algsele päringule.
        Seega kui väärtus on 1, siis tehakse kokku maksimaalselt kaks päringut.
        Välja puudumise või väärtuse 0 korral automaatseid korduvkatseid ei
        sooritata.

Näide
*****

:file:`example.election.yaml`:

.. literalinclude:: config-examples/example.election.yaml
   :language: yaml
   :linenos:

.. vim:syntax=rst:
