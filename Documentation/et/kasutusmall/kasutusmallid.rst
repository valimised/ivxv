..  IVXV kasutusmallid

Kasutusmallid
=============

Hääletamiseelne etapp
---------------------

Korraldaja - Valimise defineerimine
```````````````````````````````````

Kirjeldus
'''''''''

Rangelt võttes ei ole tegemist elektroonilise hääletamise infosüsteemi kasutusmalliga, kuid see on elektroonilise hääletamise alguspunkt - Korraldaja kirjeldab valimise, küsimused, ringkonnad, valijad ning erinevate rollide täitjad e-hääletamise põhiprotsessides - Lugeja, Koguja, Töötleja jt.

Lugeja - Häälte salastamise ja avamise võtme genereerimine
``````````````````````````````````````````````````````````

Kirjeldus
'''''''''

Lugeja genereerib võtmepaari e-häälte salastamiseks ja avamiseks. Avamisvõtme osakud jagatakse Võtmehaldurite vahel.

Eeltingimus
'''''''''''

#. Lugejal on olemas Võtmerakendus.
#. Läviskeemi tehniline seadistus on olemas.
#. Leidub Läviskeemi rakendamiseks vajalik hulk Võtmehaldureid ning võtmekandjaid.

Päästik
'''''''

Lugeja algatab võtmepaari genereerimise Võtmerakenduses.

Põhiprotsess
''''''''''''

#. Võtmerakendus verifitseerib konfiguratsiooni digitaalallkirja.
#. Võtmerakendus genereerib võtmeosakud vastavalt läviskeemi spetsifikatsioonile.
#. Võtmerakendus küsib Lugejalt võtmekandjaid osakute talletamiseks.
#. Võtmerakendus talletab osakud võtmekandjatele.
#. Võtmerakendus kasutab häälte salastamise võtit sõnumi krüpteerimiseks.
#. Võtmerakendus testib võtmekandjaid ning taastab 2 erineva kvoorumiga võtme sõnumi dekrüpteerimiseks.
#. Võtmerakendus allkirjastab krüpteerimisvõtme allkirjastamisvõtmega.

Laiendid
''''''''

- Tehniliste vigade tekkimisel kõigis põhivoo sammudes logitakse sündmus tehniliste vigade logisse. Kasutajale väljastatakse  teade veasituatsiooni kohta..

Järeltingimus
'''''''''''''

Häälte salastamise ja avamise võtmepaar on edukalt genereeritud, avamise võti on talletatud osakutena võtmekandjatel.


Korraldaja - Valijarakenduse seadistuse loomine
```````````````````````````````````````````````

Kirjeldus
'''''''''

Korraldaja konfigureerib Valijarakenduse poolt kasutatavad tekstid, sertifikaadid, ikoonid ja valimise parameetrid.

Eeltingimus
'''''''''''
Seadistamiseks on täidetud järgmised tingimused:

- Korraldaja on Valijarakenduse tekstid ette valmistanud
- Korraldaja on valimise parameetrid ette valmistanud
- Häälte salastamise võti on olemas.

Põhiprotsess
''''''''''''
#. Korraldaja käivitab Seadistamisrakenduse.
#. Korraldaja sisestab valimise ja küsimuste identifikaatorid ning tüübid.
#. Korraldaja sisestab valimise küsimused.
#. Korraldaja laadib häälte salastamise võtme.
#. Korraldaja konfigureerib Valijarakenduse kasutajaliidese – tekstid, fondid, värvid, ikoonid.
#. Korraldaja kontrollib kõigi Valijarakenduse vaadete vastavust soovitule.
#. Korraldaja salvestab loodud konfiguratsiooni.

Laiendid
''''''''

Järeltingimus
'''''''''''''
Seadistamise lõpuks on olemas Valijarakenduse konfiguratsioon, millest on puudu Kogumisteenuse tehniline info.

Korraldaja - Kontrollrakenduse seadistuse loomine
`````````````````````````````````````````````````

Eeltingimus
'''''''''''
Korraldajal on olemas

#. Hääletaja allkirja kontrolliks vajalikud sertifikaadid
#. Registreerimisteenuse kinnituse kontrolliks vajalikud sertifikaadid
#. Kogumisteenuse allkirja kontrolliks vajalikud sertifikaadid
#. Käimasoleval valimisel häälte krüpteerimiseks kasutatav avalik võti
#. Kontrollrakenduse tekstid HTML kujul
#. Kontrollrakenduse fondid
#. Kontrollrakenduse värvid
#. Käimasoleva valimise identifikaator ja küsimused

Põhiprotsess
''''''''''''
1. Korraldaja loob Kontrollrakenduse seadistuse
2. Korraldaja veendub seadistamise õnnestumises vaadates läbi rakenduse vaated

Laiendid
''''''''
1. Vaadete mittesobivuse korral muudab Korraldaja seadistusi ning alustab põhiprotsessi algusest.


Koguja - Kogumisteenuse seadistamine
````````````````````````````````````

Kirjeldus
'''''''''
Koguja valmistab Kogumisteenuse valimiseks ette.

Eeltingimus
'''''''''''
#. Kogumisteenuse operatsioonisüsteem on paigaldatud
#. Kogumisteenuse tarkvarapakid on paigaldatud.

Päästik
'''''''
Kasutusmall algab kui Koguja logib enne hääletamisperioodi algust Kogumisteenusesse.

Põhiprotsess
''''''''''''
1. Koguja viib läbi alljärgnevad sammud:

    a. Laeb konfiguratsioonipakkide verifitseerimiseks vajaliku sertifikaatide konfiguratsiooni.
    b. Laeb Tuvastusteenuse konfiguratsiooni.
    c. Laeb Allkirjastamisteenuse konfiguratsiooni.
    d. Laeb Registreerimisteenuse konfiguratsiooni.
    e. Laeb mikroteenuste võrgukonfiguratsiooni.
    f. Laeb talletustehnoloogia konfiguratsiooni.
    g. Laeb volitatud nimekirjade laadijate konfiguratsiooni.
    h. Laeb digitaalselt allkirjastatud jaoskondade/ringkondade nimekirja.
    i. Laeb digitaalselt allkirjastatud valikute nimekirja.
    j. Laeb signeeritud valijate nimekirja.
    k. Kogumisteenus kontrollib digitaalallkirju konfiguratsioonipakkidel ja nimekirjadel (tehniline kasutusmall Digitaalallkirja kehtivuse kontroll Kogumisteenuses).
    l. Kogumisteenus kontrollib tuvastatud allkirjastaja volitusi süsteemis.
    m. Kogumisteenus kontrollib konfiguratsiooni vormilist ja sisulist kooskõlalisust ning initsialiseerib sisemised andmestruktuurid.

Laiendid
''''''''
- Kui põhiprotsessi sammus 1.k  ei tuvastata digitaalallkirja  kehtivust, siis väljastatakse veateade, sündmus logitakse vigade logisse, ühtegi muudatust ei rakendata.
- Kui põhiprotsessi sammus 1.l  ei tuvastata  digitaalallkirjastaja vastavat volitust volitatud isikute nimekirjast, siis väljastatakse veateade, sündmus logitakse vigade logisse, ühtegi muudatust ei rakendata.
- Kui põhivoo sammus 1.m tuvastatakse vorminguprobleemid, siis väljastatakse veateade, sündmus logitakse vigade logisse, ühtegi muudatust ei rakendata.
- Kui põhivoo sammus 1.m tuvastatakse ebakooskõla konfiguratsioonis – valed valimised, korduv jaoskond, korduv ringkond, jaoskond olematus ringkonnas, korduv kandidaat, olematu ringkonnaga kandidaat, korduv valija, olematu jaoskonnaga Hääletaja vms. – siis väljastatakse veateade, sündmus logitakse vigade logisse, ühtegi muudatust ei rakendata.
- Tehniliste vigade tekkimisel kõigis põhivoo sammudes logitakse sündmus tehniliste vigade logisse. Kasutajale väljastatakse  teade veasituatsiooni kohta..

Järeltingimus
'''''''''''''
Kogumisteenus on kooskõlaliselt seadistatud ning seda on võimalik viia hääletamisetappi.


Koguja - Valijarakenduse valmendamine
`````````````````````````````````````

Kirjeldus
'''''''''
Koguja valmendab Valijarakenduse.

Eeltingimus
'''''''''''
Valmendamiseks on täidetud järgmised tingimused:

- Korraldaja on loonud omapoolse Valijarakenduse konfiguratsioonifaili
- Valijarakendus on kompileeritud kõigi toetatavate platvormide jaoks.
- Valijarakendus on seadistamata.
- Kogumisteenuse TLS-sertifikaadi usaldamiseks vajalikud sertifikaadid on olemas.
- Kogumisteenuse URI on teada.

Põhiprotsess
''''''''''''
#. Koguja käivitab Seadistamisrakenduse.
#. Koguja laadib Korraldaja loodud Valijarakenduse konfiguratsioonifaili.
#. Koguja laadib Kogumisteenuse TLS-sertifikaadi usaldamiseks vajalikud sertifikaadid.
#. Koguja rakendab loodud seadistuse Valijarakendusele.
#. Koguja kontrollib kõigi Valijarakenduse vaadete vastavust soovitule.

Laiendid
''''''''
Mõne vajaliku ressurssi puudumisel seadistamine katkestatakse ning protsessi alustatakse uuesti kui kõik eeltingimused on täidetud.

Järeltingimus
'''''''''''''
Seadistamise lõpuks Valijarakendus käivitub ning kasutab seadistatud ressursse.


Koguja - Kontrollrakenduse valmendamine
```````````````````````````````````````

Eeltingimus
'''''''''''
Kogujal on olemas

1. Kompileeritud Kontrollrakendus
2. Kontrollrakenduse allkirjastamiseks sobiv võtmepaar
3. Kogumisteenuse võrguaadress
4. Kogumisteenuse sertifikaadi verifitseerimiseks vajalikud sertifikaadid

Põhiprotsess
''''''''''''
1. Koguja seadistab Kontrollrakenduse Kogumisteenuse võrguaadressi ja vastavate sertifikaatidega.
2. Koguja pakendab rakenduse
3. Koguja allkirjastab pakendatud rakenduse

Korraldaja - Proovihääletamine
``````````````````````````````

Kirjeldus
'''''''''

Korraldaja koostöös Koguja, Töötleja, Lugeja ning Võtmehalduritega veendub Kogumisteenuse valmisolekus elektrooniliseks hääletamiseks ning kõigi komponentide kooskõlalises seadistatuses.

Eeltingimus
'''''''''''

#. e-häälte salastamise võti on loodud ja testitud
#. Kogumisteenus on seadistatud
#. Valijarakendus on seadistatud
#. Kontrollrakendus on seadistatud

Põhiprotsess
''''''''''''
1. Korraldaja viib läbi testhääletamise

   #. Koguja viib Kogumisteenuse hääletamisetappi
   #. Korraldaja annab Valijarakenduse abil ühe või mitu häält
   #. Korraldaja kontrollib Kontrollrakenduse abil hääli
   #. Koguja peatab Kogumisteenuse ja väljastab e-urni
   #. Töötleja genereerib e-hääletanute nimekirja
   #. Töötleja loob lugemisele minevate anonüümistatud e-häälte nimekirja
   #. Lugeja koos Võtmehalduritega aktiveerib Võtmerakenduse ja häälte avamise võtme
   #. Lugeja dekrüpteerib anonüümistatud hääled ja väljastab hääletamistulemuse

2. Koguja lõpetab testhääletamise ning viib Kogumisteenuse algsesse olekusse, kus talletatud häälte andmebaas on tühi.

Laiendid
''''''''

- Testhääletamine võib sisaldada ka miksimise ja auditeerimise töövooge.

Järeltingimus
'''''''''''''
Elektroonilise hääletamise süsteemi komponendid on kontrollitult kooskõlaliselt seadistatud.

Hääletamisetapp
---------------

Korraldaja - Hääletamise alustamine
```````````````````````````````````

Kirjeldus
'''''''''
Hääletamise alustamine viib Kogumisteenuse hääletamisetappi - algab valikute nimekirjade väljastamine, häälte talletamine ja kontrollpäringutele vastamine.

Päästik
'''''''

#. Valimise seadistustes näidatud hääletamise algusaeg jõuab kätte.
#. Korraldaja edastab Haldusteenuse vahendusel digitaalallkirjastatud korralduse hääletamise alustamiseks.

Põhiprotsess
''''''''''''

1. Kogumisteenus alustab valikute nimekirjade väljastamist, häälte talletamist ja kontrollpäringutele vastamist.

Järeltingimus
'''''''''''''

Kogumisteenus väljastab valikute nimekirju, talletab hääli ning vastab kontrollpäringutele.

Hääletaja - Elektrooniline hääletamine Valijarakendusega
````````````````````````````````````````````````````````

Kirjeldus
'''''''''
Hääletaja kasutab Valijarakendust elektroonilise hääle andmiseks käimasoleval valimisel, mille valijate nimekirja ta kuulub.

Eeltingimus
'''''''''''
Hääletaja on laadinud oma arvutisse käimasoleva valimise jaoks seadistatud Valijarakenduse.

Päästik
'''''''
Kasutusmall käivitub kui Hääletaja on käivitanud Valijarakenduse.

Põhiprotsess
''''''''''''
#. Hääletaja autendib ennast elektroonilise isikutunnistusega (tehniline kasutusmall Autentimine Valijarakenduses) Kogumisteenusele.
#. Kogumisteenus saadab isikukoodi alusel Valijarakendusele (tehniline kasutusmall Valikute nimekirjade väljastamine Valijarakendusele):

    - Hääletaja elukohajärgse valimisringkonna küsimuste ja valikute nimekirja valimisel,
    - teavituse varasema hääletamise kohta, kui sellel valimisel on sama isikukoodi kohta juba talletatud elektrooniline hääl.

3. Valijarakendus esitab Hääletaja isikuandmed, käimasoleva valimise kirjelduse ning küsimused.
#. Valijarakendus esitab Hääletaja elukohajärgse valimisringkonna valikute nimekirja.
#. Hääletaja teeb kõigi küsimuste kontekstis kuvatud valikute hulgast valiku.
#. Valijarakendus esitab Hääletajale tehtud valikute andmed (valiku nimi, valiku number, teatud valimistel ka valimisnimekirja nimi või üksikkandidaat) ning küsib kinnitust valikutele.
#. Hääletaja kinnitab tehtud valikud.
#. Valijarakendus krüpteerib Hääletaja valikud valimise avaliku võtmega ning algatab hääle digitaalse allkirjastamise (tehniline kasutusmall Digitaalne allkirjastamine Valijarakenduses) Hääletaja elektroonilise isikutunnistusega. Valijarakendus saadab digitaalselt allkirjastatud hääle Kogumisteenusele talletamiseks (tehniline kasutusmall Hääle talletamine Kogumisteenuses).
#. Valijarakendus verifitseerib Kogumisteenuse poolt vastusena saadetud Registreerimiskinnitust, kuvab Hääletajale teate hääle edukast talletamisest ning hääle kontrollimiseks vajaliku kontrollkoodi.

Laiendid
''''''''
- Hääletaja võib põhiprotsessi läbiviimisel kasutada ligipääsetavustehnoloogiaid (nt. ekraanilugerid).
- Põhiprotsessi mis tahes etapis asetleidvad vead on fataalsed ning toovad kaasa hääletamisprotsessi katkestamise. Vea põhjuse kõrvaldamise korral tuleb kasutusmall uuesti käivitada.

Olulisimad vead:

- Elektroonilise isikutunnistuse kasutamine ebaõnnestus
- Sidehäire Valijarakenduse ja Kogumisteenuse vahel
- Autentimine ebaõnnestus
- Hääletajal ei ole hääleõigust käimasoleval valimisel
- Digitaalne allkirjastamine ebaõnnestus
- Tehnilised vead

Järeltingimus
'''''''''''''
Õnnestunud elektroonilise hääletamise korral kuvatakse Hääletajale kontrollkood, mille alusel on võimalik kontrollida elektroonilise hääle jõudmist Kogumisteenusesse ning vastavust Hääletaja tahtele.
Kasutusmalli katkestamine enne elektroonilise hääle digitaalset allkirjastamist ei mõjuta Hääletaja Kogumisteenuses eelnevalt talletatud häält.

Hääletaja - Elektroonilise hääle kontrollimine Kontrollrakendusega
``````````````````````````````````````````````````````````````````

Kirjeldus
'''''''''
Hääletaja kasutab vahetult peale hääletamist Valijarakendusega Kontrollrakendust elektroonilise hääle korrektse jõudmise Kogumisteenusesse kontrollimiseks.

Eeltingimus
'''''''''''
#. Hääletaja valduses on Kontrollrakendusega mobiilseade.
#. Hääletaja on kasutanud Valijarakendust hääletamiseks.
#. Hääle kontrollimise ajaaken ei ole veel möödunud.
#. Valijarakenduse viimane vaade on ekraanil avatud ning kuvab kontrollimiseks vajalikku infot sisaldavat QR-koodi.

Päästik
'''''''
Kasutusmall käivitub kui Hääletaja on käivitanud Kontrollrakenduse.

Põhiprotsess
''''''''''''
Põhiprotsess on ühesuunaline, tagasi liikumised protsessis eeldavad Kontrollrakenduse töö lõpetamist.

#. Kontrollrakendus laadib Kogumisteenusest seadistused.
#. Kontrollrakendus kuvab tervitusteksti.
#. Hääletaja suunab mobiilseadme kaamera Valijarakenduses kuvatavale QR-koodile.
#. Kontrollrakendus analüüsib QR-koodi ning tuvastab sealt hääle krüpteerimiseks kasutatud juhuarvu ning häält Kogumisteenuses identifitseeriva sessiooniidentifikaatori.
#. Kontrollrakendus pöördub sessiooniidentifikaatoriga Kogumisteenuse poole.
#. Kontrollrakendus verifitseerib Kogumisteenuse sertifikaadi ning kuvab Hääletajale info verifitseerimise õnnestumise kohta.
#. Kontrollrakendus laadib Kogumisteenusest alla Kogumisteenuse allkirjastatud hääle, Registreerimisteenuse kinnituse ja häälega seotud valikute nimekirjad (tehniline kasutusmall Hääle väljastamine kontrollimiseks Kogumisteenusest).
#. Kontrollrakendus veendub, et hääl vastab kontrollprotokolli nõuetele. Kontrollrakendus verifitseerib Registreerimisteenuse kinnitust ning Hääletaja allkirja. Kontrollrakendus kuvab Hääletajale info verifitseerimiste õnnestumise kohta, hääle allkirjastanud Hääletaja andmed ning küsimused, mille kohta antud hääles leidub krüpteeritud tahteavaldus.
#. Hääletaja käivitab kontrollalgoritmi.
#. Kontrollrakendus puhastab vaate hääle allkirjastanud Hääletaja andmetest.
#. Kontrollrakendus tuvastab krüpteeritud tahteavalduse sisu kasutades kontrollalgoritmi ja QR-koodist saadud juhuarvu.
#. Kontrollrakendus kuvab iga krüpteeritud tahteavalduse kohta Hääletajale küsimuse identifikaatorit ja tuvastatud valikut.

Laiendid
''''''''
#. Kui Kontrollrakendus tuvastab võrguühenduse puudumise, siis suunatakse Hääletaja võrguühendust aktiveerima.
#. Kui Kontrollrakendus tuvastab vea Kogumisteenuse sertifikaadis, siis lõpetatakse rakenduse töö veateatega.
#. Kui Kogumisteenusest ei saadud häält või Registreerimisteenuse kinnitust, siis juhendatakse Hääletajat Kliendituge informeerima ja lõpetatakse rakenduse töö veateatega.
#. Kui Kogumisteenusest tuleb teade, et konkreetse hääle kontrollimine ei ole enam võimalik (aeg või korrad ületatud), siis lõpetatakse rakenduse töö vastavasisulise teatega.
#. Kui Kogumisteenusest saadud hääl ei vasta nõuetele, siis lõpetatakse rakenduse töö veateatega.
#. Kui Hääletaja allkirja verifitseerimine ebaõnnestub, siis lõpetatakse rakenduse töö veateatega.
#. Kui Registreerimisteenuse kinnituse verifitseerimine ebaõnnestub, siis lõpetatakse rakenduse töö veateatega.
#. Kui kontrollalgoritm ei leia kandidaatide nimekirjast sobivat valikut, siis kuvatakse konkreetse valimise ja küsimuse identifikaatori järel veateade.
#. Kui QR koodis sisaldub rohkem kui ühe küsimuse kontrollkood, siis teostatakse kontroll kõigi viidatud küsimuste jaoks.
#. Kui Kogumisteenusest saadetud allkirjastatud hääles on krüpteeritud hääli, mille kohta puudub kontrollkood, siis esitab Kontrollrakendus Hääletajale nende valimiste/küsimuste identifikaatorid ning vastavasisulise hoiatuse. Ülejäänud hääled kontrollitakse.
#. Kui Kogumisteenusest saadetud allkirjastatud hääles ei ole kõiki hääli, mille kohta on kontrollkood, siis esitab Kontrollrakendus Hääletajale nende valimiste/küsimuste identifikaatorid ning vastavasisulise hoiatuse. Ülejäänud hääled kontrollitakse.
#. Kui kontrollalgoritm lõpetab veateatega, siis juhendatakse Hääletajat Kliendituge informeerima.

Järeltingimus
'''''''''''''
Kui hääle kontrollimisega ületatakse lubatud kontrollimiste limiit, siis Kogumisteenus rohkem häält kontrollida ei võimalda.


Korraldaja - Valijate nimekirja uuendamine
``````````````````````````````````````````

Kirjeldus
'''''''''
Valijate nimekirja loomine toimub Kogumisteenuse seadistamise käigus. Enne e-hääletamise algust on võimalik teostada valijate nimekirja kõiki uuendusi (lisamine, muutmine = kustutamine + lisamine, kustutamine suvalisel põhjusel). Peale e-hääletamise algust saab valijate nimekirja valijaid lisada, valijale tõkendit rakendada ja valijat ühest jaoskonnast teise tõsta. Valijate nimekiri on signeeritud.

Eeltingimus
'''''''''''

#. Kogumisteenus on seadistatud algse valijate nimekirjaga.
#. Kogumisteenus on Rahvastikuregistrist laadinud valijate nimekirja uuenduse.
#. Kogumisteenus on kontrollinud nimekirja signatuuri, nimekirja uuenduse vormingut ning kooskõlalisust juba laaditud konfiguratsiooniga ning on valmis uuendust rakendama.

Päästik
'''''''
Korraldaja valib Haldusteenusest valijate nimekirja uuenduse kinnitamise funktsionaalsuse.

Põhiprotsess
''''''''''''
#. Korraldaja valib Haldusteenuses rakendatava nimekirja.
#. Kogumisteenus nõuab Korraldajalt korralduse kinnitamist digitaalallkirjaga
#. Korraldaja allkirjastab korralduse oma digitaalallkirja andmise vahendiga
#. Kogumisteenus kontrollib Korraldaja allkirja ja volitusi süsteemis
#. Kogumisteenus eemaldab muudatuste nimekirjas eemaldamiskirjega identifitseeritud valijad.
#. Kogumisteenus lisab muudatuste nimekirjas lisamiskirjega identifitseeritud valijad, kontrollides valija unikaalsust isikukoodi alusel ning valimisjaoskonna olemasolu.
#. Kogumisteenus esitab lisatud ja eemaldatud valijate arvu ning nimekirja töötlemisel toimunud vigade logi.

Laiendid
''''''''
- Kui põhiprotsessi sammus 4 Korraldaja allkirja verifitseerimine või volituse kontroll ebaõnnestuvad, siis väljastatakse veateade nii Haldusteenuses kui logis. Ühtegi muudatust ei rakendata.
- Kui põhiprotsessi  sammus 5 kustutavat valijat ei leita isikukoodi alusel, siis kasutajale väljastatakse veateade, sündmus logitakse nimekirjade uuendamise aruandesse ning jätkatakse järgmise kirje töötlemist.
- Kui põhiprotsessi  sammus 6 lisatav valija isikukoodi alusel juba eksisteerib, siis kasutajale väljastatakse veateade, sündmus logitakse nimekirjade uuendamise aruandesse ning jätkatakse järgmise kirje töötlemist.
- Kui põhiprotsessi sammus 6 lisatakse valija mitte-eksisteerivasse jaoskonda, siis kasutajale väljastatakse veateade, sündmus logitakse nimekirjade uuendamise aruandesse ning jätkatakse järgmise kirje töötlemist.
- Tehniliste vigade tekkimisel kõigis põhivoo sammudes logitakse sündmus tehniliste vigade logisse. Kasutajale väljastatakse teade veasituatsiooni kohta, kõik muudatused võetakse tagasi.

Järeltingimus
'''''''''''''
Valijate nimekirja uuendused, mis olid Kogumisteenuse olekuga kooskõlalised, on Kogumisteenusesse laaditud, ebaõnnestunud uuenduste kohta on olemas raport.

Kogumisteenuse Haldur - Kogumisteenuse seisundi kuvamine
````````````````````````````````````````````````````````

Päästik
'''''''
Kogumisteenuse Haldur siseneb Kogumisteenuse Haldusteenusesse.

Põhiprotsess
''''''''''''

#. Haldusteenus autendib kasutaja ning tuvastab autenditud kasutaja volitused Haldusteenuses
#. Vastavalt volituste määrale kuvab Kogumisteenus alamhulka järgmisest informatsioonist:

   #. Valimise identifikaator, küsimused, hääletusetapp
   #. Töötavad teenusserverid
   #. Laetud nimekirjad - ringkonnad, valijad, valikud
   #. Hääletamise statistika
   #. Kogumisteenuse volitatud kasutajad
   #. Kogumisteenuse tehniline logi


Korraldaja - Hääletamise lõpetamine
```````````````````````````````````

Kirjeldus
'''''''''
Hääletamise lõpetamine lõpetab häälte vastuvõtmise Kogumisteenuse poolt. Hääletamise lõpetamine toimub järk-järgult – Hääletajad, kes on saanud kandidaatide nimekirja enne hääletamise ametlikku lõppu, peavad saama mõistliku aja jooksul hääletada.

Päästik
'''''''

#. Valimise seadistustes näidatud hääletamise lõpetamise aeg jõuab kätte.
#. Korraldaja edastab Haldusteenuse vahendusel digitaalallkirjastatud korralduse hääletamise lõpetamiseks.

Põhiprotsess
''''''''''''
1. Kogumisteenus vastab kõigile kandidaatide nimekirjade päringutele veateatega, kuid jätkab häälte talletamispäringute teenindamist.
2. Kogumisteenus lõpetab korralduses/konfiguratsioonis näidatud aja möödudes häälte vastuvõtmise.
3. Kogumisteenus vastab kõigile Valijarakendustelt tulevatele päringutele veateatega.

Järeltingimus
'''''''''''''
Kogumisteenus ei võta enam uusi hääli vastu.


Töötlusetapp
------------

Koguja - E-urni väljastamine
````````````````````````````

Kirjeldus
'''''''''

Koguja ekspordib e-urni Kogumisteenuse andmebaasist.

Eeltingimus
'''''''''''

Elektrooniline hääletamine on lõpetatud.

Päästik
'''''''

Koguja valib Haldusteenusest e-urni eksportimise funktsionaalsuse.

Põhiprotsess
''''''''''''

#. Kogumisteenus teeb häälte andmebaasist väljavõtte, mis sisaldab kogu talletamisteenuses sisalduvat häältega seotud infot.

Järeltingimus
'''''''''''''

E-urn on eksporditud.


Töötleja - E-urni verifitseerimine
``````````````````````````````````

Kirjeldus
'''''''''

Töötleja verifitseerib Kogujalt saadud e-urni kooskõlalisust ning selle vastavust Registreerimisteenuselt saadud informatsioonile.

Eeltingimus
'''''''''''

- E-urn on väljastatud
- Registreerimiskinnitused on väljastatud

Päästik
'''''''

Töötleja käivitab vastava funktsionaalsuse Töötlemisrakenduse kasutajaliideses.

Põhiprotsess
''''''''''''

#. Töötlemisrakendus laadib seadistused
#. Töötlemisrakendus kontrollib seadistuste digitaalallkirja
#. Töötlamisrakendus kontrollib seadistuste kooskõlalisust
#. Töötlemisrakendus laadib e-urni
#. Töötlemisrakendus kontrollib e-häälte digitaalallkirja
#. Töötlemisrakendus laadib registreerimiskinnitused
#. Töötlemisrakendus verifitseerib registreerimiskinnitused
#. Töötlemisrakendus kontrollib e-urni ja registreerimiskinnituste kooskõlalisust

Laiendid
''''''''
- Tehniliste vigade tekkimisel kõigis põhiprotsessi sammudes logitakse sündmus tehniliste vigade logisse. Kasutajale väljastatakse  teade veasituatsiooni kohta ning protsess peatatakse.
- Kasutajal on võimalus viia protsess läbi selliselt et vigadega seotud Hääletajad eraldatakse muust e-urnist. Tulemuseks on probleemsete häälte raport ja puhastatud e-urn.

Järeltingimus
'''''''''''''

E-urni verifitseerimise raport identifitseerib üheselt korrektsed ja probleemsed hääled.


Töötleja - Elektrooniliselt hääletanute nimekirja koostamine
````````````````````````````````````````````````````````````
Kirjeldus
'''''''''
Peale hääletamisperioodi lõppu koostab Töötleja valimisjaoskondade kaupa nimekirjad elektrooniliselt hääletanud isikutest. Nimekirjad on nii inim- kui masinloetavad.

Eeltingimus
'''''''''''

E-urn on edukalt verifitseeritud.

Päästik
'''''''
Töötleja käivitab vastava funktsionaalsuse Töötlemisrakenduse kasutajaliideses.

Põhiprotsess
''''''''''''
1. Töötlemisrakendus tühistab korduvad hääled, jättes iga Hääletaja kohta alles ajaliselt kõige viimasena antud hääle.
2. Töötlemisrakendus sorteerib kehtivad hääled valimisjaoskondade kaupa.
3. Töötlemisrakendus loob elektrooniliselt hääletanute nimekirjade faili.

    1. Iga kehtiva hääle kohta lisab Töötlemisrakendus faili ühe kirje, kus sisaldub isikukood, nimi, valimisjaoskonna number ja rea number valimisringkonna nimekirjas. Andmed võetakse valijate nimekirjast.

4. Töötlemisrakendus esitab vastuvõetud häälte koguarvu ning elektrooniliselt hääletanute arvu.
5. Töötleja salvestab nimekirja välisele andmekandjale.

Laiendid
''''''''
Kui põhiprotsessi sammus 3 ei tuvastata isikukoodi valijate nimekirjast, siis logitakse vastavasisuline informatsioon ning jätkatakse põhiprotsessiga.

Järeltingimus
'''''''''''''
Kehtivad hääled on  kantud e-hääletanute nimekirja. Korduvad hääled on tühistatud.

Töötleja - Miksimisele/Lugemisele minevate e-häälte anonümiseerimine
````````````````````````````````````````````````````````````````````

Kirjeldus
'''''''''

Töötleja rakendab tühistus- ja ennistusnimekirju ning loob lugemisele minevate e-häälte anonümiseeritud hulga.

Eeltingimus
'''''''''''

E-urn on edukalt verifitseeritud ning elektrooniliselt hääletanute nimekiri koostatud.

Päästik
'''''''

Töötleja käivitab vastava funktsionaalsuse Töötlemisrakenduse kasutajaliideses.

Põhiprotsess
''''''''''''

    Töötlemisrakendus verifitseerib tühistus- ja ennistusnimekirjade digitaalallkirjad.
    Töötlemisrakendus kontrollib tühistus- ja ennistusnimekirjade kooskõlalisust järjekorras.
    Töötlemisrakendus rakendab tühistus- ja ennistusnimekirju järjekorras.
    Töötlemisrakendus koostab miksimisele/lugemisele minevate häälte nimekirja eraldades krüptogrammid digitaalallkirjadest.
    Töötlemisrakendus väljastab lõpliku elektrooniliselt hääletanute nimekirja masinloetavas vormingus.

Laiendid
''''''''

 - Kui põhiprotsessi sammus 1 Korraldaja allkirja verifitseerimine või volituse kontroll ebaõnnestuvad, siis väljastatakse veateade Ühtegi muudatust ei rakendata.

Järeltingimus
'''''''''''''

Miksimisele/lugemisele minevate häälte hulk on koostatud, elektrooniliselt hääletanute nimekiri on väljastatud.

Miksija - Miksimine
```````````````````

Kirjeldus
'''''''''

Eeltingimus
'''''''''''

Päästik
'''''''

Põhiprotsess
''''''''''''

#. Miksija käivitab Miksimisrakenduse ning laadib anonüümistatud e-hääled
#. Miksimisrakendus rerandomiseerib ning permuteerib e-hääled, tulemuseks on uued e-hääled
#. Miksimisrakendus genereerib tõestuse, et uued e-hääled on algsete e-häältega sisuliselt samaväärsed
#. Miksimisrakendus väljastab nii miksitud hääled kui Miksimistõendi

Laiendid
''''''''

Järeltingimus
'''''''''''''

Algsete häältega tõestatavalt samaväärsed, miksitud hääled on väljastatud.

Lugemisetapp
------------

Lugeja - Elektroonilise hääletamise tulemuse kindlaks tegemine
``````````````````````````````````````````````````````````````

Kirjeldus
'''''''''
Tühistusperioodi lõppedes sorteeritakse ümbrikud valimisringkondade kaupa. Välised ümbrikud avatakse. s.t. digitaalallkirjad eemaldatakse, järgi jäävad häälte salastamise võtmega krüpteeritud hääled, mis dekrüpteeritakse Võtmerakendusega.

Põhiprotsess
''''''''''''
#. Lugeja algatab häälte kokku lugemise Võtmerakendusega.
#. Võtmehaldurid aktiveerivad vastavalt võtmehalduse protseduuridele häälte avamise võtme.
#. Võtmerakendus  loeb krüpteeritud hääled väliselt andmekandjalt.
#. Võtmerakendus teostab häältefaili tehnilise kontrolli ning algatab häälte krüpteerimise.
#. Võtmerakendus kontrollib dekrüpteeritud hääle vastavust avakujul hääle vormingule.
#. Võtmerakendus kontrollib hääle kehtivust, veendudes et dekrüpteerimisel selgunud kandidaat kuulus antud ringkonnas valikute hulka.
#. Võtmerakendus summeerib arvesse minevad hääled jaoskondade, ringkondade ja kandidaatide kaupa.
#. Võtmerakendus kannab hääletamistulemuse andmekandjale.
#. Hääletamistulemus imporditakse valimiste infosüsteemi.

Laiendid
''''''''
- Tehniliste vigade tekkimisel kõigis põhiprotsessi sammudes logitakse sündmus tehniliste vigade logisse. Kasutajale väljastatakse  teade veasituatsiooni kohta ning protsess peatatakse.
- Põhiprotsessi sammudes 4 ja 5 tehtavate kontrollide ebaõnnestumise korral loetakse antud hääl kehtetuks ning protsessi jätkatakse järgmise krüpteeritud hääle juurest.

Järeltingimus
'''''''''''''
Hääletamistulemus on sisestatud valimiste infosüsteemi.

Auditeerimisetapp
-----------------

Audiitor - Auditeerimine
````````````````````````

Kirjeldus
'''''''''

Peale hääletamistulemuse väljaselgitamist on Audiitoril võimalik kontrollida Miksimistõendit ja Lugemistõendit.


Eeltingimus
'''''''''''

- Hääletamistulemus, krüptogrammid ning lugemistõend on olemas
- Miksimiseelsed krüptogrammid, miksimisjärgsed krüptogrammid ning miksimistõend on olemas

Põhiprotsess
''''''''''''

1. Audiitor kasutab Auditirakendust Lugemistõendi kontrollimiseks
2. Audiitor kasutab Auditirakendust Miksimistõendi kontrollimiseks

Laiendid
''''''''

Järeltingimus
'''''''''''''

Lugemistõend ja Miksimistõend on kontrollitud.


Tehnilised kasutusmallid
------------------------

Autentimine Valijarakenduses
````````````````````````````

Kirjeldus
'''''''''
Valijarakenduse vahendusel ja elektroonilise isikutunnistuse abil tuvastab Kogumisteenus Hääletaja isiku.

Päästik
'''''''
Kasutusmall käivitatakse kui Hääletaja käivitab Valijarakenduse ning suundub hääletama.

Põhiprotsess – ID kaart
'''''''''''''''''''''''
1. Hääletaja sisestab ID-kaardi lugejasse.
2. Valijarakendus pöördub Kogumisteenuse poole protokolli algatamiseks.
3. Kogumisteenus saadab Valijarakendusele oma sertifikaadi.
4. Valijarakendus kontrollib Kogumisteenuse sertifikaati.
5. Kogumisteenus nõuab Hääletaja autentimist TLS-protokolli kohaselt.
6. Valijarakendus küsib valijalt ID-kaardi autentimisvõtme kasutamiseks PIN1 koodi.
7. Hääletaja sisestab PIN1 koodi.
8. Valijarakendus ja Kogumisteenus viivad läbi TLS-protokolli, Kogumisteenuse saadetakse Hääletaja sertifikaat.
9. Kogumisteenus kontrollib Hääletaja sertifikaati.
10. Kogumisteenus tuvastab Hääletaja isikukoodi.

Põhiprotsess – Mobiil-ID
''''''''''''''''''''''''
1. Hääletaja sisestab Valijarakendusse oma Mobiil-ID SIM kaarti sisaldava mobiiltelefoni numbri.
2. Valijarakendus pöördub Kogumisteenuse poole protokolli algatamiseks.
3. Kogumisteenus saadab Valijarakendusele oma sertifikaadi.
4. Valijarakendus kontrollib Kogumisteenuse sertifikaati.
5. Valijarakendus saadab telefoninumbri Kogumisteenusele.
6. Kogumisteenus algatab autentimise Mobiil-ID teenuse kaudu.
7. Kogumisteenus saadab Mobiil-ID kontrollkoodi Valijarakendusele, mis kuvab seda valijale.
8. Hääletaja saab mobiiltelefonile autentimissõnumi.
9. Hääletaja võrdleb autentimissõnumi kontrollkoodi Valijarakenduses kuvatavaga.
10. Hääletaja sisestab autentimisvõtme kasutamiseks PIN1 koodi.
11. Kogumisteenus ja Mobiil-ID teenus viivad läbi autentimise, Kogumisteenuse saadetakse Hääletaja sertifikaat.
12. Kogumisteenus kontrollib Hääletaja sertifikaati.
13. Kogumisteenus tuvastab Hääletaja isikukoodi.
14. Valijarakendus küsib regulaarselt, kas autentimine on toimunud, Kogumisteenus vastab.

Laiendid
''''''''
- Kui ID-kaarti ei ole lugejas, siis katkestab Valijarakendus töö veateatega (ID-kaart).
- Kui Kogumisteenus ei ole kättesaadav, siis katkestab Valijarakendus töö veateatega.
- Kui Kogumisteenuse sertifikaadi kehtivust ei õnnestu kontrollida või kui sertifikaat ei vasta Kogumisteenuse nimele, siis katkestab Valijarakendus töö veateatega.
- Kui ID-kaardi kasutamine ebaõnnestub, siis katkestab Valijarakendus töö veateatega. (ID-kaart)
- Kui Kogumisteenuses annab Hääletaja sertifikaadi kontroll negatiivse tulemuse, siis saadetakse Valijarakendusele veateade ning autentimine katkestatakse.
- Kui Mobiil-ID protokolli läbiviimine ebaõnnestub, siis vastab Kogumisteenus Valijarakenduse päringule veateatega ning autentimine katkestatakse. (Mobiil-ID)

Järeltingimus
'''''''''''''
Kogumisteenus teab Hääletaja sertifikaati ning on tuvastanud sellest isikukoodi.

Digitaalne allkirjastamine Valijarakenduses
```````````````````````````````````````````

Kirjeldus
'''''''''
Valijarakenduse abil antakse digitaalallkirja vahendi abil krüpteeritud häälele digitaalallkiri.

Eeltingimus
'''''''''''
Hääletaja on teinud oma valikud ja kinnitanud need. Valijarakendus on valikud krüpteerinud Kogumisteenuse avaliku võtmega.
Mobiil-ID korral teab Kogumisteenus juba Hääletaja Mobiil-ID SIM kaardiga mobiiltelefoninumbrit, kuna autentimine on edukalt toimunud.

Päästik
'''''''
Hääletaja on valinud Valijarakenduses hääletamise funktsionaalsuse, kõik valikud on krüpteeritud Kogumisteenuse avaliku võtmega.

Põhiprotsess – ID-kaart
'''''''''''''''''''''''
1. Valijarakendus moodustab BES vormingus BDOC konteineri.
2. Valijarakendus lisab konteinerisse andmefailidena krüpteeritud valikud.
3. Valijarakendus moodustab allkirjastatava räsi BDOC standardi kohaselt.
4. Valijarakendus algatab räsi signeerimise ID-kaardiga.
5. Hääletaja sisestab PIN2 koodi allkirjastamisvõtme kasutamiseks.
6. Valijarakendus lisab Hääletaja signatuuri ja allkirjastamissertifikaadi BDOC konteinerile.

Põhiprotsess – Mobiil-ID
''''''''''''''''''''''''
1. Valijarakendus moodustab BES vormingus BDOC konteineri.
2. Valijarakendus lisab konteinerisse andmefailidena krüpteeritud valikud.
3. Valijarakendus saadab konteineri Kogumisteenusele.
4. Kogumisteenus moodustab allkirjastatava räsi BDOC standardi kohaselt.
5. Kogumisteenus algatab räsi signeerimise Mobiil-ID teenusega.
6. Kogumisteenus saadab Mobiil-ID kontrollkoodi Valijarakendusele, mis kuvab seda valijale.
7. Hääletaja saab mobiiltelefonile allkirjastamissõnumi.
8. Hääletaja võrdleb allkirjastamissõnumi kontrollkoodi Valijarakenduses kuvatavaga.
9. Hääletaja sisestab allkirjastamisvõtme kasutamiseks PIN2 koodi.
10. Kogumisteenus ja Mobiil-ID teenus viivad läbi allkirjastamise, Kogumisteenuse saadetakse digitaalselt allkirjastatud ja kehtivuskinnitusega TM vormingus BDOC konteiner.
11. Valijarakendus küsib regulaarselt, kas allkirjastamine on toimunud, Kogumisteenus vastab.
12. TM vormingus BDOC konteiner edastatakse Valijarakendusele verifitseerimiseks.

Laiendid
''''''''
- Kui ID-kaarti ei ole lugejas, siis katkestab Valijarakendus töö veateatega (ID-kaart).
- Kui Kogumisteenus ei ole kättesaadav, siis katkestab Valijarakendus töö veateatega.
- Kui ID-kaardi kasutamine ebaõnnestub, siis katkestab Valijarakendus töö veateatega. (ID-kaart)
- Kui Mobiil-ID protokolli läbiviimine ebaõnnestub, siis vastab Kogumisteenus Valijarakenduse päringule veateatega ning allkirjastamine katkestatakse. (Mobiil-ID)

Järeltingimus
'''''''''''''
Hääletaja hääl on digitaalselt allkirjastatud.

Digitaalallkirja kehtivuse kontroll Kogumisteenuses
```````````````````````````````````````````````````

Kirjeldus
'''''''''
Kogumisteenus kontrollib digitaalallkirju ID-kaardi ja Mobiil-ID'ga allkirjastatud e-häältel ning seadistamisel ja tühistamisel kasutatavatel sisendfailidel.

Eeltingimus
'''''''''''
Kogumisteenusede sertifikaatide konfiguratsioon on tehtud. Kogumisteenuse kehtivuskinnitusteenuse konfiguratsioon on tehtud.

Päästik
'''''''
Kogumisteenus algatab digitaalallkirja kehtivuse kontrolli.

Põhiprotsess
''''''''''''
1. Kogumisteenus tuvastab profiili, millest lähtudes kontroll läbi viia:

    1. ID-kaardiga allkirjastatud hääl võib olla BES või TM vormingus, tohib sisaldada mitut andmefaili ja täpselt ühte allkirja.
    2. Mobiil-ID'ga allkirjastatud hääl peab olema TM vormingus, tohib sisaldada mitut andmefaili ja täpselt ühte allkirja.
    3. Muud failid peavad olema TM vormingus, tohivad sisaldada täpselt ühte andmefaili ja täpselt ühte allkirja.

2. Kogumisteenus veendub allkirja vastavuses nõutud profiilile.
3. Kui allkiri on BES vormingus, siis pöördub Kogumisteenus kehtivuskinnitusteenuse poole sertifikaadi kehtivuse kontrolliks. Kogumisteenus liidab kehtivuskinnituse digitaalallkirjale, tulemuseks on TM vormingus allkiri.
4. Kogumisteenus kontrollib TM vormingus allkirja:

    1. Signatuuri.
    2. Sertifikaadi kehtivuskinnitust.

Laiendid
''''''''
- Tehniliste vigade tekkimisel kõigis põhiprotsessi sammudes logitakse sündmus tehniliste vigade logisse. Välja kutsuvale protsessile väljastatakse  teade veasituatsiooni kohta.
- Kehtivuskinnituse laadimise ebaõnnestumisel põhiprotsessi sammus 3 logitakse sündmus tehniliste vigade logisse. Välja kutsuvale protsessile väljastatakse  teade veasituatsiooni kohta.
- Kontrollide ebaõnnestumisel põhiprotsessi sammus 4 logitakse sündmus tehniliste vigade logisse. Välja kutsuvale protsessile väljastatakse teade veasituatsiooni kohta.

Valikute nimekirjade väljastamine Valijarakendusele
```````````````````````````````````````````````````

Kirjeldus
'''''''''
Kogumisteenus väljastab Valijarakendusele valikute nimekirja.

Eeltingimus
'''''''''''
Hääletaja autentimine on õnnestunult toimunud.

Päästik
'''''''
Valijarakendus on pöördunud Kogumisteenuse poole valikute nimekirja alla laadimiseks.

Põhiprotsess
''''''''''''
1. Kogumisteenus kontrollib isikukoodi kasutades, millisesse valimisringkonda Hääletaja kuulub.
2. Kogumisteenus kontrollib, kas Hääletaja kohta on juba mõni hääl talletatud.
3. Kogumisteenus saadab Valijarakendusele ringkonnakohase kandidaatide nimekirja kõigi valimise kohta, kus isikukood on nimekirja kantud ning informatsiooni võimaliku korduvhääletamise kohta.

Laiendid
''''''''
- Tehniliste vigade tekkimisel kõigis põhiprotsessi sammudes logitakse sündmus tehniliste vigade logisse. Valijarakendust teavitatakse protsessi ebaõnnestumisest.
- Kui põhiprotsessi sammus 2 on selge, et Hääletaja ei kuulu käimasoleva valimise valijate nimekirja, siis saadetakse Valijarakendusele vastavasisuline veateade.

Hääle talletamine Kogumisteenuses
`````````````````````````````````

Kirjeldus
'''''''''
Kogumisteenus talletab Hääletaja hääle.

Eeltingimus
'''''''''''
Hääle digitaalne allkirjastamine Valijarakenduses on edukalt toimunud.

Päästik
'''''''
Valijarakendus on saatnud hääle talletamisele.

Põhiprotsess
''''''''''''
1. Kogumisteenus kontrollib hääle digitaalallkirja (tehniline kasutusmall Digitaalallkirja kehtivuse kontroll Kogumisteenuses). Kontrolli käigus tagatakse muu hulgas kehtivuskinnituse lisamine allkirjale.
2. Kogumisteenus kontrollib hääle allkirjastanud ja valikute nimekirja laadimisel TLS protokollis autenditud isikukoodide samasust.
3. Kogumisteenus registreerib hääle vastuvõtmise Registreerimisteenuses.
4. Kogumisteenus genereerib häälega seostatud unikaalse identifikaatori.
5. Kogumisteenus salvestab hääle koos unikaalse identifikaatoriga.
6. Kogumisteenus tagastab Valijarakendusele õnnestumise teate koos häälele viitava unikaalse identifikaatoriga ja Registreerimisteenuse kinnitusega

Laiendid
''''''''
- Tehniliste vigade tekkimisel kõigis põhiprotsessi sammudes logitakse sündmus tehniliste vigade logisse. Valijarakendust teavitatakse protsessi ebaõnnestumisest. Häält ei talletata.
- Kui põhiprotsessi sammus 1 digitaalallkirja kontroll ebaõnnestub, siis logitakse sündmus tehniliste vigade logisse. Valijarakendust teavitatakse protsessi ebaõnnestumisest. Häält ei talletata.
- Kui põhiprotsessi sammus 2 isikukoodid erinevad, siis logitakse sündmus tehniliste vigade logisse. Valijarakendust teavitatakse protsessi ebaõnnestumisest. Häält ei talletata.

Järeltingimus
'''''''''''''
Digitaalselt allkirjastatud hääl on varustatud kehtivuskinnitusega, Kogumisteenuse poolse kinnitusega ning talletatud Kogumisteenuses seostatuna unikaalse identifikaatoriga.

Hääle väljastamine kontrollimiseks Kogumisteenusest
```````````````````````````````````````````````````

Kirjeldus
'''''''''
Kogumisteenus väljastab Kontrollrakendusele kontrollimiseks allkirjastatud hääle ja registreerimiskinnituse.

Päästik
'''''''
Kontrollrakendus on esitanud Kogumisteenusele unikaalse identifikaatoriga hääle kontrollimise päringu.

Põhiprotsess
''''''''''''
1. Kogumisteenus kontrollib, kas unikaalsele identifikaatorile vastab elektrooniline hääl.
2. Kogumisteenus kontrollib, kas unikaalsele identifikaatorile vastava hääle kontrollimine on veel võimalik vastavalt aja ja kordade tingimustele.
3. Kogumisteenus suurendab konkreetse hääle kontrollimise kordade loendurit.
4. Kogumisteenus väljastab Kontrollrakendusele unikaalse identifikaatori poolt viidatud hääle koos ringkonna valikute nimekirjaga ning häälele vastava registreerimiskinnitusega.

Laiendid
''''''''
- Tehniliste vigade tekkimisel kõigis põhiprotsessi sammudes logitakse sündmus tehniliste vigade logisse. Kontrollrakendust teavitatakse protsessi ebaõnnestumisest. Häält ei väljastata.
- Kui põhiprotsessi sammus 1 ei leita unikaalsele identifikaatorile vastavat häält, siis logitakse sündmus tehniliste vigade logisse. Kontrollrakendust teavitatakse protsessi ebaõnnestumisest. Häält ei väljastata.
- Kui põhiprotsessi sammus 2 tehtavad kontrollid annavad negatiivse tulemuse, siis logitakse sündmus tehniliste vigade logisse. Kontrollrakendust teavitatakse protsessi ebaõnnestumisest. Häält ei väljastata.

Elektrooniliste häälte avalduse alusel tühistamine/ennistamine
``````````````````````````````````````````````````````````````

Kirjeldus
'''''''''
Korraldaja poolt kinnitatud digitaalallkirjastatud tühistus-/ennistuskannete faili import Töötleja poolt. Töötlemisrakendus teostab tühistused/ennistused, logides tegevused. Töötlejale kuvatakse tühistamise/ennistamise tulem – aruanne tühistustest/ennistustest koos tulemusega.

Päästik
'''''''

Töötlemisrakendus käivitab vastava funktsionaalsuse.

Põhiprotsess
''''''''''''
#. Töötlemisrakendus verifitseerib avalduse digitaalallkirja.
#. Töötlemisrakendus kontrollib, et avaldus oleks allkirjastatud volitatud kasutaja poolt.
#. Töötlemisrakendus loeb failist kandeid ning viib operatsioonid täide märkides hääle kas tühistatuks või ennistatuks.
#. Kogumisteenus kuvab avalduse rakendamise tulemusi.

Laiendid
''''''''
- Kui põhiprotsessi sammus 1 tuvastatakse digitaalallkirja mittekehtivus, siis logitakse tehniliste vigade logisse vastav kanne ning protsess peatatakse.
- Kui põhiprotsessi sammus 2 tuvastatakse, et avalduse allkirjastajal ei ole vastavaid volitusi, siis logitakse tehniliste vigade logisse vastav kanne ning protsess peatatakse.
- Kui põhiprotsessi sammus 3 tuvastatakse, et isikukoodiga identifitseeritud Hääletaja hääl on mittesobivas olekus või Hääletaja ei ole hääletanud, siis logitakse aruandesse vastav kanne ning protsessi jätkatakse järgmise kandega.
- Tehniliste vigade tekkimisel kõigis põhivoo sammudes logitakse sündmus tehniliste vigade logisse. Kasutajale väljastatakse  teade veasituatsiooni kohta ning protsess peatatakse.

Järeltingimus
'''''''''''''
Tühistatud hääled ei lähe lugemisele, ennistatud hääled lähevad lugemisele.

.. vim: sts=3 sw=3 et:
