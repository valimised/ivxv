..  IVXV kogumisteenuse haldusteenuse kirjeldus

.. _app-processor:

Töötlemisrakendus
=================

Töötlemisrakendus on käsurearakendus e-urni kontrollimiseks ja edasiseks
töötlemiseks peale e-hääletamise lõppu.

Töötlemisrakenduse põhilised tööriistad on *check*, *squash*, *revoke* ja
*anonymize*, mis käivitatakse loetletud järjekorras vastavalt ette nähtud
valimisprotseduuridele.
Põhitööriistade sisendi hulgas on alati kas kogumisteenuse või eelmise tööriista
poolt väljastatud e-urn ja e-urni digitaalselt allkirjastatud räsi.
Väljundi hulgas on töötlemisetapi tulemuseks olev e-urn koos allkirjastamata
räsiga. Kuna rakendused käivitatakse internetiühenduseta arvutis, tuleb
räsifailid tõsta digitaalseks allkirjastamiseks välisesse seadmesse.
E-urni räsi arvutatakse funktsiooniga ``hex(sha256(<fail>))``.

Lisaks põhitööriistadele on rakendusel veel neli täiendavat tööriista:
*export*, *verify*, *stats* ja *statsdiff*.

Kõigi tööriistade kasutamine eeldab allkirjastatud usaldusjuure ja konkreetse
tööriista seadistuste olemasolu.
Faile väljastavatel tööriistadel tuleb seadistustes määrata väljundkausta
asukoht. Väljundkausta ei tohi käivitamise ajal olemas olla, selle loob rakendus.
Alljärgnevalt on kirjeldatud tööriistade seadistusi.

.. _processor-check:

E-urni töötlemine - verifitseerimine
------------------------------------

Kogumisteenusest väljastatud e-urni verifitseerimiseks kasutatakse tööriista
*check*. Urni verifitseeritakse usaldusjuure, valijate nimekirjade, ringkondade
nimekirja ja registreerimisteenuse väljundi vastu.

Verifitseerimise käigus kontrollitakse järgmiseid põhilisi omadusi:

* Ringkondade nimekirja ja valijate nimekirjade andmeterviklus ja
  kooskõlalisus;

* E-urni andmeterviklus;

* E-hääletajate valimisõigus e. kuuluvus valijate nimekirja (kontrollitakse
  juhul kui valijate nimekirjad on seadistustes kirjeldatud);

* E-urnis sisalduvate häälte vastavus digiallkirja vormingule;

* Registreerimisandmete andmeterviklus;

* E-urnis sisalduvate häälte vastavus registreerimisandmetega.

E-urni verifitseerimine on töömahukas protsess. 4-tuumalise *i7* protsessoriga
arvuti suudab ühe sekundi jooksul töödelda umbes 200 häält. Töötlemise jooksul
kuvatakse kasutajale edenemisriba, mille alusel on võimalik ennustada
töötlemisele kuluvat aega.

Suure e-urni verifitseerimisel võib olla tarvilik protsessile mälu juurde anda.
Selleks tuleb seada keskkonnamuutuja ``PROCESSOR_OPTS``. Näiteks 6 gigabaidi
mälu eraldamiseks ``PROCESSOR_OPTS=-Xmx6G``.

:check.ballotbox:
        Kogumisteenusest väljastatud e-urn.

:check.ballotbox_checksum:
        Kogumisteenusest väljastatud e-urni digitaalselt allkirjastatud räsi.

        Kui määramata, siis ei väljastata korrastatud e-urni järgmisteks
        etappideks. Kasulik mitte-lõpliku e-urni valimisaegseks kontrolliks.

:check.districts:
        Digitaalselt allkirjastatud ringkondade nimekiri.

:check.registrationlist:
        Registreerimisteenusest pärit registreerimisandmed. Kui määramata, siis
        ei kontrollita e-urnis sisalduvate häälte vastavust
        registreerimisandmetega.

:check.registrationlist_checksum:
        Registreerimisandmete digitaalselt allkirjastatud räsi. Võib puududa,
        kui ``registrationlist`` puudub.

:check.tskey:
        Registreerimispäringute verifitseerimiseks kasutatav kogumisteenuse
        avalik võti registreerimispäringute tegemise sertifikaadist.

:check.vlkey:
        Valijate nimekirjade verifitseerimiseks kasutatav avalik võti.
        Argument on kohustuslik, kui valijate nimekirjad on antud.

:check.voterlists:
        Valijate nimekirjade loend. Kui on määramata, siis e-hääletanute
        hääleõigust ei kontrollita.

:check.voterlists.path:
        Valijate nimekirja fail.

:check.voterlists.signature:
        Valijate nimekirja allkiri, mis on antud algoritmiga
        ``sha256WithRSAEncryption``.

:check.districts_mapping:
        Valijate nimekirjas oleva ringkonna ja jaoskonna teisendusfail
        (valikuline).

:check.election_start:
        Hääletamise algusaeg. Sellest varasema hääletusajaga hääli käsitletakse
        proovihäältena ning need lugemisele ei lähe.

