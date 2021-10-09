..  IVXV protokollid

================================================================================
Ülevaade
================================================================================

Elektroonilise hääletamise protokollistik (edaspidi protokollistik) defineerib
elektroonilise hääletamise süsteemi komponentide vahelise sõnumivahetuse,
kasutatavad andmestruktuurid, algoritmid ning liidesed väliste süsteemidega.
Sõnumivahetus esitatakse UML suhtlusskeemidena, mis üheselt defineerivad
sõnumite järgnevuse. Andmestruktuuride kirjeldused on varustatud Backus-Naur
või JSON-schema notatsioonis spetsifikatsioonidega.  Andmestruktuuride väljade
eraldajateks kasutatakse reavahetusmärki ``LF`` (ASCII-kood ``0x0A``) ja
tabeldusmärki ``TAB`` (ASCII-kood ``0x09``). Algoritmid esitatakse
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

Protokoll eeldab, et valimise korraldaja on defineerinud valimise ning
genereerinud häälte salastamise võtmepaari, mille avalik komponent on tehtud
valijarakendusele kättesaadavaks.

Protokolli vahendusel liigub valija tahe kogumisteenuses talletatavasse e-valimiskasti
ning võetakse tulemuse kujunemisel arvesse järgmist sündmusterida pidi:

#. Valija kasutab valijarakendust oma tahteavalduse elektrooniliseks
   vormistamiseks:

   #. tahteavaldus vormistatakse elektroonilise häälena;

   #. vormistatud hääl krüpteeritakse;

   #. krüpteeritud hääl allkirjastatakse digitaalselt.

#. Kogumisteenus talletab elektroonilise hääle:

   #. digitaalselt allkirjastatud häälele võetakse valija sertifikaadi
      kehtivust kinnitavad elemendid;

   #. elektrooniline hääl registreeritakse välises registreerimisteenuses;

   #. valijale võimaldatakse kvalifitseeritud elektroonilise hääle
      kontrollimine kontrollrakenduse abil.

#. Valija võib kasutada kontrollrakendust veendumaks oma hääle korrektses
   käitlemises kogumisteenuse poolt;

#. Hääletamisperioodi lõppedes väljastab kogumisteenus valimise korraldajale
   e-valimiskasti ning registreerimisteenus loendi kogumisteenuse poolt registreeritud
   häältest;

#. Valimise korraldaja arvutab hääletamistulemuse:

   #. veendutakse, et kõik registreerimisteenuses registreeritud hääled on
      e-valimiskasti koosseisus üle antud;

   #. eraldatakse krüpteeritud hääled ja digitaalallkirjad;

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
