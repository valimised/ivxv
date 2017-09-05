..  IVXV kogumisteenuse haldusteenuse kirjeldus

IVXV seadistused valimise korraldamise protsessis
-------------------------------------------------

IVXV kasutamiseks valimise kontekstis
tuleb süsteem ja sellega seotud rakendused seadistada selliselt, et on
võimalik valijatelt häälte vastu võtmine ning nende käitlemine süsteemile seatud
terviklus-, konfidentsiaalsus ja käideldavusnõuete raamides. Tehniline dokument
annab ülevaate olulisimatest seadistustoimingutest ning on mõeldud täiendama
elektroonilise hääletamise käsiraamatu poolt kirjeldatud protseduurireeglite
täitmist.

Hääletamisperioodi algusele eelnevad andmed ja tegevused
********************************************************

Valimise üldparameetrid
  Valimise üldparameetrid määravad valimise unikaalse identifikaatori
  kasutamiseks kõigi seotud komponentide poolt, küsimuste arvu ning
  identifikaatorid, hääletamisperioodi alguse- ja lõpuaja ning hääle
  kontrollimise seadistuse. Valimise üldparameetrite spetsifikatsiooni
  käsitletakse käesolevas dokumendis.

Algne valijate nimekiri
  Algne valijate nimekiri on kohandatud vormingus fail, mille vorming ning
  seotud protokollid on defineeritud dokumendis #TODO-IVXV-liidesed. Eesti
  riiklike valimiste korral tuleb algne valijate nimekiri Rahvastikuregistrist.

Valikute nimekiri
  Valikute nimekiri on JSON vormingus fail, mille vorming ning seotud
  protokollid on defineeritud dokumendis #TODO-IVXV-liidesed. Eesti riiklike
  valimiste korral tuleb valikute nimekiri valimiste infosüsteemist.

Ringkondade nimekiri
  Ringkondade nimekiri on JSON vormingus fail, mille vorming ning seotud
  protokollid on defineeritud dokumendis #TODO-IVXV-liidesed. Eesti riiklike
  valimiste korral tuleb ringkondade nimekiri valimiste infosüsteemist.

Rakenduste ja kogumisteenuse usaldusjuur
  Rakenduste ja kogumisteenuse usaldusjuur defineerib
  sertifitseerimishierarhia(d), mille alusel IVXV süsteemi komponendid
  verifitseerivad digitaalallkirju. Eesti riiklike valimiste korral määrab
  usaldusjuure koosseisu Riigi Valimisteenistus. Usaldusjuure vormingut ning
  seotud protokolle käsitletakse peatükis :ref:`Usaldusjuur <usaldusjuur>`.

Kogumisteenuse tehniline seadistus
  Kogumisteenuse tehniline seadistus kirjeldab IVXV mikroteenuste seadistuse
  ning jagunemise erinevate üksuste vahel. Eesti riiklike valimiste korral
  leiab kogumisteenuse osutaja Riigi Valimisteenistus. Tehniline seadistus
  kooskõlastatakse valimiste omaniku ja kogumisteenuse osutaja vahel. Tehnilist
  seadistust käsitletakse peatükis :ref:`Kogumisteenuse tehniline seadistus <kt-tehniline-seadistus>`.

Kogumisteenuse võtmed ja sertifikaadid
  Kogumisteenuse mikroteenused suhtlevad omavahel TLS protokolli vahendusel.
  Vastavad sertifikaadid tuleb eksportida Valijarakendusse ja
  Kontrollrakendusse.

Häälte salastamise võtme spetsifikatsioon
  Häälte salastamise võtme jaoks kasutatav algoritm ning seotud tehnilised
  parameetrid fikseeritakse enne häälte salastamise võtme genereerimist. Võtme
  spetsifikatsiooni käsitletakse peatükis :ref:`key-groupgen`.

Enne hääletamisperioodi algust teostatakse lähtuvalt eelnevatest andmetest
järgmised tegevused:

#. :ref:`app-trust`
#. :ref:`key-groupgen`
#. :ref:`key-init`
#. Kogumisteenuse seadistamine
#. Valijarakenduse seadistamine
#. Kontrollrakenduse seadistamine

Hääletamisperioodi aegsed tegevused
***********************************

Valijate nimekirjade uuendused


Hääletamisperioodi lõpule järgnevad tegevused
*********************************************

#. :ref:`processor-check`
#. :ref:`processor-squash`
#. :ref:`processor-revoke`
#. :ref:`processor-anonymize`
#. Häälte miksimine (valikuline)

#. Miksimistõendi kontrollimine 

#. :ref:`key-decrypt`
#. :ref:`auditor-decrypt`
#. :ref:`auditor-mix`


.. vim: sts=3 sw=3 et:
