..  IVXV kogumisteenuse haldusteenuse kirjeldus

.. _app-key:

Võtmerakendus
=============

Võtmerakendus `key` koosneb tööriistadest *groupgen*, *init*, *testkey*,
*decrypt* ja *util*. Kõigi tööriistade kasutamine eeldab allkirjastatud
usaldusjuure ja konkreetse tööriista seadistuste olemasolu. Alljärgnevalt
kirjeldame konkreetsete tööriistade seadistusi.

.. _key-groupgen:

Häälte salastamise võtme spetsifikatsiooni valimine
--------------------------------------------------------------------------------

Kasutamaks ElGamali krüptosüsteemi häälte krüpteerimiseks, on oluline häälte
salastamise võtme spetsifikatsiooni valimine ehk kasutatavate rühma parameetrite
valimine, milles tehakse matemaatilisi operatsioone. Oluline on, et antud
parameetrid oleksid valitud läbipaistvalt, vältimaks tagauste olemasolu, mille
abil oleks võimalik ilma salajast võtit omamata krüpteeritud hääli avada.

Kuna turvalisuse jaoks peavad rühma parameetrid vastama teatud tingimustele,
siis nende valimiseks pole kiiret meetodit. Sobivate rühmaparameetrite
leidmiseks tuleb juhuslikult valida mingid parameetrid ja kontrollida, kas need
vastavad antud tingimustele.

Rühma parameetrite genereerimise protsessi on võimalik läbipaistvalt läbi viia
kahel viisil:

  #. Kasutades teadaolevaid defineeritud parameetreid
  #. Parameetreid avaliku algoritmi alusel deterministlikult genereerides


Teadaolevate parameetrite kasutamine
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Mitmed standardid ja rakendused on juba defineerinud parameetrid, mida on sobiv
kasutada ElGamal krüptosüsteemis. Kasutades laialt levinud parameetreid on
suurem tõenäosus, et neid on sõltumatult kontrollitud.

Üheks selliseks standardiks on [RFC3526]_, mis kasutab samuti deterministlikku
parameetrite genereerimist. Antud standardi korral saab kontrollida defineeritud
parameetrite korrektsust järgneva Sage skriptiga:

.. literalinclude:: genparam.py
   :language: python
   :linenos:


Uute parameetrite deterministlik ja kontrollitav genereerimine
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

ElGamal krüptosüsteemi jaoks sobilikke parameetreid saab genereerida kasutades
tööriista *groupgen*.

Võtmespetsifikatsiooni genereerimine on ajaliselt mahukas tegevus, mis võib
olenevalt riistvarast kesta tunde. Ühekordselt genereeritud rühm on kasutatav
mitmetel valimistel.

:groupgen.paramtype: ElGamal'i krüptosüsteemi töö aluseks oleva rühma
                     tüüp. Toetatud väärtused:

                     #. mod - jäägiklassiring Zp
                     #. ec  - elliptkõverad

:groupgen.length: ElGamal'i krüptosüsteemi töö aluseks olevat rühma iseloomustav
                  turvaparameeter. Jäägiklassiringide korral on sobiv väärtus
                  3072. Elliptkõveraid kasutades on toetatud kõver P-384, mille
                  kasutamiseks tuleb sisestada väärtus 384.

:groupgen.init_template: Asukoht, kuhu kirjutatakse rühma parameetrid. Väljund
                         sobib kasutamiseks võtme genereerimise seadistuse
                         koostamisel.

:groupgen.random_source: Juhuarvugeneraatori sisendiks kasutatavate allikate
                         loetelu. Vaata ka :numref:`random-gen`.


Kasutades juhuslike parameetrite leidmiseks juhuarvugeneraatorit, mille
algväärtus on üheselt defineeritav ning avalikustatud, võivad kolmandad
osapooled kontrollida, et avaldatud rühma parameetrid on esimesed sellised
leitud parameetrid, mis vastavad tingimustele. Näitekonfiguratsioon kasutab
DPRNG'd avaliku seemnefailiga. Vaata ka :numref:`random-gen`.

:file:`key.groupgen.yaml`:

.. literalinclude:: config-examples/key.groupgen.yaml
   :language: yaml
   :linenos:

Sellise seadistuse korral loetakse juhuarvugeneraatori algväärtus failist
``public_seed_file``. Oluline on, et sellisel juhul käivitatakse võtmerakendus
ühelõimelisena::

  $ key groupgen --conf usaldusjuur.asice --params key.groupgen.asice --threads 1

.. _key-init:

Häälte salastamise võtme genereerimine
--------------------------------------

Häälte salastamise võtme genereerimiseks kasutatakse võtmerakenduse tööriista
*init*. Võti genereeritakse seadistustes näidatud läviskeemiga MofN, mis
tähendab, et N võtmehaldurist peavad häälte dekrüpteerimisel osalema vähemalt M
haldurit, vastasel juhul ei ole dekrüpteerimine võimalik.


:init.identifier: Valimise unikaalne identifikaator.

:init.out:

       Võtmerakenduse tööriista *init* väljundkataloog. Sellesse kataloogi
       tekivad

       #. PEM vormingus allkirjavõtme sertifikaat (sign.pem)
       #. PEM vormingus krüpteerimisvõtme sertifikaat (enc.pem)
       #. PEM vormingus krüpteerimisvõti (pub.pem)
       #. DER vormingus krüpteerimisvõti (pub.der)

:init.skiptest: Võtmeosakute kontrolltestide vahelejätmine.

:init.fastmode: Kaartidele automaatne terminalide määramine. Vaikimise väärtus
                on tõene.

----

:init.paramtype: ElGamal krüptosüsteemi aluseks oleva rühma parameetrid, mis
                 ühtlasi määravad võtme turvataseme.

:init.paramtype.mod: Jäägiklassiringi määravad parameetrid kümnendesituses.
                     Parameetrid võib luua võtmerakenduse tööriista *groupgen*
                     kasutades.

:init.paramtype.mod.p: Jäägiklassiringi moodul.

:init.paramtype.mod.g: Jäägiklassiringi generaator.

----

:init.signaturekeylen: Võtmerakenduse poolt genereeritava allkirjastamise võtme
                       pikkus.

:init.signcn: Võtmerakenduse poolt loodava allkirjastamise sertifikaadi subjekti
              nimi (väli *CN*).

:init.signsn: Võtmerakenduse poolt loodava allkirjastamise sertifikaadi
              järjekorranumber.

:init.enccn: Võtmerakenduse poolt loodava krüpteerimise sertifikaadi subjekti
             nimi (väli *CN*).

:init.encsn: Võtmerakenduse poolt loodava krüpteerimise sertifikaadi
             järjekorranumber.

----

:init.required_randomness: Juhuslikkuse allikatest loetava entroopia kohustuslik
                           hulk baitides.

:init.random_source: Juhuarvugeneraatori sisendiks kasutatavate allikate
                     loetelu. Vaata ka :numref:`random-gen`.

----

:init.genprotocol: Võtme genereerimiseks kasutatava algoritmi ja läviskeemi
                   spetsifikatsioon.

:init.genprotocol.desmedt:

      Algoritmi Desmedt korral genereeritakse võti usaldatava osakujagaja poolt
      ehk võtmerakenduse mälus. Privaatvõtme osakud talletatakse kiipkaartidel.

      Täiendavalt tuleb määrata läviskeemi osaliste arv ja minimaalne kvoorum.

      Kaartide arv 7 - võimalikud kvoorumid 1,2,3,4 - soovitatav kvoorum 4
      Kaartide arv 8 - võimalikud kvoorumid 1,2,3,4 - soovitatav kvoorum 4
      Kaartide arv 9 - võimalikud kvoorumid 1,2,3,4,5 - soovitatav kvoorum 5

:init.genprotocol.desmedt.threshold: Läviskeemi M väärtus - kvoorum.

:init.genprotocol.desmedt.parties: Läviskeemi N väärtus.

:file:`key.init.yaml`:

.. literalinclude:: config-examples/key.init.yaml
   :language: yaml
   :linenos:



.. _key-testkey:

Häälte salastamise võtme testimine
----------------------------------

Häälte salastamise võtme testimine kontrollib võtme rekonstrueerimise võimekust
selliselt, et iga osak osaleb vähemalt kahes kvoorumis. Testimiseks on vajalik
kõigi osakute osalemine.

:testkey.identifier: Valimise unikaalne identifikaator.

:testkey.out: Krüpteerimise avaliku võtme asukoha kataloog.

:testkey.threshold: Testimiseks kasutatav lävi, sama mis võtme loomisel
                    spetsifitseeritud.

:testkey.parties: Testimiseks kasutatav osapoolte arv, sama mis võtme loomisel
                  spetsifitseeritud.

:testkey.fastmode: Kaartidele automaatne terminalide määramine. Vaikimise
                   väärtus on tõene.


:file:`key.testkey.yaml`:

.. literalinclude:: config-examples/key.testkey.yaml
   :language: yaml
   :linenos:


.. _key-decrypt:

E-häälte dekrüpteerimine
--------------------------------------

Elektrooniliste häälte dekrüpteerimiseks kasutatakse võtmerakenduse tööriista
*decrypt*. Dekrüpteerimise õnnestumiseks peab osalema läviskeemi poolt määratud
kvoorumi jagu võtmehaldureid. Kui rakendati skeemi 5of9, siis osaleb
dekrüpteerimisel täpselt 5 võtmehaldurit. Vähema arvu haldurite korral ei ole
dekrüpteerimine võimalik.

:decrypt.identifier:

        Valimise unikaalne identifikaator.

----

:decrypt.protocol:

:decrypt.protocol.recover:

      Algoritmi Desmedt korral genereeritakse võti usaldatava osakujagaja poolt
      ehk võtmerakenduse mälus. Privaatvõtme osakud talletatakse kiipkaartidel.

:decrypt.protocol.recover.threshold:

      Läviskeemi M väärtus - kvoorum, mis spetsifitseeriti võtme loomisel.

:decrypt.protocol.recover.parties:

      Läviskeemi N väärtus, mis spetsifitseeriti võtme loomisel.

----

:decrypt.anonballotbox:

      Töötlemisrakenduse või miksimisrakenduse poolt loodud e-valimiskast anonüümistatud
      häältega.

:decrypt.anonballotbox_checksum:

      Anonüümistatud häältega e-valimiskasti allkirjastatud SHA256 kontrollsummafail.

:decrypt.questioncount:

      Küsimuste arv anonüümistatud e-valimiskastis. Vaikimisi väärtus on 1.

:decrypt.candidates:

      Valimise valikute nimekiri allkirjastatud kujul.

:decrypt.districts:

      Valimise ringkondade nimekiri allkirjastatud kujul.

:decrypt.provable:

      Valikuline korrektse dekrüpteerimise tõestuse väljastamine. Vaikimisi
      väärtus on tõene.

:decrypt.check_decodable:

      Krüptogrammide korrektsuse kontrollimine enne dekrüpteerimist. Juhul kui
      krüptogrammide sisend ei tule usaldatud allikast, siis tuleb kontrollida
      krüptogrammide korrektsust. Usaldatud allikad on töötlemisrakendus ning
      miksija. Vaikimisi väärus on väär.

:decrypt.out:

      Võtmerakenduse tööriista *decrypt* väljundkataloog. Eduka dekrüpteerimise
      korral tekivad siia kausta:

      #. Elektroonilise hääletamise tulemus
      #. Elektroonilise hääletamise tulemuse signatuur
      #. Loend kehtetutest sedelitest
      #. Lugemistõend


:file:`key.decrypt.yaml`:

.. literalinclude:: config-examples/key.decrypt.yaml
   :language: yaml
   :linenos:


Pärast dekrüpteerimist on võimalik kontrollida väljastatud elektroonilise
hääletamise tulemuse signatuuri korrektsust. Selleks tuleb teha järgnevad
sammud:

1. Eraldada allkirja kontrollimise võti allkirjastamise sertifikaadist::

    openssl x509 -in initout/sign.pem -noout -pubkey > sign.pub

2. Kontrollida hääletamise tulemuse allkirja::

    openssl dgst -sha256 -sigopt rsa_padding_mode:pss -sigopt \
    rsa_pss_saltlen:32 -sigopt rsa_mgf1_md:sha256 -verify sign.pub \
    -signature decout/TESTCONF.tally.signature decout/TESTCONF.tally

Korrektse allkirja korral kuvatakse väärtust `Verified OK`.

Võtmerakenduse täiendavad tööriistad
------------------------------------

:util.listreaders: Loetle ühendatud kaardilugejad.

:file:`key.util.yaml`:

.. literalinclude:: config-examples/key.util.yaml
   :language: yaml
   :linenos:


.. _random-gen:

Juhuarvude genereerimine võtmerakenduses
----------------------------------------

Võtmerakenduse tööriistad *groupgen* ja *init* vajavad oma tööks juhuarve, mille
genereerimiseks on võimalik kasutada erinevaid entroopiaallikaid, mis
võtmerakenduse poolt üheks allikaks kombineeritakse.

Kombineerimisel on oluline, et säiliks sisendite sõltumatus, st. kombineeritud
väljund ei tohi olla kehvem ühestki sisendist. IVXV raamistikus toimub entroopia
kombineerimine SHAKE-256 muutuva väljundipikkusega räsifunktsiooni abil (XOF),
kasutades skeemi nagu on kirjeldatud [BDPA10]_.

Lõpliku pikkusega entroopiaallika kasutamisel loetakse kogu väärtus ning antakse
see SHAKE-256 sisendiks. Piiramata pikkusega allika lisamisel salvestatakse
selle viide kombineerija mällu.

Pärides kombineerijast töödeldud juhuslikkust, loetakse kõigepealt igast
salvestatud entroopia allikast sama palju baite ning antakse see SHAKE-256
sisendiks. Seejärel kopeeritakse SHAKE-256 isend, muudetakse kopeeritud
SHAKE-256 režiim lugemisele ning loetakse nõutud baitide jagu väljundit.


:random_source: Juhuarvugeneraatori sisendiks kasutatavate allikate loetelu.

:random_source.random_source_type: Juhuarvugeneraatori allika tüüp.

:random_source.random_source_path: Juhuarvugeneraatori allika seadistatav
                                   asukoht. Argument on valikuline sõltuvalt
                                   allika tüübist.

----

:random_source_type\: file: Entroopia lugemine failist.

:random_source_path\: `randomness_file`: Kasutatav fail.

----

:random_source_type\: system: Operatsioonisüsteemi poolt pakutav
                              entroopiaallikas (Linuxil `/dev/urandom`).

----

:random_source_type\: DPRNG: Deterministlik pseudojuhuslik generaator (DPRNG) on
                             mõeldud baidijadade genereerimiseks, kasutades
                             etteantud seemneväärtust. Sama seemneväärtuse
                             korral genereerib meetod alati sama jada.
:random_source_path\: `seed_file`: DPRNG seemneväärtus saadakse viidatud faili
                                   SHA256 räsides.

----

:random_source_type\: stream: Entroopia lugemine voogseadmelt.

:random_source_path\: `/dev/urandom`: Kasutatav seade.

----

:random_source_type\: user: Välise programmi käest üle sokli juhuslikkust hankiv
                            entroopiaallikas. Kasutusel on IVXV-spetsiifiline
                            protokoll.

:random_source_path\: `user_entropy.exe`: Tee programmini, mis tuleb
                                          juhuslikkuse hankimiseks käivitada.


Entroopiaallikate kombineerimine kirjeldatud viisil võimaldab realiseerida
erinevaid juhuarvude genereerimise stsenaariumeid. Näiteks häälte salastamise
võtme genereerimise korral on vaja tagada võtme konfidentsiaalsus ning olla
kindel, et genereerimisprotsess ei ole hiljem korratav. Näitekonfiguratsioon
kasutab välist programmi kasutaja sisendi lugemiseks ning süsteemset
juhuarvugeneraatorit:

:file:`rnd.init.yaml`:

.. literalinclude:: config-examples/rnd.init.yaml
   :language: yaml
   :linenos:


Rakendus käivitatakse mitmelõimelisena::

  $ key init --conf usaldusjuur.asice --params rnd.init.asice

Häälte salastamise võtme spetsifikatsioon on erinevalt häälte salastamise
võtmest avalik ning selle genereerimisel kasutatud juhuslikkuse võib samuti teha
avalikuks. Näitekonfiguratsioon kasutab DPRNG'd avaliku seemnefailiga.

:file:`rnd.groupgen.yaml`:

.. literalinclude:: config-examples/rnd.groupgen.yaml
   :language: yaml
   :linenos:

Kui eesmärk on genereerimisprotsessi korratavus, tuleb rakendus käivitada
ühelõimelisena::

  $ key groupgen --conf usaldusjuur.asice --params rnd.groupgen.asice --threads 1



.. [BDPA10] G. Bertoni, J. Daemen, M. Peeters, G. Van Assche: Sponge-Based
   Pseudo-Random Number Generators. CHES 2010: 33-47

.. [RFC3526] T. Kivinen, M. Kojo: More Modular Exponential (MODP) Diffie-Hellman
   groups for Internet Key Exchange (IKE). IETF RFC3526, 2003
