..  IVXV kogumisteenuse haldusteenuse kirjeldus

.. _kontroll:

Kontrollrakenduse seadistamine
==============================

Kontrollrakenduste seadistus on JSON formaadis.

#. Android-seadmed otsivad seadistust aadressilt
   "https://www.valimised.ee/kontrollrakendus/config.json".
#. iOS-seadmed otsivad seadistust aadressilt
   "https://www.valimised.ee/kontrollrakendus/config.json"

Reaalne sisu võib kahel seadistusel olla samasugune ka siis kui
võimalike erinevuste tekkimise jaoks kasutatakse ennetavalt erinevaid
URLe. Varasemalt on erinevate URLide kasutamist õigustanud nt. iOS
rakenduste pikem tarnetsükkel, mis nõuab testseadistuste
avalikustamist.

Seadistus koosneb viiest peamisest rühmast

* :token:`versions` - Rakenduse nõutud versioon
* :token:`texts` - Kasutajaliideses kasutatavad tekstid
* :token:`errors` - Kasutajaliideses kasutatavad veateated
* :token:`colors` - Kasutajaliidese värvide koodid
* :token:`params` - Rakenduse tööks vajalikud parameetrid
* :token:`elections` - Igale küsimuse identifikaatorile vastav tekst
  kasutajaliideses

Kõiki seadistatavaid väärtusi näeb näidisseadistusest. Kõik väärtused
on kohustuslikud.

Versioonide seadistamine
------------------------

Kontrollrakenduse versioon peab olema suurem või võrdne seadistuses määratud
versiooniga. Versioonid kuuluvad rühma :token:`versions`:

* :token:`android_version_code` - Android-rakenduse minimaalne versioonikood.
  Väärtus peab olema positiivne JSON täisarv.
* :token:`ios_bundle_version` - iOS-rakenduse minimaalne versioonisõne. Väärtus
  peab olema JSON sõne, mis koosneb punktidega eraldatud positiivsetest
  täisarvudest.

Parameetrite seadistamine
-------------------------

Rakenduse tööks vajalikud parameetrid kuuluvad rühma :token:`params`:

* :token:`verification_url` - Nimekiri kogumisteenuse hostinimedest
  või IP-aadressidest koos pordiga. Järjekord pole oluline. Väärtus
  peab olema JSON loend ka ühe URL-i puhul.
* :token:`verification_tls` - Nimekiri kogumisteenuse TLS
  sertifikaatidest PEM vormingus. Järjekord pole oluline. Väärtus peab
  olema JSON loend ka ühe sertifikaadi puhul.
* :token:`help_url` - Abiinfo vaate URL
* :token:`close_timeout` - Ajaaken, mil on kasutajal võimalik oma
  valikut näha enne rakenduse sulgumist. Millisekundites.
* :token:`close_interval` - Intervall, millega uuendatakse
  :token:`close_timeout` väärtust kasutajaliideses. Millisekundites.
* :token:`con_timeout_1` - Kogumisteenusega ühenduse saamise esimese
  katse ajapiirang. Millisekundites.
* :token:`con_timeout_2` - Kui esimese ringiga ei saadud ühendust
  ühegi kogumisteenuse instantsiga, proovitakse uuesti selle
  ajapiiranguga. Millisekundites.
* :token:`public_key` - Valimiste avalik võti, millega krüpteeritakse
  valijate hääli. PEM vormingus.
* :token:`tspreg_service_cert` - Ajatembeldusteenuse sertifikaat PEM
  vormingus.
* :token:`ocsp_service_cert` - OCSP-teenuse sertifikaadid PEM
  vormingus. Järjekord pole oluline. Väärtus peab olema JSON loend ka
  ühe väärtuse puhul. Kui väli on tühi, siis tuvastatakse OCSP
  responderi sertifikaat automaatselt.
* :token:`tspreg_client_cert` - Kogumisteenuse sertifikaat
  registreerimispäringute tegemiseks PEM vormingus.

Tekstide seadistamine
---------------------

Kasutajaliideses kasutatavad tekstid kuuluvad rühma :token:`texts`. Järgmised
tekstid on parametriseeritavad:

* :token:`lbl_close_timeout` - Kontrollrakenduse sulgemisteade koos
  loenduriga. Tekst peab sisaldama märgendit XX, mis asendatakse automaatselt
  rakenduse sulgemiseni jäänud ajaga sekundites.

Näide
-----

.. literalinclude:: config-examples/android-ios-config.json
   :language: json
