..  IVXV kogumisteenuse haldusteenuse kirjeldus

Töötlemisrakendus
=================

Töötlemisrakendus on käsurearakendus e-urni kontrollimiseks ja edasiseks
töötlemiseks peale e-hääletamise lõppu.

Töötlemisrakenduse põhilised tööriistad on *check*, *squash*, *revoke* ja
*anonymize*, mis käivitatakse loetletud järjekorras vastavalt ette nähtud
valimisprotseduuridele.
Põhitööriistade sisendi hulgas on alati kas koguja või eelmise tööriista
poolt väljastatud e-urn ja e-urni digitaalselt allkirjastatud räsi.
Väljundi hulgas on töötlemisetapi tulemuseks olev e-urn koos allkirjastamata
räsiga. Kuna rakendused käivitatakse internetiühenduseta arvutis, tuleb
räsifailid tõsta digitaalseks allkirjastamiseks välisesse seadmesse.
E-urni räsi arvutatakse funktsiooniga ``hex(sha256(<fail>))``.

Lisaks põhitööriistadele on rakendusel veel kaks täiendavat tööriista:
*export* ja *verify*.

Kõigi tööriistade kasutamine eeldab allkirjastatud usaldusjuure ja konkreetse
tööriista seadistuste olemasolu.
Tööriistadel, mis väljastavad faile, tuleb seadistustes määrata väljundkausta
asukoht, mida ei tohi käivitamise ajal eksisteerida.
Alljärgnevalt kirjeldame konkreetsete tööriistade seadistusi.

.. _processor-check:

E-urni töötlemine - verifitseerimine
------------------------------------

Kogujast väljastatud e-urni verifitseerimiseks kasutatakse tööriista *check*.
Urni verifitseeritakse usaldusjuure, valijate nimekirjade, ringkondade nimekirja
ja registreerimisteenuse väljundi vastu.
Verifitseerimise käigus kontrollitakse järgmiseid põhilisi omadusi:

* Ringkondade nimekirja ja valijate nimekirjade andmeterviklus ja kooskõlalisus,
* E-urni andmeterviklus,
* E-hääletajate valimisõigus e. kuuluvus valijate nimekirja (kontrollitakse
  juhul kui valijate nimekirjad on seadistustega antud),
* E-urnis sisalduvate häälte digiallkirja vormingule vastavus,
* Registreerimisandmete andmeterviklus,
* E-urnis sisalduvate häälte vastavus registreerimisandmetega.

E-urni verifitseerimine on töömahukas protsess. 4-tuumalise *i7* protsessoriga
arvuti suudab 1 sekundi jooksul töödelda umbes 200 häält. Töötlemise jooksul
kuvatakse kasutajale edenemisriba, mille alusel on võimalik ennustada
töötlemisele kuluvat aega.

:check.ballotbox:
        Kogujast väljastatud e-urn.

:check.ballotbox_checksum:
        Kogujast väljastatud e-urni digitaalselt allkirjastatud räsi.

:check.districts:
        Digitaalselt allkirjastatud ringkondade nimekiri.

:check.registrationlist:
        Registreerimisteenusest pärit registreerimisandmed.

:check.registrationlist_checksum:
        Registreerimisandmete digitaalselt allkirjastatud räsi.

:check.tskey:
        Registreerimispäringute verifitseerimiseks kasutatav koguja avalik võti.

:check.vlkey:
        Valijate nimekirjade verifitseerimiseks kasutatav avalik võti.
        Argument on kohustuslik, kui valijate nimekirjad on antud.

:check.voterlists:
        Valijate nimekirjade loend. Võib olla tühi, mis juhul e-hääletanute
        hääleõigust ei kontrollita.

:check.voterlists.path:
        Valijate nimekirja fail.

:check.voterlists.signature:
        Valijate nimekirja allkiri, mis on antud algoritmiga
        ``sha256WithRSAEncryption``.

:check.election_start:
        Hääletamise algus. Varasema hääletusajaga hääli käsitletakse
        proovihäältena ning need lugemisele ei lähe.

:check.out:
        Tööriista väljundkaust. Sellesse kausta tekivad:

        #. Tervikluskontrolliga korrastatud e-urn ``bb-1.json``,
        #. Tervikluskontrolliga korrastatud e-urni räsi ``bb-1.json.sha256sum``,
        #. E-urni töötlemisvigade raport ``ballotbox_errors.txt`` (valikuline),
        #. Valijate nimekirjade töötlemisvigade raport ``voterlist_errors.txt``
           (valikuline),
        #. *Log1* fail e. vastvõetud hääled ``<valimise id>.<küsimuse id>.log1``.


:file:`processor.check.yaml`:

.. literalinclude:: config-examples/processor.check.yaml
   :language: yaml
   :linenos:

.. _processor-squash:

E-urni töötlemine - korduvhäälte tühistamine
--------------------------------------------

Korduvate e-häälte tühistamiseks kasutatakse tööriista *squash*.
Tööriist saab sisendiks tööriista *check* poolt koostatud e-urni ning eemaldab
sellest iga hääletaja kohta kõik hääled peale hiliseima.

:squash.ballotbox:
        Tervikluskontrolliga korrastatud e-urn.

:squash.ballotbox_checksum:
        Tervikluskontrolliga korrastatud e-urni digitaalselt allkirjastatud
        räsi.

:squash.districts:
        Digitaalselt allkirjastatud ringkondade nimekiri.

:squash.out:
        Tööriista väljundkaust. Sellesse kausta luuakse:

        #. Korduvhäältest puhastatud e-urn ``bb-2.json``,
        #. Korduvhäältest puhastatud e-urni räsi ``bb-2.json.sha256sum``,
        #. E-hääletanute nimekiri *JSON* vormingus ``ivoterlist.json``,
        #. E-hääletanute nimekiri *PDF* vormingus ``ivoterlist.pdf``,
        #. Tühistamiste ja ennistamiste aruanne ``revocation-report.csv``,
        #. *Log2* fail e. tühistatud hääled ``<valimise id>.<küsimuse id>.log2``.


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

        #. Topelthääletajate häältest puhastatud e-urn ``bb-3.json``,
        #. Topelthääletajate häältest puhastatud e-urni räsi
           ``bb-3.json.sha256sum``,
        #. Tühistamiste ja ennistamiste aruanne ``revocation-report.csv``,
        #. E-hääletanute nimekiri *JSON* vormingus ``ivoterlist.json``,
        #. E-hääletanute nimekiri *PDF* vormingus ``ivoterlist.pdf``,
        #. *Log2* fail e. tühistatud hääled
           ``<valimise id>.<küsimuse id>.log2``.


:file:`processor.revoke.yaml`:

.. literalinclude:: config-examples/processor.revoke.yaml
   :language: yaml
   :linenos:


.. _processor-anonymize:

E-urni anonüümistamine
----------------------

E-urni anonüümistamiseks kasutatakse tööriista *anonymize*.
Tööriist saab sisendiks tööriista *revoke* poolt koostatud e-urni ning eemaldab
sellest valijate info.

:anonymize.ballotbox:
        Topelthääletajate häältest puhastatud e-urn.

:anonymize.ballotbox_checksum:
        Topelthääletajate häältest puhastatud e-urni digitaalselt allkirjastatud
        räsi.

:anonymize.out:
        Tööriista väljundkaust. Sellesse kausta luuakse:

        #. Topelthääletajate häältest puhastatud e-urn ``bb-4.json``,
        #. Topelthääletajate häältest puhastatud e-urni räsi
           ``bb-4.json.sha256sum``,
        #. *Log3* fail e. lugemisele läinud hääled
           ``<valimise id>.<küsimuse id>.log3``.


:file:`processor.anonymize.yaml`:

.. literalinclude:: config-examples/processor.anonymize.yaml
   :language: yaml
   :linenos:


Täiendavad tööriistad
---------------------

Tööriist *verify*
*****************

Tööriist *verify* on täiendav tööriist, millega saab verifitseerida etteantud
digitaalselt allkirjastatud konteineri digiallkirja ning kuvada välja konteineri
info.

:verify.file:
        Verifitseeritav fail.


:file:`processor.verify.yaml`:

.. literalinclude:: config-examples/processor.verify.yaml
   :language: yaml
   :linenos:


Tööriist *export*
*****************

Tööriist *export* on täiendav tööriist, millega saab eksportida kogujast saadud
e-urni seest täielikke digitaalselt allkirjastatud hääle konteinereid. On
võimalik eksportida nii kõiki hääli korraga, kui konkreetse valija hääli.

:export.ballotbox:
        Kogujast väljastatud e-urn.

:export.ballotbox_checksum:
        Kogujast väljastatud e-urni digitaalselt allkirjastatud räsi.

:export.voter_id:
        Valija identifikaator (valikuline).

:export.out:
        Tööriista väljundkaust. Sellesse kausta tekivad:

        #. E-urni töötlemisvigade raport ``ballotbox_errors.txt`` (valikuline),
        #. E-urnist eksporditud häälte digitaalselt allkirjastatud konteinerid.


:file:`processor.export.yaml`:

.. literalinclude:: config-examples/processor.export.yaml
   :language: yaml
   :linenos:

.. vim: sts=3 sw=3 et:
