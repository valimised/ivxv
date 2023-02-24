..  IVXV kogumisteenuse haldusteenuse kirjeldus

.. _app-processor:

Töötlemisrakendus
=================

Töötlemisrakendus on käsurearakendus e-valimiskasti kontrollimiseks ja edasiseks
töötlemiseks peale e-hääletamise lõppu.

Töötlemisrakenduse põhilised tööriistad on *check*, *squash*, *revoke* ja
*anonymize*, mis käivitatakse loetletud järjekorras vastavalt ette nähtud
valimisprotseduuridele.
Põhitööriistade sisendi hulgas on alati kas kogumisteenuse või eelmise tööriista
poolt väljastatud e-valimiskast ja e-valimiskasti digitaalselt allkirjastatud räsi.
Väljundi hulgas on töötlemisetapi tulemuseks olev e-valimiskast koos allkirjastamata
räsiga. Kuna rakendused käivitatakse internetiühenduseta arvutis, tuleb
räsifailid tõsta digitaalseks allkirjastamiseks välisesse seadmesse.
E-valimiskasti räsi arvutatakse funktsiooniga ``hex(sha256(<fail>))``.

Lisaks põhitööriistadele on rakendusel veel neli täiendavat tööriista:
*export*, *verify*, *stats* ja *statsdiff*.

Kõigi tööriistade kasutamine eeldab allkirjastatud usaldusjuure ja konkreetse
tööriista seadistuste olemasolu.
Faile väljastavatel tööriistadel tuleb seadistustes määrata väljundkausta
asukoht. Väljundkausta ei tohi käivitamise ajal olemas olla, selle loob rakendus.
Alljärgnevalt on kirjeldatud tööriistade seadistusi.

.. _processor-check:

E-valimiskasti töötlemine - verifitseerimine
--------------------------------------------

Kogumisteenusest väljastatud e-valimiskasti verifitseerimiseks kasutatakse tööriista
*check*. valimiskasti verifitseeritakse usaldusjuure, valijate nimekirjade, ringkondade
nimekirja ja registreerimisteenuse väljundi vastu.

Verifitseerimise käigus kontrollitakse järgmiseid põhilisi omadusi:

* Ringkondade nimekirja ja valijate nimekirjade andmeterviklus ja
  kooskõlalisus;

* E-valimiskasti andmeterviklus;

* E-hääletajate valimisõigus e. kuuluvus valijate nimekirja (kontrollitakse
  juhul kui valijate nimekirjad on seadistustes kirjeldatud);

* E-valimiskastis sisalduvate häälte vastavus digiallkirja vormingule;

* Registreerimisandmete andmeterviklus;

* E-valimiskastis sisalduvate häälte vastavus registreerimisandmetega.

E-valimiskasti verifitseerimine on töömahukas protsess. 4-tuumalise *i7* protsessoriga
arvuti suudab ühe sekundi jooksul töödelda umbes 200 häält. Töötlemise jooksul
kuvatakse kasutajale edenemisriba, mille alusel on võimalik ennustada
töötlemisele kuluvat aega.

:check.ballotbox:
        Kogumisteenusest väljastatud e-valimiskast.

:check.ballotbox_checksum:
        Kogumisteenusest väljastatud e-valimiskasti digitaalselt allkirjastatud räsi.

        Kui määramata, siis ei väljastata korrastatud e-valimiskasti järgmisteks
        etappideks. Kasulik mitte-lõpliku e-valimiskasti valimisaegseks kontrolliks.

:check.districts:
        Digitaalselt allkirjastatud ringkondade nimekiri.

:check.registrationlist:
        Registreerimisteenusest pärit registreerimisandmed. Kui määramata, siis
        ei kontrollita e-valimiskastis sisalduvate häälte vastavust
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

:check.voterlists_dir:
        Valijate nimekirjade loendi kaust. Kui on määramata, siis e-hääletanute
        hääleõigust ei kontrollita.

:check.voterlists:
        Valijate nimekirjade loend. Kui on määramata, siis e-hääletanute
        hääleõigust ei kontrollita.

:check.voterlists.path:
        Valijate nimekirja fail.

