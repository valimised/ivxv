..  IVXV kasutusmallid

Tegijad
=======

Korraldaja
----------

*Korraldaja* on isik, kes korraldab valimisi mille raamides elektroonilise hääletamise süsteemi kasutatakse. Korraldaja määrab elektroonilise hääletamise seadistuse, sealhulgas kõik ülejäänud rollitäitjad. Korraldaja haldab Kogumisteenust sisuliselt.

Korraldaja tegeleb hääletamiseelsel etapil valijate, valimisjaoskondade, ringkondade ja valikute- kandidaatide (valimistel) või vastusevariantide (rahvahääletusel) nimekirja koostamise ja/või kinnitamisega.

Korraldaja loob ja/või kinnitab Valimisrakenduse ning Kontrollrakenduse seadistused.

Korraladaja tegeleb hääletamisetapil valijate nimekirjade täiendamise ja/või täienduste kinnitamisega.

Korraldaja tegeleb töötlusetapil e-häälte tühistamis- ja ennistamisavalduste koostamise ja/või kinnitamisega.

Korraldaja võib täita ka elektroonilise hääletamise põhiprotsesside toimimisega seotud konkreetseid rolle. Näiteks hoiab Korraldaja tavajuhtumitel ka e-hääletamise süsteemi põhisaladust – häälte avamise võtit ning seega täidab ta ka häälte avaja ja summeerija ehk Lugeja rolli.

Hääletaja
---------

*Hääletaja* on isik, kes kasutab elektroonilise hääletamise süsteemi hääletamisetapil

- hääletamiseks Valijarakendusega,
- hääle kontrollimiseks Kontrollrakendusega.

Hääletajal on elektrooniline isikutunnistus – ID-kaart või Mobiil-ID.

Koguja
------

*Koguja* on isik, kes haldab Kogumisteenust, Valijarakendust ja Kontrollrakendust tehniliselt.

*Kogumisteenus* on serverisüsteem, mis tuvastab hääleõiguslikkust Tuvastusteenuse abil, väljstab kandidaatide nimekirja, aitab Valijarakendusel Allkirjastamisteenuse abil e-häält moodustada, talletab hääli e-urni ning registreerib neid Registreerimisteenuses. Kogumisteenus vastab Kontrollrakenduse poolt tehtud hääle tervikluse kontrollpäringutele.

*Kogumisteenuse Haldur* on kas Korraldaja, Koguja või Klienditugi. Koguja on Kogumisteenuse tehniline haldur. Kogumisteenuse sisuline haldur on Korraldaja. Kogumisteenus pakub informatiivset liidest Klienditoele.

*Haldusteenus* on teenus, mida Kogumisteenuse Haldur kasutab Kogumisteenuselt informatsiooni saamiseks või Kogumisteenusele korralduste edastamiseks.

*Valijarakendus* on rakendus, mida Hääletaja kasutab e-hääle andmiseks käimasolevatel valimistel. Valijarakendus suhtleb Kogumisteenusega ning võimaldab Hääletajal teha valikut, seda krüpteerida ja digitaalselt allkirjastada. Valijarakendus kuvab QR-koodi, mille alusel saab Hääletaja Kontrollrakendusega kontrollida e-hääle korrektset jõudmist Kogumisteenusesse.

Koguja valmistab kompileeritud valijarakendusest ette seadistatud valijarakenduse, mis allkirjastatakse ja levitatakse valijateni.

*Kontrollrakendus* on rakendus, mis võimaldab Hääletajal arvutist erineval nutiseadme platvormil veenduda, et tema e-hääl jõudis Kogumisteenusesse ja Registreermisteenusesse ning väljendas tema tahet korrektselt.

Koguja seadistab kontrollrakenduse selliselt, et too on võimeline võrgust tegelikke seadistusi laadima ning allkirjastab kontrollrakenduse.

Koguja on Valijarakenduse ja Kontrollrakenduse tehniline haldur, sisuline haldur on Korraldaja.

Koguja poolt täidetavad põhiprotsessid leiavad aset hääletamisetapil, alustavad ning lõpetavad tegevused toimuvad nii hääletamiseelsel kui töötlusetapil.

Koguja digiallkirjastab hääletamisperioodi lõpul Töötlejale üle antavad andmed (e-hääled ja logid).

Töötleja
--------

*Töötleja* on isik, kes kasutades Töötlemisrakendust töötleb töötlusetapil hääletamisperioodil kogutud e-hääli:

- kontrollib digitaalallkirju ja Kogujalt saadud andmete täielikkust,
- tühistab korduvad e-hääled ning paralleelhääletamise kasutamisel ka nende Hääletajate e-hääled, kes hääletasid jaoskonnas eelhääletamise ajal,
- anonüümistab e-hääled, eemaldades nendelt isikulised digitaalallkirjad, olles eelnevalt need sorteerinud ringkondade kaupa

Töötleja võib e-hääli täiendavalt krüptograafiliselt anonüümistada kasutades Miksimisrakendust.

*Töötlemisrakendus* on rakendus, mille abil kontrollitakse häälte individuaalset terviklust ja e-urni terviklust, tühistatakse hääli, väljastatakse hääletanute nimekirju ning ringkondade kaupa rühmitatud anonüümistatud hääli. Töötlemisrakenduse sisendi annavad Koguja, Registreerimisteenus ja Korraldaja. Töötlemisrakendust võib käitada ka Audiitor Töötleja töötulemuste kontrollimiseks.

Miksija
-------

*Miksija* on isik, kes töötlusetapil anonüümistab e-hääled krüptograafiliselt kasutades Miksimisrakendust.

*Miksimisrakendus* on rakendus, mille sisendiks on ringkondade kaupa rühmitatud anonüümistatud krüpteeritud hääled ning mis väljastab krüptograafiliselt segatud hääled selliselt, et neid ei ole võimalik sisendiga vastavusse viia. Miksimine toimub sellisel moel, et nii sisend- kui ka väljundhäälte dekrüpteerimine ja tulemi summeerimine annab sama resultaadi. Miksimirakendus väljastab lisaks segatud häältele miksimistõendi, mis kinnitab sisend- ja väljundhäälte semantilist samaväärsust.

Lugeja
------

*Lugeja* on isik, kes kasutades Võtmerakendust

- hääletamiseelsel etapil genereerib häälte salastamise ja avamise võtme,
- lugemisetapil avab krüpteeritud hääled ning summeerib need e-hääletamise tulemusteks.

Lugeja võib toimetada ainuisikuliselt, üldjuhul kaitstakse e-hääletamise salastamise võti läviskeemiga, kus ühe tervikvõtme asemel luuakse mitu võtmeosakut ning võtmeoperatsioonid on teostatavad ainult teatud kvoorumi võtmeosakute osavõtul. Sellisel juhul assisteerivad Lugejat Võtmehaldurid.

*Võtmehaldur* on isik, kelle ülesandeks on talle usaldatud võtmeosaku alalhoidmine ning selle andmine Lugeja käsutusse ainult siis kui selleks on seaduslik alus.

*Võtmerakendus* on rakendus, millega genereeritakse iga hääletamise jaoks häälte salastamise ja häälte avamise võti. Võtmerakenduse abil toimub ka häälte lugemine ja tulemuse väljastamine.

Audiitor
--------

*Audiitor* on isik, kes auditeerimisetapil kontrollib süsteemi kirjelduse ja Korraldaja poolt avaldatud süsteemi kesksete osapoolte vahel liikunud andmete põhjal andmete terviklust ning kooskõlalisust. Audiitor kasutab oma töös Auditirakendust. Kui Audiitor kontrollib ka Töötleja toimimise korrektsust, siis kasutab Audiitor ka Töötlemisrakendust.

*Auditirakendus* on rakendus mis võimaldab kontrollida Lugeja ja Miksija töö korrektsust. Lugeja töö korrektsust on võimalik kontrollida ka avalikult.

Klienditugi
-----------

*Klienditugi* on isik, kelle poole Hääletaja hääletamisetapil probleemide korral pöördub. Klienditugi abistab Kogumisteenusest saadud info abil Hääletajat probleemide lahendamisel.

Tuvastusteenus
--------------

*Tuvastusteenus* on teenus, mida kasutatakse vajadusel hääletaja identiteedi tuvastamiseks.

Allkirjastamisteenus
--------------------

*Allkirjastamisteenus* on teenus, mida kasutatakse vajadusel hääle allkirjastamisel ja sellele kehtivuskinnituse saamisel. Vajadus Allkirjastamisteenuse järele sõltub allkirja andmise vahendist - Mobiil-ID ja ID-kaardi korral on nende teenuste ülesehitus erinev.

Registreerimisteenus
--------------------

*Registreerimisteenus* on teenus, mille abil Kogumisteenus peab registreerima kõik Valijarakendustelt saadud hääled. Pärast hääletamisperioodi lõppu edastab teenuseosutaja info registreeritud häälte kohta Töötlejale.