:check.out:
        Tööriista väljundkaust. Sellesse kausta tekivad:

        #. Tervikluskontrolliga korrastatud e-urn :file:`bb-1.json`;

        #. Tervikluskontrolliga korrastatud e-urni räsi
           :file:`bb-1.json.sha256sum`;

        #. E-urni töötlemisvigade raport :file:`ballotbox_errors.txt`
           (valikuline);

        #. Valijate nimekirjade töötlemisvigade raport
           :file:`voterlist_errors.txt` (valikuline);

        #. *Log1* fail ehk vastvõetud hääled
           :file:`<valimise id>.<küsimuse id>.log1`.


:file:`processor.check.yaml`:

.. literalinclude:: config-examples/processor.check.yaml
   :language: yaml
   :linenos:

.. _processor-squash:

E-urni töötlemine - korduvhäälte tühistamine
--------------------------------------------

Korduvate e-häälte tühistamiseks kasutatakse tööriista *squash*.
Tööriista sisendiks on tööriista *check* poolt koostatud e-urn.
Korduvhäälte tühistamisel jäetakse alles iga hääletaja kõige hilisema hääl ja
eemaldatakse kõik varasemad hääled.

:squash.ballotbox:
        Tervikluskontrolliga korrastatud e-urn.

:squash.ballotbox_checksum:
        Tervikluskontrolliga korrastatud e-urni digitaalselt allkirjastatud
        räsi.

:squash.districts:
        Digitaalselt allkirjastatud ringkondade nimekiri.

:squash.out:
        Tööriista väljundkaust. Sellesse kausta luuakse:

        #. Korduvhäältest puhastatud e-urn :file:`bb-2.json`;

        #. Korduvhäältest puhastatud e-urni räsi :file:`bb-2.json.sha256sum`;

        #. E-hääletanute nimekiri JSON-vormingus :file:`ivoterlist.json`;

        #. E-hääletanute nimekiri PDF-vormingus :file:`ivoterlist.pdf`;

        #. Tühistamiste ja ennistamiste aruanne :file:`revocation-report.csv`;

        #. *Log2* fail ehk tühistatud hääled :file:`<valimise id>.<küsimuse
           id>.log2`.

:file:`processor.squash.yaml`:

.. literalinclude:: config-examples/processor.squash.yaml
   :language: yaml
   :linenos:


.. _processor-revoke:

E-urni töötlemine - häälte tühistamine ja ennistamine jaoskonnainfo põhjal
--------------------------------------------------------------------------

Häälte tühistamiseks ja ennistamiseks jaoskonnainfo põhjal kasutatakse tööriista
*revoke*. Tööriist saab sisendiks tööriista *squash* poolt koostatud e-urni ning
rakendab sellele sisendiks antud tühistus- ja ennistusnimekirjad.

:revoke.ballotbox:
        Korduvhäältest puhastatud e-urn.

:revoke.ballotbox_checksum:
        Korduvhäältest puhastatud e-urni digitaalselt allkirjastatud räsi.

:revoke.districts:
        Digitaalselt allkirjastatud ringkondade nimekiri.

:revoke.revocationlists:
        Tühistus- ja ennistusnimekirjade loend. Võib olla tühi.

:revoke.out:
        Tööriista väljundkaust. Sellesse kausta tekivad:

        #. Korduvhääletajate häältest puhastatud e-urn :file:`bb-3.json`;

        #. Korduvhääletajate häältest puhastatud e-urni räsi
           :file:`bb-3.json.sha256sum`;

        #. Tühistamiste ja ennistamiste aruanne :file:`revocation-report.csv`;

        #. E-hääletanute nimekiri JSON-vormingus :file:`ivoterlist.json``;

        #. *Log2* fail e. tühistatud hääled
           :file:`<valimise id>.<küsimuse id>.log2`.

:file:`processor.revoke.yaml`:

.. literalinclude:: config-examples/processor.revoke.yaml
   :language: yaml
   :linenos:


.. _processor-anonymize:

E-urni töötlemine - anonüümistamine
-----------------------------------

E-urni anonüümistamiseks kasutatakse tööriista *anonymize*.
Tööriist saab sisendiks tööriista *revoke* poolt koostatud e-urni ning eemaldab
sellest valijate info.

:anonymize.ballotbox:
        Korduvhääletajate häältest puhastatud e-urn.

:anonymize.ballotbox_checksum:
        Korduvhääletajate häältest puhastatud e-urni digitaalselt allkirjastatud
        räsi.

:anonymize.enckey:
        Krüpteerimise avaliku võtme faili asukoht (võtmerakenduse väljund).
        Võtit kasutatakse krüpteeritud häälte eelkontrolliks, eristamaks
        päriselt krüpteeritud hääli suvalisest binaarsest prügist.

:anonymize.out:
        Tööriista väljundkaust. Sellesse kausta luuakse:

        #. Hääletajate isikuandmetest puhastatud e-urn :file:`bb-4.json`;

        #. Hääletajate isikuandmetest puhastatud e-urni räsi
           :file:`bb-4.json.sha256sum`;

        #. *Log3* fail e. lugemisele läinud hääled
           :file:`<valimise id>.<küsimuse id>.log3`.


:file:`processor.anonymize.yaml`:

.. literalinclude:: config-examples/processor.anonymize.yaml
   :language: yaml
   :linenos:


Töötlemisrakenduse täiendavad tööriistad
----------------------------------------

Tööriist *verify*
*****************

*Verify* on lisavahend, millega saab verifitseerida digitaalselt allkirjastatud
konteineri allkirja ning kuvada konteineri andmed.

:verify.file:
        Verifitseeritav fail.


:file:`processor.verify.yaml`:

.. literalinclude:: config-examples/processor.verify.yaml
   :language: yaml
   :linenos:


Tööriist *export*
*****************

*Export* on lisavahend, millega saab eksportida kogumisteenusest väljastatud
e-urni seest täielikke digitaalselt allkirjastatud hääle konteinereid. On
võimalik eksportida nii kõiki hääli korraga, kui konkreetse valija hääli.

:export.ballotbox:
        Kogumisteenusest väljastatud e-urn.

:export.ballotbox_checksum:
        Kogumisteenusest väljastatud e-urni digitaalselt allkirjastatud räsi.

:export.voter_id:
        Valija identifikaator (valikuline).

:export.out:
        Tööriista väljundkaust. Sellesse kausta tekivad:

        #. E-urni töötlemisvigade raport :file:`ballotbox_errors.txt`
           (valikuline);

        #. E-urnist eksporditud häälte digitaalselt allkirjastatud konteinerid.


:file:`processor.export.yaml`:

.. literalinclude:: config-examples/processor.export.yaml
   :language: yaml
   :linenos:


Tööriist *stats*
****************

*Stats* on lisavahend, millega saab arvutada häälte ja hääletajate statistikat
kogumisteenusest saadud e-urni põhjal. Statistikat on võimalik piiritleda
ajavahemikuga ning väljundit on võimalik piiritleda koondandmetega kui ka
ringkondade kaupa. NB! Tööriist ei kontrolli digitaalallkirju, häälte
töötlemiseks tuleb kasutada *check*, *squash*, *revoke*, *anonymize*
töövoogu.

:stats.ballotbox:
        Kogumisteenusest väljastatud e-urn.

:stats.election_day:
        Valimispäev. Kõikide e-hääletanute vanused arvutatakse statistika
        tarbeks selle kuupäeva suhtes.

:stats.period_start:
        Statistikaperioodi algusaeg (valikuline). Sellest varasema
        hääletusajaga hääli statistikasse ei kaasata.

:stats.period_end:
        Statistikaperioodi lõppaeg (valikuline). Sellest hilisema hääletusajaga
        hääli statistikasse ei kaasata.

:stats.districts:
        Digitaalselt allkirjastatud ringkondade nimekiri.
        Argument on kohustuslik, kui valijate nimekirjad on antud.

:stats.vlkey:
        Valijate nimekirjade verifitseerimiseks kasutatav avalik võti.
        Argument on kohustuslik, kui valijate nimekirjad on antud.

:stats.voterlists:
        Valijate nimekirjade loend. Vajalik ringkondade kaupa statistika
        väljastamiseks. Kui on määramata, siis väljastatakse ainult
        koondstatistika.

:stats.voterlists.path:
        Valijate nimekirja fail.

:stats.voterlists.signature:
        Valijate nimekirja allkiri, mis on antud algoritmiga
        ``sha256WithRSAEncryption``.

:stats.out:
        Tööriista väljundkaust. Sellesse kausta tekivad:

        #. E-urni statistika JSON-vormingus :file:`stats.json`;

        #. E-urni statistika CSV-vormingus :file:`stats.csv`;

        #. E-urni töötlemisvigade raport :file:`ballotbox_errors.txt`
           (tekib vigade korral);

        #. Valijate nimekirjade töötlemisvigade raport
           :file:`voterlist_errors.txt` (tekib vigade korral).


:file:`processor.stats.yaml`:

.. literalinclude:: config-examples/processor.stats.yaml
   :language: yaml
   :linenos:


Tööriist *statsdiff*
********************

*Statsdiff* on lisavahend, millega saab arvutada kahe statistikafaili vahet.
Tulemuseks on kolmas statistikafail, mille kõik väärtused on pärit alusfailist,
kust on lahutatud võrreldava faili väärtused.

:statsdiff.compare:
        Statistika võrdluse alusfail JSON-vormingus.

:statsdiff.to:
        Võrreldav statistika fail JSON-vormingus.

:statsdiff.diff:
        Tööriista väljundfail. Sellesse faili salvestatakse statistikate vahe
        JSON-vormingus.


:file:`processor.statsdiff.yaml`:

.. literalinclude:: config-examples/processor.statsdiff.yaml
   :language: yaml
   :linenos:
