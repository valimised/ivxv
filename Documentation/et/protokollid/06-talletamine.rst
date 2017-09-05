..  IVXV protokollid

====================================================
Elektroonilise hääle kvalifitseerimine talletamiseks
====================================================

Kvalifitseeritud hääl
=====================

Valijarakenduse töö tulemusena saadetakse kogumisteenusesse talletamiseks
topeltümbrik, mis sisaldab endas valija tahteavaldust krüpteeritud kujul, valija
allkirja krüpteeritud tahteavaldusel kooskõlastatud allkirja- ja
konteinervormingus ning valija allkirjastamissertifikaati X509 vormingus.

Hääle edukaks talletamiseks näeb IVXV protokoll ette hääle registreerimise
välise registreerimisteenuse osutaja juures ning registreerimistõendi
valijarakendusele kättesaadavaks tegemise. Valimise korraldaja võib hääle
kvalifitseerimiseks näha ette täiendavaid samme lisaks registreerimisele --
näiteks kehtivuskinnituse hankimist hääle allkirjastanud sertifikaadi kohta.

Kõik kogumisteenuse poolt hangitavad kvalifitseerivad elemendid, mis määravad
hääle staatuse hilisemates töötlusetappides tuleb esitada valijarakendusele ning
nõudmise korral ka kontrollrakendusele tagamaks, et valija saab oma hääle
korrektse menetlemise võimalikkusest õigeaegselt teada.

OCSP kehtivuskinnitus
---------------------

OCSP (Online Certificate Status Protocol) on standartne protokoll X509
sertifikaatide kehtivusinfo pärimiseks. Kogumisteenus võib seda protokolli
kasutada hääle allkirjastanud sertifikaadi kehtivuse teada saamiseks. OCSP
vastus ütleb, et sertifikaat kehtis päringu tegemise ajahetkel, kuid ei seosta
OCSP vastust konkreetse allkirjaga.

OCSP-TM kehtivuskinnitus
------------------------

BDOC 2.1 standard kirjeldab BDOC-TM profiili, kus OCSP protokolliga hangitud
kehtivuskinnitus toimib ka ajamärgendina, mis kinnitab, et konkreetne allkiri
eksisteeris enne OCSP kehtivuskinnituse võtmist.

RFC3161 ajatempel
-----------------

RFC3161 ajatempli protokolliga saadakse usaldusteenuse pakkujalt kinnitus, et
mingi andmekogum eksisteeris enne teatud ajahetke. BDOC-TS kontekstis
ajatembeldatakse allkirja element `SignatureValue` kanoniseeritud kujul.
Klassikaline OCSP vastus koos RFC 3161 vormingus ajatempliga kvalifitseerivad
BDOC-TS allkirja.


Hääle registreerimine
====================================================

IVXV registreerimisprotokoll on kirjeldatud dokumendis "Elektroonilise
hääletamise infosüsteemi IVXV registreerimisteenus". Registreerimisteenus toimub
RFC3161 ajatempli protokolli baasil. Protokolli on laiendatud selliselt, et
kogumisteenus saab ajatempli päringule anda oma signatuuri, mis teeb võimalikuks
hilisema võrdleva väljavõtte registreerimisteenusest. Sõltumatu
registreerimisteenuse olemasolu vähendab häälte kogumisteenuse poolt 'kaotamise'
riski.

Puudub olemuslik vajadus registreerimisprotokolli sidumiseks ajatempli
protokolliga.


Talletamine
====================================================

Elektroonilise hääle talletamine kogumisteenuses tähendab

#. hääle vastuvõtmist valijarakenduselt ning hääletaja allkirja
   verifitseerimist;
#. hääle võimalikku kvalifitseerimist -- näiteks sertifikaadi kehtivuse
   tõendamist hääle allkirjastamisele lähedasel ajahetkel;
#. hääle registreerimist sõltumatus registreerimisteenuses;
#. häält kvalifitseerivate elementide vahendamist valijarakendusele.

Erinevad kombinatsioonid allkirjavormingust ning hääli kvalifitseerivatest
teenustest võivad tekitada erinevaid IVXV-profiile. Konkreetse dokumendi raames
on IVXV profiil:

#. Allkirjastatud hääle vorming on BDOC-TS.
#. Kehtivuskinnitusprotokolliks on standartne OCSP.
#. BDOC-TS kvalifitseerimiseks kasutatav RFC3161 ajatempel on kasutusel ka
   registreerimistõendina.

.. vim: sts=3 sw=3 et:
