..  IVXV kogumisteenuse haldusteenuse kirjeldus

IVXV seadistused valimise korraldamise protsessis
-------------------------------------------------

IVXV kasutamiseks valimise kontekstis tuleb süsteem ja sellega seotud
rakendused seadistada nii, et on võimalik valijatelt häälte vastuvõtmine ning
nende käitlemine vastavalt süsteemile seatud terviklus-, konfidentsiaalsus ja
käideldavusnõuetele. Käesolev tehniline dokument annab ülevaate olulisimatest
seadistustoimingutest ning on mõeldud täiendama elektroonilise hääletamise
käsiraamatu poolt kirjeldatud protseduurireeglite täitmist.

Seadistuste koostamiseks vajalikud andmed
*****************************************

Valimise üldparameetrid
  Valimise üldparameetrid määravad valimise unikaalse identifikaatori
  kasutamiseks kõigi seotud komponentide poolt, küsimuste arvu ning
  identifikaatorid, hääletamisperioodi alguse- ja lõpuaja ning hääle
  kontrollimise seadistuse. Valimise üldparameetrite spetsifikatsiooni
  käsitletakse käesolevas dokumendis.

Algne valijate nimekiri
  Algne valijate nimekiri on kohandatud vormingus fail, mille vorming ning
  seotud protokollid on defineeritud dokumendis "IVXV protokollid". Eesti
  riiklike valimiste korral tuleb algne valijate nimekiri Rahvastikuregistrist.

Valikute nimekiri
  Valikute nimekiri on JSON vormingus fail, mille vorming ning seotud
  protokollid on defineeritud dokumendis "IVXV protokollid". Eesti riiklike
  valimiste korral tuleb valikute nimekiri valimiste infosüsteemist.

Ringkondade nimekiri
  Ringkondade nimekiri on JSON vormingus fail, mille vorming ning seotud
  protokollid on defineeritud dokumendis "IVXV protokollid". Eesti riiklike
  valimiste korral tuleb ringkondade nimekiri valimiste infosüsteemist.

Rakenduste usaldusjuur
  Rakenduste usaldusjuur defineerib sertifitseerimishierarhia(d), mille alusel
  IVXV rakendused verifitseerivad digitaalallkirju. Eesti riiklike valimiste
  korral määrab usaldusjuure koosseisu Riigi Valimisteenistus. Rakenduste
  usaldusjuure vormingut käsitletakse peatükis :numref:`ivxv-rakendused`.

Kogumisteenuse usaldusjuur
  Kogumisteenuse usaldusjuur defineerib sertifitseerimishierarhia(d), mille
  alusel IVXV kogumisteenuse komponendid verifitseerivad digitaalallkirju. Eesti
  riiklike valimiste korral määrab usaldusjuure koosseisu Riigi
  Valimisteenistus. Kogumisteenuse usaldusjuure vormingut ning seotud protokolle
  käsitletakse peatükis :numref:`kogumisteenus`.

Kogumisteenuse tehniline seadistus
  Kogumisteenuse tehniline seadistus kirjeldab IVXV mikroteenuste seadistuse
  ning isendite jaotumise. Eesti riiklike valimiste korral leiab kogumisteenuse
  osutaja Riigi Valimisteenistus. Tehniline seadistus kooskõlastatakse
  valimiste omaniku ja kogumisteenuse osutaja vahel. Tehnilist seadistust
  käsitletakse peatükis :numref:`kt-technical`.

Kogumisteenuse võtmed ja sertifikaadid
  Kogumisteenuse mikroteenused suhtlevad omavahel TLS protokolli vahendusel.
  Vastavad sertifikaadid tuleb eksportida Valijarakendusse ja
  Kontrollrakendusse. Kogumisteenusega seotud võtmete loomist käsitletakse
  peatükis :numref:`kt-krypto`.

Häälte salastamise võtme spetsifikatsioon
  Häälte salastamise võtme jaoks kasutatav algoritm ning seotud tehnilised
  parameetrid fikseeritakse enne häälte salastamise võtme genereerimist. Võtme
  spetsifikatsiooni käsitletakse peatükis :numref:`key-groupgen`.

Hääletamisperioodile eelnevad tegevused
***************************************

Enne hääletamisperioodi algust teostatakse lähtuvalt eelnevatest andmetest
järgmised tegevused:

#. :numref:`app-install`
#. :numref:`app-trust`
#. :numref:`key-groupgen`
#. :numref:`key-init`
#. :numref:`key-testkey`
#. :numref:`kt-trust`
#. :numref:`kt-technical`
#. :numref:`kt-election`
#. Ringkondade nimekirja laadimine Kogumisteenusesse
#. Valikute nimekirja laadimine Kogumisteenusesse
#. Valijate nimekirja laadimine Kogumisteenusesse (algne)
#. :numref:`kt-management`
#. :numref:`valijarakendus`
#. :numref:`kontroll`

Hääletamisperioodi tegevused
****************************

#. Valijate nimekirjade laadimine Kogumisteenusesse (muudatused)

Hääletamisperioodile järgnevad tegevused
****************************************

E-urni töötlemine
^^^^^^^^^^^^^^^^^

#. :numref:`processor-check`
#. :numref:`processor-squash`
#. :numref:`processor-revoke`
#. :numref:`processor-anonymize`

Häälte miksimine
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

#. :numref:`mix-install`
#. :numref:`mix-mix`
#. :numref:`mix-verify`

Hääletamistulemuse väljaselgitamine ja andmeaudit
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

#. :numref:`key-decrypt`
#. :numref:`auditor-convert`
#. :numref:`auditor-mix`
#. :numref:`auditor-decrypt`