:check.voterlists.signature:
        Valijate nimekirja allkiri, mis on antud algoritmiga
        ``ecdsa-with-SHA256``.

:check.districts_mapping:
        Valijate nimekirjas oleva ringkonna ja jaoskonna teisendusfail
        (valikuline).

:check.election_start:
        Hääletamise algusaeg. Sellest varasema hääletusajaga hääli käsitletakse
        proovihäältena ning need lugemisele ei lähe.

:check.voterforeignehak:
        Alaliselt välisriigis elavate valijate ringkonnakuuluvuse tuvastamiseks
        kasutatav EHAK-kood. Vaikeväärtus "0000".

:check.out:
        Tööriista väljundkaust. Sellesse kausta tekivad:

        #. Tervikluskontrolliga korrastatud e-valimiskast :file:`<valimise id>-bb-1.json`;

        #. Tervikluskontrolliga korrastatud e-valimiskasti räsi
           :file:`<valimise id>-bb-1.json.sha256sum`;

        #. E-valimiskasti töötlemisvigade raport :file:`ballotbox_errors.txt`
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

E-valimiskasti töötlemine - korduvhäälte tühistamine
----------------------------------------------------

Korduvate e-häälte tühistamiseks kasutatakse tööriista *squash*.
Tööriista sisendiks on tööriista *check* poolt koostatud e-valimiskast.
Korduvhäälte tühistamisel jäetakse alles iga hääletaja kõige hilisema hääl ja
eemaldatakse kõik varasemad hääled.

:squash.ballotbox:
        Tervikluskontrolliga korrastatud e-valimiskast.

:squash.ballotbox_checksum:
        Tervikluskontrolliga korrastatud e-valimiskasti digitaalselt allkirjastatud
        räsi.

:squash.districts:
        Digitaalselt allkirjastatud ringkondade nimekiri.

:squash.enckey:
        Krüpteerimise avaliku võtme faili asukoht (võtmerakenduse väljund).
        Võtit kasutatakse krüpteeritud häälte eelkontrolliks, eristamaks
        päriselt krüpteeritud hääli suvalisest binaarsest prügist.

:squash.out:
        Tööriista väljundkaust. Sellesse kausta luuakse:

        #. Korduvhäältest puhastatud e-valimiskast :file:`<valimise id>-bb-2.json`;

        #. Korduvhäältest puhastatud e-valimiskasti räsi :file:`<valimise
           id>-bb-2.json.sha256sum`;

        #. E-hääletanute nimekiri JSON-vormingus :file:`<valimise
           id>-ivoterlist.json`;

        #. E-hääletanute nimekiri PDF-vormingus :file:`<valimise
           id>-ivoterlist.pdf`;

        #. Tühistamiste ja ennistamiste aruanne :file:`<valimise
           id>-revocation-report.csv`;

        #. Tühistamiste ja ennistamiste aruanne ilma isikuandmedeta
           :file:`<valimise
           id>-revocation-report.csv.anonymous`;

        #. *Log2* fail ehk tühistatud hääled :file:`<valimise id>.<küsimuse
           id>.log2`.

:file:`processor.squash.yaml`:

.. literalinclude:: config-examples/processor.squash.yaml
   :language: yaml
   :linenos:


.. _processor-revoke:

E-valimiskasti töötlemine - häälte tühistamine ja ennistamine jaoskonnainfo põhjal
----------------------------------------------------------------------------------

Häälte tühistamiseks ja ennistamiseks jaoskonnainfo põhjal kasutatakse tööriista
*revoke*. Tööriist saab sisendiks tööriista *squash* poolt koostatud e-valimiskasti ning
rakendab sellele sisendiks antud tühistus- ja ennistusnimekirjad.

:revoke.ballotbox:
        Korduvhäältest puhastatud e-valimiskast.

:revoke.ballotbox_checksum:
        Korduvhäältest puhastatud e-valimiskasti digitaalselt allkirjastatud räsi.

:revoke.districts:
        Digitaalselt allkirjastatud ringkondade nimekiri.

:revoke.revocationlists:
        Tühistus- ja ennistusnimekirjade loend. Võib olla tühi.

:revoke.out:
        Tööriista väljundkaust. Sellesse kausta tekivad:

        #. Korduvhääletajate häältest puhastatud e-valimiskast :file:`<valimise
           id>-bb-3.json`;

        #. Korduvhääletajate häältest puhastatud e-valimiskasti räsi
           :file:`<valimise id>-bb-3.json.sha256sum`;

        #. Tühistamiste ja ennistamiste aruanne :file:`<valimise
           id>-revocation-report.csv`;

        #. Tühistamiste ja ennistamiste aruanne ilma isikuandmedeta
           :file:`<valimise
           id>-revocation-report.csv.anonymous`;

        #. E-hääletanute nimekiri JSON-vormingus :file:`<valimise
           id>-ivoterlist.json``;

        #. *Log2* fail e. tühistatud hääled
           :file:`<valimise id>.<küsimuse id>.log2`.

:file:`processor.revoke.yaml`:

.. literalinclude:: config-examples/processor.revoke.yaml
   :language: yaml
   :linenos:


.. _processor-anonymize:

E-valimiskasti töötlemine - anonüümistamine
-------------------------------------------

E-valimiskasti anonüümistamiseks kasutatakse tööriista *anonymize*.
Tööriist saab sisendiks tööriista *revoke* poolt koostatud e-valimiskasti ning eemaldab
sellest valijate info.

:anonymize.ballotbox:
        Korduvhääletajate häältest puhastatud e-valimiskast.

:anonymize.ballotbox_checksum:
        Korduvhääletajate häältest puhastatud e-valimiskasti digitaalselt allkirjastatud
        räsi.

:anonymize.out:
        Tööriista väljundkaust. Sellesse kausta luuakse:

        #. Hääletajate isikuandmetest puhastatud e-valimiskast :file:`<valimis
           id>-bb-4.json`;

        #. Hääletajate isikuandmetest puhastatud e-valimiskasti räsi :file:`<valimise
           id>-bb-4.json.sha256sum`;

        #. *Log3* fail e. lugemisele läinud hääled :file:`<valimise
           id>.<küsimuse id>.log3`.


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
e-valimiskasti seest täielikke digitaalselt allkirjastatud hääle konteinereid. On
võimalik eksportida nii kõiki hääli korraga, kui konkreetse valija hääli.

:export.ballotbox:
        Kogumisteenusest väljastatud e-valimiskast.

:export.ballotbox_checksum:
        Kogumisteenusest väljastatud e-valimiskasti digitaalselt allkirjastatud räsi.

:export.voter_id:
        Valija identifikaator (valikuline).

:export.out:
        Tööriista väljundkaust. Sellesse kausta tekivad:

        #. E-valimiskasti töötlemisvigade raport :file:`ballotbox_errors.txt`
           (valikuline);

        #. E-valimiskastist eksporditud häälte digitaalselt allkirjastatud konteinerid.


:file:`processor.export.yaml`:

.. literalinclude:: config-examples/processor.export.yaml
   :language: yaml
   :linenos:


Tööriist *stats*
****************

*Stats* on lisavahend, millega saab arvutada häälte ja hääletajate statistikat
e-valimiskasti põhjal. Statistikat on võimalik piiritleda ajavahemikuga ning väljundit
on võimalik piiritleda koondandmetega kui ka ringkondade kaupa. NB! Tööriist ei
kontrolli digitaalallkirju, häälte töötlemiseks tuleb kasutada *check*,
*squash*, *revoke*, *anonymize* töövoogu.

:stats.ballotbox:
        E-valimiskast, mille põhjal statistika koostada. Kui faili laiendiks on
        ``.json``, siis peab see olema olema töödeldud e-valimiskast. Vastasel juhul
        peab see olema kogumisteenusest väljastatud e-valimiskast.

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
        Digitaalselt allkirjastatud ringkondade nimekiri. Vajalik ringkondade
        kaupa statistika väljastamiseks. Kui on määramata, siis väljastatakse
        ainult koondstatistika.

:stats.vlkey:
        Valijate nimekirjade verifitseerimiseks kasutatav avalik võti.
        Argument on kohustuslik valijate nimekirjade kasutamise korral.

:stats.voterlists:
        Valijate nimekirjade loend. Vajalik kogumisteenusest väljastatud
        e-valimiskastist valija ringkonna tuvastamiseks.

        Argument on kohustuslik, kui e-valimiskast on väljastatud kogumisteenusest ja
        statistikat väljastatakse ringkondade kaupa.

:stats.voterlists.path:
        Valijate nimekirja fail.

:stats.voterlists.signature:
        Valijate nimekirja allkiri, mis on antud algoritmiga
        ``ecdsa-with-SHA256``.

:check.voterforeignehak:
        Alaliselt välisriigis elavate valijate ringkonnakuuluvuse tuvastamiseks
        kasutatav EHAK-kood. Vaikeväärtus "0000".

:stats.out:
        Tööriista väljundkaust. Sellesse kausta tekivad:

        #. E-valimiskasti statistika JSON-vormingus :file:`<valimise id>-stats.json`
           (:file:`ELECTION-stats.json` kui valimist ei suudeta tuvastada);

        #. E-valimiskasti statistika CSV-vormingus :file:`<valimise id>-stats.csv`
           (:file:`ELECTION-stats.csv` kui valimist ei suudeta tuvastada);

        #. E-valimiskasti töötlemisvigade raport :file:`ballotbox_errors.txt`
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

.. _processor-checkAndSquash:

E-valimiskasti töötlemine - verifitseerimine ja korduvhäälte tühistamine
------------------------------------------------------------------------

Antud tööriist teostab nii verifitseerimist, kui ka korduvhäälte tühistamist.
Rohkem infot teostavate operatsioonide kohta leidub alapeatükkides:

* *E-valimiskasti töötlemine - verifitseerimine*
* *E-valimiskasti töötlemine - korduvhäälte tühistamine*

:checkAndSquash.ballotbox:
        Kogumisteenusest väljastatud e-valimiskast.

:checkAndSquash.ballotbox_checksum:
        Kogumisteenusest väljastatud e-valimiskasti digitaalselt allkirjastatud räsi.

        Kui määramata, siis ei väljastata korrastatud e-valimiskasti järgmisteks
        etappideks. Kasulik mitte-lõpliku e-valimiskasti valimisaegseks kontrolliks.

:checkAndSquash.districts:
        Digitaalselt allkirjastatud ringkondade nimekiri.

:checkAndSquash.registrationlist:
        Registreerimisteenusest pärit registreerimisandmed. Kui määramata, siis
        ei kontrollita e-valimiskastis sisalduvate häälte vastavust
        registreerimisandmetega.

:checkAndSquash.registrationlist_checksum:
        Registreerimisandmete digitaalselt allkirjastatud räsi. Võib puududa,
        kui ``registrationlist`` puudub.

:checkAndSquash.tskey:
        Registreerimispäringute verifitseerimiseks kasutatav kogumisteenuse
        avalik võti registreerimispäringute tegemise sertifikaadist.

:checkAndSquash.vlkey:
        Valijate nimekirjade verifitseerimiseks kasutatav avalik võti.
        Argument on kohustuslik, kui valijate nimekirjad on antud.

:checkAndSquash.voterlists_dir:
        Valijate nimekirjade loendi kaust. Kui on määramata, siis e-hääletanute
        hääleõigust ei kontrollita.

:checkAndSquash.voterlists:
        Valijate nimekirjade loend. Kui on määramata, siis e-hääletanute
        hääleõigust ei kontrollita.

:checkAndSquash.voterlists.path:
        Valijate nimekirja fail.

:checkAndSquash.voterlists.signature:
        Valijate nimekirja allkiri, mis on antud algoritmiga
        ``ecdsa-with-SHA256``.

:checkAndSquash.districts_mapping:
        Valijate nimekirjas oleva ringkonna ja jaoskonna teisendusfail
        (valikuline).

:checkAndSquash.election_start:
        Hääletamise algusaeg. Sellest varasema hääletusajaga hääli käsitletakse
        proovihäältena ning need lugemisele ei lähe.

:checkAndSquash.voterforeignehak:
        Alaliselt välisriigis elavate valijate ringkonnakuuluvuse tuvastamiseks
        kasutatav EHAK-kood. Vaikeväärtus "0000".

:checkAndSquash.enckey:
        Krüpteerimise avaliku võtme faili asukoht (võtmerakenduse väljund).
        Võtit kasutatakse krüpteeritud häälte eelkontrolliks, eristamaks
        päriselt krüpteeritud hääli suvalisest binaarsest prügist.

:checkAndSquash.out:
        Tööriista väljundkaust. Sellesse kausta tekivad:

        #. Korduvhäältest puhastatud e-valimiskast :file:`<valimise id>-bb-2.json`;

        #. Korduvhäältest puhastatud e-valimiskasti räsi :file:`<valimise
           id>-bb-2.json.sha256sum`;

        #. E-hääletanute nimekiri JSON-vormingus :file:`<valimise
           id>-ivoterlist.json`;

        #. E-hääletanute nimekiri PDF-vormingus :file:`<valimise
           id>-ivoterlist.pdf`;

        #. Tühistamiste ja ennistamiste aruanne :file:`<valimise
           id>-revocation-report.csv`;

        #. Tühistamiste ja ennistamiste aruanne ilma isikuandmedeta
           :file:`<valimise
           id>-revocation-report.csv.anonymous`;

        #. *Log1* fail ehk vastvõetud hääled
           :file:`<valimise id>.<küsimuse id>.log1`.

        #. *Log2* fail ehk tühistatud hääled :file:`<valimise id>.<küsimuse
           id>.log2`.

        #. E-valimiskasti töötlemisvigade raport :file:`ballotbox_errors.txt`
           (valikuline);

        #. Valijate nimekirjade töötlemisvigade raport
           :file:`voterlist_errors.txt` (valikuline);


:file:`processor.checkAndSquash.yaml`:

.. literalinclude:: config-examples/processor.checkAndSquash.yaml
   :language: yaml
   :linenos:

.. _processor-revokeAndAnonymize:

E-valimiskasti töötlemine - häälte tühistamine, ennistamine jaoskonnainfo põhjal ja anonüümistamine
---------------------------------------------------------------------------------------------------

Häälte tühistamiseks, ennistamiseks jaoskonnainfo põhjal ning anonüümistamiseks
kasutatakse tööriista *revokeAndAnonymize*. Tööriist saab sisendiks tööriista *squash*
või *checkAndSquash* poolt koostatud e-valimiskasti ning rakendab sellele sisendiks antud
tühistus- ja ennistusnimekirjad.

:revokeAndAnonymize.ballotbox:
        Korduvhäältest puhastatud e-valimiskast.

:revokeAndAnonymize.ballotbox_checksum:
        Korduvhäältest puhastatud e-valimiskasti digitaalselt allkirjastatud räsi.

:revokeAndAnonymize.districts:
        Digitaalselt allkirjastatud ringkondade nimekiri.

:revokeAndAnonymize.revocationlists:
        Tühistus- ja ennistusnimekirjade loend. Võib olla tühi.

:revokeAndAnonymize.out:
        Tööriista väljundkaust. Sellesse kausta tekivad:

        #. Korduvhääletajate häältest puhastatud ning anonüümistamistatud e-valimiskast
           :file:`<valimise id>-bb-4.json`;

        #. Korduvhääletajate häältest puhastatud ning anonüümistamistatud e-valimiskasti räsi
           :file:`<valimise id>-bb-4.json.sha256sum`;

        #. Tühistamiste ja ennistamiste aruanne :file:`<valimise
           id>-revocation-report.csv`;

        #. Tühistamiste ja ennistamiste aruanne ilma isikuandmedeta
           :file:`<valimise
           id>-revocation-report.csv.anonymous`;

        #. E-hääletanute nimekiri JSON-vormingus :file:`<valimise
           id>-ivoterlist.json``;

        #. *Log2* fail e. tühistatud hääled
           :file:`<valimise id>.<küsimuse id>.log2`.

        #. *Log3* fail e. lugemisele läinud hääled :file:`<valimise
           id>.<küsimuse id>.log3`.

:file:`processor.revokeAndAnonymize.yaml`:

.. literalinclude:: config-examples/processor.revokeAndAnonymize.yaml
   :language: yaml
   :linenos:
