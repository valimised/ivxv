..  IVXV protokollid

================================================================================
Ülevaade
================================================================================

Elektroonilise hääletamise protokollistik (edaspidi protokollistik) defineerib
elektroonilise hääletamise süsteemi komponentide vahelise sõnumivahetuse,
kasutatavad andmestruktuurid, algoritmid ning liidesed väliste süsteemidega.
Sõnumivahetus esitatakse UML suhtlusskeemidena, mis üheselt defineerivad
sõnumite järgnevuse. Andmestruktuuride kirjeldused on varustatud BNF, ASN.1
või JSON-schema notatsioonis spetsifikatsioonidega. Algoritmid esitatakse
pseudokoodina.

NB! Kõigis protokollistiku andmestruktuuride väljades tuleb rangelt kinni pidada
lubatud märkidest ning väljade minimaalsetest ja maksimaalsetest pikkustest.
Täiendavate tühikute, tabulaatorite jms. kasutamine on keelatud ning
spetsifikatsiooni realiseerivad rakendused peavad vorminguga mitte-vastavate
andmete töötlemisest keelduma.

Protokollistik defineerib elektroonilise hääletamise protokolli ning selle
protokolli realiseerimiseks vajalikud tugistruktuurid.

Elektroonilise hääletamise protokoll
====================================

Elektroonilise hääletamise protokoll spetsifitseerib:

#. elektroonilise hääle vormingu, mis võimaldab üheselt määratleda valija tahte
   konkreetsel valimisel;

#. elektroonilise hääle krüpteerimise hääle salajasuse tagamiseks;

#. elektroonilise hääle digitaalse allkirjastamise tervikluse ja valija
   identifitseerimise tagamiseks;

#. elektroonilise hääle kvalifitseerimise kogumisteenuse poolt, hääle
   vastuvõtmise tähistamiseks;

Protokoll eeldab, et valimise korraldaja defineerib valimise ning genereerib
häälte salastamise võtmepaari, mille avalik komponent tehakse valijarakendusele
kättesaadavaks.

Protokolli vahendusel liigub valija tahe kogumisteenuses talletatavasse e-valimiskasti
ning võetakse tulemuse kujunemisel arvesse järgmist sündmusterida pidi:

#. Valija kasutab valijarakendust oma tahteavalduse elektrooniliseks
   vormistamiseks:

   #. tahteavaldus vormistatakse elektroonilise häälena;

   #. vormistatud hääl krüpteeritakse;

   #. krüpteeritud hääl signeeritakse valija arvutis.

#. Kogumisteenus talletab elektroonilise hääle, moodustades selle käigus
   häälele kvalifitseeritud digitaalallkirja:

   #. elektrooniline hääl registreeritakse välises registreerimisteenuses;

   #. elektroonilisele häälele võetakse digitaalne ajatempel;

   #. elektroonilisele häälele võetakse valija sertifikaadi
      kehtivuskinnitus;

   #. elektroonilist häält kvalifitseerivad elemendid tagastatakse mh. ka
      valijarakendusele kontrollimiseks ning valija informeerimiseks
      kvalifitseerimise tulemustest;

   #. valijale võimaldatakse kvalifitseeritud elektroonilise hääle
      kontrollimine kontrollrakenduse abil.

.. note::

   Elektroonilise hääle digitaalne allkirjastamine erineb tavapärasest
   dokumentide digitaalallkirjastamisest, kus kõik allkirja kvalifitseerimiseks
   vajalikud toimingud algatatakse vahetult allkirjastaja seadmes.
   Elektroonilise hääle kvalifitseerimise kohustus on kogumisteenusel, kelle
   ülesanne on veenduda vastuvõetavate häälte korrektses allkirjastatuses. Kuna
   e-hääletamise perioodil on koormus seotud teenustele kõrge, võimaldab
   kogumisteenuse poolt juhitud kvalifitseerimine tagada paremat teenuse
   kvaliteeti.

#. Valija võib kasutada kontrollrakendust veendumaks oma hääle korrektses
   käitlemises kogumisteenuse poolt;

#. Hääletamisperioodi lõppedes väljastab kogumisteenus valimise korraldajale
   e-valimiskasti ning registreerimisteenus väljavõtte kogumisteenuse poolt
   registreeritud häältest;

   #. e-valimiskasti koosseisus antakse valimise korraldajale üle:

      #. valija krüpteeritud tahteavaldus koos signatuuriga;

      #. registreerimisteenuse kinnitus hääle registreerimisest;

      #. ajatempliteenuse poolt väljastatud digitaalne ajatempel
         elektroonilisele häälele;

      #. kehtivuskinnitusteenuse poolt väljastatud kinnitus valija sertifikaadi
         kehtivuse kohta;

      #. registreerimisteenuse väljavõtte koosseisus antakse valimise korraldajale üle:

         #. kõik e-hääletamise perioodil kogumisteenuse poolt
            registreerimisteenusele saadetud päringud elektrooniliste häälte
            registreerimiseks.

#. Valimise korraldaja arvutab hääletamistulemuse:

   #. kontrollitaks üle antud elektrooniliste häälte allkirjade kehtivust

   #. kontrollitakse, et kõik registreerimisteenuses registreeritud hääled on
      e-valimiskasti koosseisus üle antud;

   #. eraldatakse krüpteeritud hääled ja digitaalallkirjad;

   #. anonüümitakse krüpteeritud hääled krüptograafiliselt;

   #. dekrüpteeritakse krüpteeritud hääled;

   #. dekrüpteeritud häälte põhjal arvutatakse hääletamistulemus.

Protokoll on analoogne paberil posti teel hääletamise protokolliga, kus valija
tahe liigub valimiskomisjonini kahes ümbrikus – välimise ümbriku sees on
sisemine ümbrik, mis omakorda sisaldab valija tahteavaldusega hääletussedelit.
Välimine ümbrik kannab valijat identifitseerivat infot ning võimaldab
mh. kontrollida valija õigust hääletada. Sisemine ümbrik on anonüümne ning
kaitseb hääle salajasust. Enne häälte kokkulugemist eraldatakse sisemised
ümbrikud välimistest.

Elektroonilise hääletamise kontekstis on sisemine ümbrik vormistatud
krüpteeritud häälena ning välimine ümbrik digitaalselt allkirjastatud
dokumendina.
