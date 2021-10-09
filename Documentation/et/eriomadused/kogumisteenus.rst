..  IVXV eriomadused

Kogumisteenuse omadused
-----------------------

Kogumisteenuse arhitektuur
^^^^^^^^^^^^^^^^^^^^^^^^^^

#. Kogumisteenuse tarkvara on projekteeritud jätkusuutlikult;

#. Kogumisteenuse alamteenused on projekteeritud mikroteenuste komplektina;

#. Kogumisteenuse osad komponendid on taaskasutatavad;

#. Kogumisteenus on evitatav mitmeastmeliselt (*n-tier application*).


Kogumisteenuse funktsionaalsus
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

#. Kogumisteenus toetab järgnevaid valimiste liike:

   #. Riigikogu valimised;

   #. Kohaliku volikogu valimised;

   #. Euroopa Parlamendi valimised;

   #. Rahvahääletused.

#. Kogumisteenus toetab mitme erineva valimise üheaegset läbiviimist;

#. Kogumisteenus toetab mitme küsimusega valimiste läbiviimist;

#. Kogumisteenus toetab mitut paralleelset autentimismeetodit;

#. Kogumisteenuses toimub hääletaja tuvastamine Eesti ametlike
   e-identimissüsteemide abil (ID-kaart, digi-ID, mobiil-ID);

#. Kogumisteenuses toimub hääle kontrollimine
   vastavalt üldraamistikus kirjeldatule;

#. Kogumisteenuses toimub hääle talletamine
   vastavalt üldraamistikus kirjeldatule;

#. Kogumisteenus viib läbi ja logib korrapärastelt enesediagnostikat;

#. Kogumisteenuse operaatorite tegevused on kontrollitavad;

#. Kogumisteenuse toetab hääletajat hääletamisprotsessis;

#. Hääletamisprotsessi tehnilised logid on üleantavad korraldajale;

#. Kogumisteenuse logid on ingliskeelsed;

#. Kogumisteenus logib kõik olulised sündmused;

#. Kogumisteenus toodab tegevusstatistikat valimise kulgemise kohta;

#. Kogumisteenuse logimine lähtub parimast praktikast;

#. Kogumisteenuse logiväljund on seadistatav.


Kogumisteenuse andmestruktuurid ja protokollid
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

#. E-valimiskasti edastamiseks kasutatakse nõuetekohast protokolli;

#. Kogumisteenuse komponentide vahelised protokollid
   võimaldavad mõlemapoolset autentimist;

#. Kogumisteenus vormistab hääle BDOC vormingus (EVS 821:2014),
   vastavalt seadistatud allkirjastamismeetodi(te)le;

#. Kogumisteenus saadab kontrollrakendusele krüptitud hääle täiemahulise
   digitaalallkirja struktuurina, mis sisaldab ka registreerimistõendit;

#. Kogumisteenuse seansid on serialiseeruvad.


Kogumisteenuse evitamine
^^^^^^^^^^^^^^^^^^^^^^^^

#. Kogumisteenus on paigaldatav Ubuntu 20.04 LTS (Focal Fossa) platvormile;

#. Kogumisteenuse platvorm sobitub turulolevate
   pilveteenustega (Amazon Web Services);

#. Kogumisteenus on lihtsalt paigaldatav ühele riistvarale;

#. Kogumisteenuse on projekteeritud töötama ka virtualiseeritud keskkonnas.


Kogumisteenuse seadistamine
^^^^^^^^^^^^^^^^^^^^^^^^^^^

#. Kogumisteenusel on valimiste seadistamiseks ja
   juhtimiseks on kasutajasõbralik haldusliides;

#. Kogumisteenuse haldamiseks on käsureapõhine haldusliides;

#. Kogumisteenuse haldusteenuse kasutajate pääsuhaldus on lahendatud
   paindlikult.

#. Kogumisteenuses on valimised seadistatavad
   hääletamiseelsel ja hääletamise etapil;

#. Kogumisteenuse valimiste konfiguratsioon
   rakendatakse atomaarse seadistusfailina;

#. Konfiguratsioon läbib enne rakendamist
   tervikluse ja kooskõlalisuse kontrolli;

#. Kõik korraldaja korraldused peavad olema signeeritud;

#. Kõigi süsteemiga seotud toimingute teostamise tingumused
   vastavad rangelt käimasolevale e-hääletuse etapile;

#. Hääletamise autentimismeetodid on konfigureeritavad;

#. Kogumisteenus võimaldab valikute nimekirja laadimist;

#. Kogumisteenus võimaldab valijate nimekirja laadimist ja uuendamist;

#. Nimekirjade toimingute ajalugu on haldusliidesest jälgitav.


Kogumisteenuse liidesed
^^^^^^^^^^^^^^^^^^^^^^^

#. Kogumisteenuse kasutajaliides on eesti keeles;

#. Kogumisteenusel on masinliides valijarakenduse jaoks;

#. Kogumisteenusel on masinliides kontrollrakenduse jaoks;

#. Kogumisteenus on liidestatud Eesti ametlike identimissüsteemidega;

#. Kogumisteenuse registreerimisteenus toetab OCSP liidest;

#. Kogumisteenuse registreerimisteenus toetab RFC3161 liidest;

#. Kogumisteenuse tuvastusteenus toetab OCSP liidest;

#. Kogumisteenuse tuvastusteenus toetab Mobiil-ID liidest;

#. Kogumisteenuse kehtivuskinnitusteenus toetab OCSP liidest;

#. Kogumisteenuse allkirjastamisteenus toetab Mobiil-ID liidest;

#. Kogumisteenusel on klienditeenindaja liides;

#. Kogumisteenuse liidesed on ligipääsetavad ainult määratud osapooltele;

#. Kogumisteenusel on API uute tuvastusteenuste liidestamiseks;

#. Kogumisteenusel on API uute autentimismeetodite liidestamiseks.


Kogumisteenuse käideldavus
^^^^^^^^^^^^^^^^^^^^^^^^^^

#. Kogumisteenus võimaldab pakkuda kõrgkäideldavat teenust;

#. Kogumisteenust on võimalik käivitada mitmes
   üksteisest eraldatud ja rööbiti töötava isendina;

#. Kogumisteenuse rööbiti töötavad instantsid toetavad tõrkesiirde (failover)
   funktsionaalsust;

#. Kogumisteenuse hääle talletamise ja registreerimise transaktsioonid on
   terviklikud (ACID);

#. Kogumisteenus säilitab kõik transaktsioonid;

#. Kogumisteenus kasutab tõrkekindlat andmesäilitustehnoloogiat;

#. Kogumisteenuse andmeid on võimalik terviklikult varundada;

#. Kogumisteenuse läbilaskevõime on vähemalt 100 häält sekundis.


Kogumisteenuse monitooring
^^^^^^^^^^^^^^^^^^^^^^^^^^

#. Kogumisteenuse tehnilisi andmeid ja tegevust on võimalik monitoorida
   (tehniline monitooring ja tegevusmonitooring);

#. Kogumisteenuse monitooringulahendus on kogumisteenusest eraldiseisev;

#. Kogumisteenuse serverisüsteemis on logimiseks vajalik masinliides;

#. Kogumisteenusest saadetakse logi monitooringuserverisse;

#. Kogumisteenusest monitooringuserverisse saadetakse logi üle turvalise kanali;

#. Kogumisteenuse logid ajatembeldatakse monitooringuserveris;

#. Monitooringuserver väljastab hoiatusteateid nii e-posti kui SMS-i kaudu.


Tehniline monitooring
"""""""""""""""""""""

#. Koguja tehniliseks monitooringuks kasutatakse Zabbixit;

#. Kogumisteenus väljastab Zabbixile monitooringuandmeid.


Tegevusmonitooring
""""""""""""""""""

#. Monitooringusüsteem genereerib hääletamise
   kulgemise kohta üksikasjalikku statistikat;

#. Monitooringusüsteem genereerib perioodiliselt hääletamise üldstatistikat;

#. Hääletamise üldstatistika on kättesaadav haldusliidese kaudu.


Monitooringu liidesed
"""""""""""""""""""""

#. Hääletaja kohta registreeritud sündmused on
   klienditeenindaja liidese kaudu jälgitavad;

#. Monitooringuandmed ja statistika on kättesaadav masinliidese kaudu;

#. Monitooringuserveri kogutud andmetele on võimalik ligi pääseda
   pääsunimekirjade alusel.


Kogumisteenuse varundamine
^^^^^^^^^^^^^^^^^^^^^^^^^^

#. Varundada on võimalik e-valimiskasti, logifaile ja süsteemi seadistusi;

#. Varundada on võimalik irdmeediale (ja muule selleks ettenähtud meediale);

#. Varundamine on automatiseeritud, varundamise meetod ja ajastus on seadistatav;

#. Varukoopiat on võimalik krüptida konfigureeritava krüptosüsteemiga;

#. Varukoopia on teisaldatav kogumisteenuse
   käitluskohast füüsiliselt erinevasse kohta.
