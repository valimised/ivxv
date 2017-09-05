..  IVXV protokollid

================================================================================
Ülevaade
================================================================================

Elektroonilise hääletamise protokollistik (edaspidi protokollistik) defineerib
elektroonilise hääletamise süsteemi komponentide vahelise sõnumivahetuse,
kasutatavad andmestruktuurid, algoritmid ning liidesed väliste süsteemidega.
Sõnumivahetus esitatakse UML suhtlusskeemidena, mis üheselt defineerivad
sõnumite järgnevuse. Andmestruktuuride kirjeldused on varustatud Backus-Naur või
JSON-schema notatsioonis spetsifikatsioonidega.  Andmestruktuuride väljade
eraldajateks kasutatakse reavahetuse sümbolit LF, ASCII-koodiga 0x0A ja
tabulaatori sümbolit TAB, ASCII-koodiga 0x09. Algoritmid esitatakse
pseudokoodina.

NB! Kõigis protokollistiku andmestruktuuride väljades tuleb rangelt kinni pidada
lubatud sümbolitest ning väljade minimaalsetest-maksimaalsetest pikkustest.
Täiendavate tühikute, tabulaatorite jms. kasutamine on keelatud ning
spetsifikatsiooni realiseerivad rakendused peavad formaadile mitte-vastavate
andmete töötlemisest keelduma.

Protokollistik defineerib elektroonilise hääletamise protokolli ning selle
protokolli realiseerimiseks vajalikud tugistrutkuurid.

Elektroonilise hääletamise protokoll
====================================

Elektroonilise hääletamise protokoll spetsifitseerib

#. elektroonilise hääle formaadi, mis võimaldab üheselt määratleda valija
   tahte konkreetsel valimisel;
#. elektroonilise hääle krüpteerimise hääle salajasuse tagamiseks;
#. elektroonilise hääle digitaalse allkirjastamise tervikluse ja valija
   identifitseerimise tagamiseks;
#. elektroonilise hääle kvalifitseerimise kogumisteenuse poolt, hääle
   vastu võtmise tähistamiseks;

Protokoll eeldab, et valimise korraldaja on defineerinud valimise ning
genereerinud häälte salastamise võtmepaari, mille avalik komponent on tehtud
valijarakendusele kättesaadavaks.

Protokolli vahendusel liigub valija tahe kogumisteenuses talletatavasse e-urni
ning võetakse tulemuse kujunemisel arvesse järgmist sündmusterida pidi:

#. Valija kasutab valijarakendust oma tahteavalduse vormistamiseks
   elektrooniliselt

   #. tahteavaldus vormistatakse elektroonilise häälena,

   #. vormistatud hääl krüpteeritakse,

   #. krüpteeritud hääl allkirjastatakse digitaalselt.

#. Kogumisteenus talletab elektroonilise hääle

   #. digitaalselt allkirjastatud häälele võetakse valija sertifikaadi
      kehtivust kinnitavad elemendid,

   #. elektrooniline hääl registreeritakse välises registreerimisteenuses,

   #. valijale võimaldatakse kvalifitseeritud elektroonilise hääle
      kontrollimine kontrollrakendusega.

#. Valija võib kasutada kontrollrakendust veendumaks oma hääle korrektses
   käitlemises kogumisteenuse poolt.

#. Hääletamisperioodi lõppedes väljastab kogumisteenus valimise korraldajale
   e-urni ning registreerimisteenus loendi kogumisteenuse poolt registreeritud
   häältest.

#. Valimise korraldaja arvutab hääletamistulemuse

   #. veendutakse, et kõik registreerimisteenuses registreeritud hääled on
      e-urni koosseisus üle antud

   #. eraldatakse krüpteeritud hääled ja digitaalallkirjad,

   #. dekrüpteeritakse krüpteeritud hääled,

   #. dekrüpteeritud häälte põhjal arvutatakse hääletamistulemus.

Protokoll on analoogne paberil posti teel hääletamise protokolliga, kus valija
tahe liigub valimiskomisjonini kahes ümbrikus – välimise ümbriku sees on
sisemine ümbrik, mis omakorda sisaldab valija tahteavaldusega hääletussedelit.
Välimine ümbrik kannab valijat identifitseerivat informatsiooni ning võimaldab
mh. kontrollida valija õigust hääletada. Sisemine ümbrik on anonüümne ning
kaitseb hääle salajasust. Enne häälte kokkulugemist eraldatakse sisemised
ümbrikud välimistest.

Elektroonilise hääletamise kontekstis on sisemine ümbrik vormistatud
krüpteeritud häälena ning välimine ümbrik digitaalselt allkirjastatud
dokumendina.

.. vim: sts=3 sw=3 et:
