..  IVXV kogumisteenuse haldusteenuse kirjeldus


.. _ivxv-rakendused:

IVXV rakendused
===============

.. _app-install:

Rakenduste paigaldamine
--------------------------------------------------------------------------------

IVXV rakendused on:

* võtmerakendus `key` (:numref:`app-key`),
* töötlemisrakendus `processor` (:numref:`app-processor`),
* auditirakendus `auditor` (:numref:`app-auditor`).

IVXV rakendused on arendatud programmeerimiskeeles Java, kasutusel on Java 11.
Rakendused on testitud Windows 10 ja Ubuntu 20.04 platvormil kasutades OpenJDK-d
või Oracle Javat.

Rakendused tarnitakse ZIP-vormingus failidena::

  <rakendus>-<tarnenumber>.zip

Peale ZIP-faili lahti pakkimist tekib kataloogipuu::

   <rakendus>-<tarnenumber>
   |-- bin
   |   |-- <rakendus>
   |   |-- <rakendus.bat>
   |-- lib
   |   |-- *.jar

Kui kataloogitee `<rakendus>-<tarnenumber>/bin` panna `PATH`'i, saab rakendust
edaspidi käivitada käsurealt::

  $ <rakendus>

Rakendusi paigaldades tuleb arvestada, et kesksüsteemi protokollides kirjeldatud
raportid kasutavad ilma ajavööndita aja vormingut (`yyyymmddhhmmss`).
Ajavööndiga ajamärgendi (näiteks hääletamise aeg kogumisteenuselt saadud
e-valimiskastis) esitamiseks raportis teisendavad Java rakendused ajamärgendi esmalt
operatsioonisüsteemi ajavööndisse ning seejärel eemaldavad ajavööndi info.
Seetõttu tuleb rakendusi käivitavates masinates seadistada ajavöönd selliseks,
millise kohalikus ajas soovitakse ajamärgendeid näha.

.. _app-trust:

Rakenduste usaldusjuure kirjeldamine
------------------------------------

Rakenduste kasutamine eeldab digitaalselt allkirjastatud seadistuste kasutamist.
Allkirjade verifitseerimiseks vajalikud sertifikaadid tuleb rakendusele ette
anda usaldusjuure koosseisus. Usaldusjuur on samuti digitaalselt allkirjastatud.

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
elemendid on konteinerisse laaditud.


Näide
*****

:file:`ivxv.properties`:

.. literalinclude:: config-examples/ivxv.properties.real
   :linenos:


Rakenduste käivitamine
--------------------------------------------

Rakendusi käivitatakse käsurealt, nende toimimist juhitakse käsureaparameetrite
ja digitaalselt allkirjastatud seadistustega. Kõik rakendused väljastavad
vajadusel abiinfot::

  $ <rakendus> --help

  Rakendus 'rakendus'        - Rakendus

  Kasutamine:
    <rakendus> <tööriist> --conf <conf> [--params <params>] [--force <force>] [--quiet <quiet>] [--lang <lang>] [--container_threads <container_threads>] [--threads <threads>]
    <rakendus> <tööriist> -h | --help
    <rakendus> -h | --help

  Tööriistad:
    tool_foo         - Tegevuse FOO teostamine
    tool_bar         - Tegevuse BAR teostamine

  Käsurea argumendid:
    -h --help             - Abi
    -c --conf (*)         - Konfiguratsioon
    -p --params           - Tööriista parameetrid
    -f --force            - Ära küsi kasutajalt kinnitust
    -q --quiet            - Vaikne käivitusrežiim
    --lang                - Keel
    -ct --container_threads - Allkirjastatud konteinerite teegi poolt kasutatav lõimede arv (<= 0 korral dünaamiline)
    -t --threads          - Rakenduse poolt paralleeltöötluse korral kasutatav lõimede arv (<= 0 korral dünaamiline)
  Rakendus lõpetas töö ilma vigadeta

Rakenduste kasutamisel tuleb määrata konkreetne tööriist, usaldusjuur ning
seadistusfail::

  $ <rakendus> tool_foo --conf usaldusjuur.asice --params tool_foo.conf.asice

  Konfiguratsiooni laadimine failist usaldusjuur.asice
  Konfiguratsiooni allkirja kontrollimine
  Konfiguratsiooni allkirja on andnud NIMI NIMESTE
  Konfiguratsiooni allkirja andmise aeg on 24.12.2018 18:00
  Konfiguratsiooni allkiri on korrektne ja kehtiv

  FOO!

  Rakendus lõpetas töö ilma vigadeta

Juhised rakenduste tööriistade ning nende seadistusfailide koostamise kohta
antakse järgmistes peatükkides. Käsureaargumendid on kõigil rakendustel samad:

:-h --help:
    Abiinfo kuvamine kas rakenduse või konkreetse tööriista kohta.

:-c --conf (*):
    Digitaalselt allkirjastatud fail usaldusjuurega. Kohustuslik parameeter.

:-p --params:
    Digitaalselt allkirjastatud tööriista parameetrid.

:-f --force:
    Ära küsi kasutajalt kinnitust.

:-q --quiet:
    Vaikne käivitusrežiim.

:--lang:
    Juhul kui rakendus on kompileeritud mitmekeelsena, siis keele valik.
    Vaikimisi on rakendustes võimaldatud ainult eesti keel.

:-ct --container_threads:
    Allkirjastatud konteinerite teegi poolt kasutatav lõimede arv. Vaikimisi
    valitakse lõimede arv teegi poolt dünaamiliselt lähtudes saadaolevate
    tuumade arvust.

:-t --threads:
    Rakenduse poolt paralleeltöötluse korral kasutatav lõimede arv. Vaikimisi
    valitakse lõimede arv rakenduse poolt dünaamiliselt lähtudes saadaolevate
    tuumade arvust.


Rakendustest eksisteerivad nii tooteversioonid kui testversioonid.
Testrakendused on kohaldatud protseduuride efektiivseks testimiseks, kuid ei
sobi valimiste tegelikuks läbiviimiseks. Näiteks ei võimalda võtmerakenduse
testversioon kasutada kiipkaarte. Testversioonid rakendustest kuvavad
käivitamisel hoiatuse::

  ********************************************************************
  *                           !!! HOIATUS !!!                        *
  *                                                                  *
  * Rakendus on käivitatud arendusrežiimis ning rakenduse käitumine  *
  * võib erineda tavarežiimist.                                      *
  * Rakenduse käivitamiseks tavarežiimis tuleb rakendus ümber        *
  * kompileerida.                                                    *
  ********************************************************************

Rakenduste käivituskeskkonna parameetrid
----------------------------------------

Suure e-valimiskasti auditeerimisel, töötlemisel või dekrüpteerimisel, võib olla
tarvilik suurendada protsessi mälupiirangut.

Seda saab teha kasutades rakendusespetsiifilist keskkonnamuutujat
``{RAKENDUS}_OPTS``, mis defineerib täiendavad argumendid Java virtuaalmasinale.
``{RAKENDUS}`` on üks kolmest ``AUDITOR``, ``KEY`` või ``PROCESSOR``. Protsessi
mälupiirangu suurendamiseks tuleb kasutada argumenti ``-Xmx{N}G``, kus ``{N}``
on mälupiirangu suurus gigabaitides.

Näiteks 10 gigabaidi mälu eraldamiseks töötlemisrakendusele tuleb seada
``PROCESSOR_OPTS=-Xmx10G``.


.. list-table:: Rakenduste mälupiirangu parameetrid
   :header-rows: 1

   * - Rakendus
     - Vaikimisi mälupiirang
     - Keskkonnamuutuja
   * - Auditirakendus
     - 8GB
     - ``AUDITOR_OPTS``
   * - Töötlemisrakendus
     - 8GB
     - ``PROCESSOR_OPTS``
   * - Võtmerakendus
     - Puudub
     - ``KEY_OPTS``


Rakendused töötavad nii 32-bitise kui 64-bitise Java andmemudeliga, samas
efektiivseimaks toimimiseks tuleb rakendusi kasutada 64-bitisel platvormil
64-bitise Java andmemudeliga. Juhul kui rakendus ei suuda käivitamisel 64-bitist
mudelit tuvastada kuvatakse hoiatus::

  ********************************************************************
  *                           !!! HOIATUS !!!                        *
  *                                                                  *
  * 64-bitise Java andmemudeli tuvastamine ebaõnnestus. Rakendus on  *
  * vähemefektiivsem. Rakenduse jõudluse suurendamiseks tuleb        *
  * kasutada 64-bitise andmemudeliga Java keskkonda.                 *
  ********************************************************************

Juhul kui rakenduse mälupiirang on 4GB või rohkem, ei ole 32-bitise
andmemudeliga Java võimeline rakendust käivitama. Kuvatakse järgmine veateade::

  Invalid maximum heap size: -Xmx4G
  The specified size exceeds the maximum representable size.
  Error: Could not create the Java Virtual Machine.
  Error: A fatal exception has occurred. Program will exit.



