..  IVXV kogumisteenuse haldusteenuse kirjeldus

Kontrollrakendus
================

Kontrollrakenduste seadistus on JSON formaadis.
Android-seadmed otsivad seadistust aadressilt "https://eh.valimised.ee/apps/config_android.json".
iOS-seadmed otsivad seadistust aadressilt "https://eh.valimised.ee/apps/config_ios.json"

Reaalne sisu võib olla neil samasugune, kahe URL'i olemasolu on eelkõige võimalike erinevuste tekkimise jaoks
ennatlik valik.

Seadistus koosneb viiest peamisest rühmast

* :token:`texts` - Kasutajaliideses kasutatavad tekstid
* :token:`errors` - Kasutajaliideses kasutatavad veateated
* :token:`colors` - Kasutajaliidese värvide koodid
* :token:`params` - Rakenduse tööks vajalikud parameetrid
* :token:`elections` - Igale küsimuse identifikaatorile vastav tekst kasutajaliideses

Kõiki seadistatavaid väärtusi näeb näidisseadistusest. Kõik väärtused on kohustuslikud.

:token:`params`
---------------

* :token:`verification_url` - Nimekiri kogumisteenuse hostinimedest või IP-aadressidest koos pordiga. Järjekord pole oluline. Väärtus peab olema JSON loend ka ühe URL-i puhul.
* :token:`verification_tls` - Nimekiri kogumisteenuse TLS sertifikaatidest PEM vormingus. Järjekord pole oluline. Väärtus peab olema JSON loend ka ühe sertifikaadi puhul.
* :token:`help_url` - Abiinfo vaate URL
* :token:`close_timeout` - Ajaaken, mil on kasutajal võimalik oma valikut näha enne rakenduse sulgumist. Millisekundites.
* :token:`close_interval` - Intervall, millega uuendatakse :token:`close_timeout` väärtust kasutajaliideses. Millisekundites.
* :token:`con_timeout_1` - Kogumisteenusega ühenduse saamise esimese katse ajapiirang. Millisekundites.
* :token:`con_timeout_2` - Kui esimese ringiga ei saadud ühendust ühegi kogumisteenuse instantsiga, proovitakse uuesti selle ajapiiranguga. Millisekundites.
* :token:`public_key` - Valimiste avalik võti, millega krüpteeritakse valijate hääli. PEM vormingus. 
* :token:`tspreg_service_cert` - Ajatembeldusteenuse sertifikaat PEM vormingus. 
* :token:`ocsp_service_cert` - OCSP-teenuse sertifikaadid PEM vormingus. Järjekord pole oluline. Väärtus peab olema JSON loend ka ühe väärtuse puhul.
* :token:`tspreg_client_cert` - Kogumisteenuse registreerimissertifikaat PEM vormingus.

Näide
-----

.. literalinclude:: config-examples/conf_verification_template.json
   :language: json

.. vim: sts=3 sw=3 et:
