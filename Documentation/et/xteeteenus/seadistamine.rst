..  IVXV tehniline dokumentatsioon

Seadistamine
============

Teenuse konfigureerimiseks kasutatakse ``xroad-service.json`` faili.

``server.address`` - Serveri port

``server.batchmaxsize`` - Paki suurus, soovitatud suurus 1000

``server.openapipath`` - OpenAPI faili asukoht. Serveeritakse https://host/openapi.

``server.tls`` - Serveri TLS konfiguratsioon

``xroad.certificate`` -  X-tee turvaserveri sertifikaat

``elections``- Valimissündmuste list

``elections.name``- Valimissündmuse nimi

``elections.address`` - IVXV serveri aadress

``elections.servername`` - Järjekorrateenuse SNI

``elections.rootca`` - IVXV CA sertifikaat

``elections.clientcert`` - Kliendi sertifikaat, kliendi CA tuleb lisada IVXV konfiguratsiooni

``elections.clientkey`` - Kliendi võti


