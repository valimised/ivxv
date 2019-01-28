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

* :token:`texts` - Kasutajaliideses kasutatavad tekstid
* :token:`errors` - Kasutajaliideses kasutatavad veateated
* :token:`colors` - Kasutajaliidese värvide koodid
* :token:`params` - Rakenduse tööks vajalikud parameetrid
* :token:`elections` - Igale küsimuse identifikaatorile vastav tekst
  kasutajaliideses

Kõiki seadistatavaid väärtusi näeb näidisseadistusest. Kõik väärtused
on kohustuslikud.

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
  loenduriga. Kui tekst sisaldab märgendit XX, siis asendatakse see
  märgend automaatselt ajaga mis on jäänud rakenduse sulgumiseni
  sekundites. Märgendi tekstist eemaldamise korral aega ei kuvata.

Näide
-----

.. literalinclude:: config-examples/conf_verification_template.json
   :language: json
