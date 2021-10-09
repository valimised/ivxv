..  IVXV protokollid

===================
Suhtlusprotokollid
===================

Liides
------

Kogumisteenuse valijale suunatud mikroteenused suhtlevad valijarakendusega ja
kontrollrakendusega JSON-RPC protokolli vahendusel.

:id: JSON-RPC päringuidentifikaator
:method: RPC-meetod
:params: Konkreetse RPC-meetodi parameetrid

.. literalinclude:: ../../common/examples/json.rpc.method.query.json
   :language: json
   :linenos:

:error: Võimalik veainfo või ``null`` vea puudumisel
:id: JSON-RPC päringuidentifikaator, peab ühtima päringus kasutatud id-ga
:result: Meetodipõhine vastusandmestruktuur

.. literalinclude:: ../../common/examples/json.rpc.method.response.json
   :language: json
   :linenos:

Esimese päringuvahetuse käigus mõne IVXV mikroteenusega väljastatakse suhtlevale
rakendusele HEX-kodeeritud unikaalne seansiidentifkaator (``result.SessionID``),
mida rakendus kasutab edaspidi kõigis kogumisteenuse suunalistes päringutes
(``params.SessionID``). Seansiidentifikaatori abil seostatakse
hääletamisega seotud RPC-päringud üheks seansiks. Seostamine on informatiivne
ning selle eesmärk on logianalüüsi lihtsustamine, hääle ringkonnakuuluvust jm.
sisulisi aspekte puudutavad otsused tehakse digiallkirjastatud andmete põhjal.

Transpordiprotokollina on kasutusel TLS. Krüpteeritud kanali termineerimine
toimub konkreetses mikroteenuses. Võimaldamaks koormuse jaotamist ning
mikroteenuste paindlikku evitamist kasutatakse TLS-i SNI laiendust, mis lubab
vahendusteenusel TLS voogu termineerimata õigesse mikroteenusinstantsi suunata.
Vahendusteenus on tüüpiliselt kättesaadav kogumisteenuse välise liidese pordis
443.

Valikute nimekirja hankimine
----------------------------

Valikute nimekirja hankimine tähendab valijarakenduse suhtlemist
nimekirjateenusega (SNI ``choices.ivxv.invalid``). Valikute nimekirja hankimine
eeldab valija autentimist ning tema ringkonnakuuluvuse tuvastamist.

Valijarakendus teeb päringu ``RPC.VoterChoices`` nimekirjade hankimiseks.

:params.AuthMethod: Toetatud valikud on meetodid ``tls`` ja ``ticket``.
:params.OS: Operatsioonisüsteem, millel valijarakendust kasutatakse.

Päring ``RPC.VoterChoices`` ID-kaardiga autentimise korral - autentimine toimub
TLS-protokolli tasemel päringu töötlemise ajal kasutades ID-kaardi
autentimissertifikaati.

.. literalinclude:: ../../common/examples/id.rpc.voterchoices.query.json
   :language: json
   :linenos:

Päring ``RPC.VoterChoices`` Mobiil-ID'ga autentimise korral - päringu
sooritamiseks tuleb eelnevalt kasutada Mobiil-ID vahendusteenuse (SNI
``mid.ivxv.invalid``) abi allkirjastatud autentimistõendi saamiseks.

:params.AuthToken: Autentimisteenuse vahendusel allkirjastatud tõend, mis
                   sisaldab endas valija unikaalset identifikaatorit.

:params.SessionID: Kuna Mobiil-ID korral on nimekirja hankimisele eelnenud
                   interaktsioon autentimistõendi saamiseks, on olemas
                   seansiidentifikaator, mida tuleb kasutada.

.. literalinclude:: ../../common/examples/mid.rpc.voterchoices.query.json
   :language: json
   :linenos:

Nimekirjateenuse vastus päringule ``RPC.VoterChoices``.

:result.Choices: Valija ringkonnakuuluvuse identifikaator ``VoterDistrict``
:result.List: BASE64-kodeeritud ringkonna valikute nimekiri ``DistrictChoices``
:result.Voted: Kui valija on juba hääletanud, siis ``true``, vastasel juhul
               seda välja vastuses ei ole.

.. literalinclude:: ../../common/examples/id.rpc.voterchoices.response.json
   :language: json
   :linenos:

Võimalikud veateated päringu ``RPC.VoterChoices`` korral.

:BAD_CERTIFICATE: Viga valija isikutuvastussertifikaadiga.
:BAD_REQUEST: Vigane päring.
:INELIGIBLE_VOTER: Valijal pole õigust hääletada.
:INTERNAL_SERVER_ERROR: Viga serveri sisemises töös.
:UNAUTHENTICATED: Autentimata päring.
:VOTER_TOO_YOUNG: Valija on liiga noor.
:VOTING_END: Hääletusperiood on lõppenud.


Allkirjastatud hääle saatmine talletamiseks
-------------------------------------------

Allkirjastatud hääle saatmine talletamiseks tähendab valijarakenduse suhtlemist
hääletamisteenusega (SNI ``voting.ivxv.invalid``).

Valijarakendus teeb päringu ``RPC.Vote`` allkirjastatud hääle talletamiseks
saatmiseks.

:params.AuthMethod: Toetatud valikud on meetodid ``tls`` ja ``ticket``.
:params.Choices: Valija ringkonnakuuluvuse identifikaator ``VoterDistrict`` mis
                 kehtis valikute nimekirja hankimise ajal. Parameetri korrektne
                 kasutamine võimaldab kogumisteenusel valijat hoiatada kui tema
                 ringkonnakuuluvus on võrreldes hääletamise algushetkega
                 muutunud.
:params.OS: Operatsioonisüsteem, millel valijarakendust kasutatakse.
:params.Type: Allkirjastatud hääle vorming. Hetkel on ainus toetatud väärtus
              ``bdoc``.
:params.Vote: BASE64-kodeeritud hääl ``SignedVote`` eelpoolmääratud vormingus
              (:ref:`signed-vote`).

Päring ``RPC.Vote`` ID-kaardiga autentimise korral.

.. literalinclude:: ../../common/examples/id.rpc.vote.query.json
   :language: json
   :linenos:

Päring ``RPC.Vote`` Mobiil-ID'ga autentimise korral.

.. literalinclude:: ../../common/examples/mid.rpc.vote.query.json
   :language: json
   :linenos:


Hääletamisteenuse vastus päringule ``RPC.Vote``.

:result.Qualification.ocsp:
:result.Qualification.tspreg:
    Kogumisteenuse poolt hangitud täiendavad tõendid valijarakenduse poolt
    loodud hääle ``SignedVote`` (:ref:`signed-vote`) kvalifitseerimiseks ning
    korrektseks registreerimiseks. Vastuse koosseis sõltub kogumisteenuse
    konkreetsest seadistusest, antud juhul kasutatakse standardset OCSP
    protokolli valija allkirjasertifikaadi kehtivuse kontrolliks ning PKIX
    ajatempliprotokolli põhist registreerimisteenust nii hääle andmise aja
    fikseerimiseks kui elektroonilise hääle registreerimiseks välises sõltumatus
    teenuses.  Valijarakendusele kontrollimiseks edastatakse nii OCSP vastus kui
    PKIX vormingus ajatempel koos registreerimisteenusele vajalike täiendustega.
:result.TestVote: Kui hääl esitati enne hääletamise algust ning läks arvesse
                  proovihäälena, siis ``true``, vastasel juhul seda välja
                  vastuses ei ole. Valijarakendus kuvab valijale proovihääle
                  korral sellekohase hoiatuse.
:result.VoteID: Hääle identifikaator talletusteenuses, mille alusel on
                kontrollrakendusel võimalik häält hilisemaks analüüsiks välja
                nõuda.

.. literalinclude:: ../../common/examples/id.rpc.vote.response.json
   :language: json
   :linenos:

Võimalikud veateated päringu ``RPC.Vote`` korral.

:BAD_CERTIFICATE: Viga valija isikutuvastus- või allkirjastamissertifikaadiga.
:BAD_REQUEST: Vigane päring.
:IDENTITY_MISMATCH: Isikutuvastus- ning allkirjastamissertifikaadi isikukoodid
                    ei kattu.
:INELIGIBLE_VOTER: Valijal pole õigust hääletada.
:INTERNAL_SERVER_ERROR: Viga serveri sisemises töös.
:OUTDATED_CHOICES: Valija ringkonnakuuluvus on nimekirja hankimisest muutunud.
:UNAUTHENTICATED: Autentimata päring.
:VOTER_TOO_YOUNG: Valija on liiga noor.
:VOTING_END: Hääletusperiood on lõppenud.


Hääletamine Mobiil-ID'ga
------------------------

Mobiil-ID kasutamine allkirjastamis- ning autentimisvahendina tingib Mobiil-ID
teenusega liidestuva abiteenuse (SNI ``mid.ivxv.invalid``) kasutamise
autentimistõendi hankimiseks enne valikute nimekirja hankimist ning hääle
allkirjastamiseks enne talletamist.


Autentimistõendi hankimine
**************************

Valijarakendus teeb päringu ``RPC.Authenticate`` Mobiil-ID autentimise
algatamiseks.

:params.OS: Operatsioonisüsteem, millel valijarakendust kasutatakse.
:params.IDCode: Mobiil-ID kasutaja isikukood.
:params.PhoneNo: Mobiil-ID kasutaja telefoninumber.

.. literalinclude:: ../../common/examples/mid.rpc.authenticate.query.json
   :language: json
   :linenos:

:result.Challenge: Räsi, millest arvutada Mobiil-ID kontrollkood valijarakenduses
                   kuvamiseks
:result.SessionCode: Mobiil-ID seansiidentifikaator edasiste poll-päringute
                     jaoks

.. literalinclude:: ../../common/examples/mid.rpc.authenticate.response.json
   :language: json
   :linenos:

Võimalikud veateated päringu ``RPC.Authenticate`` korral.

:BAD_REQUEST: Vigane päring.
:INTERNAL_SERVER_ERROR: Viga serveri sisemises töös.
:VOTING_END: Hääletusperiood on lõppenud.

Valijarakendus teeb päringu ``RPC.AuthenticateStatus`` autentimisprotsessi oleku
hindamiseks.

:params.OS: Operatsioonisüsteem, millel valijarakendust kasutatakse.
:params.SessionCode: Autentimisseansi identifikaator

.. literalinclude:: ../../common/examples/mid.rpc.authenticatestatus.query.json
   :language: json
   :linenos:


:result.AuthToken: Autentimistõend teistele IVXV teenustele esitamiseks või
                   ``null``, kui päringu töötlemine alles käib.
:result.GivenName: Eduka autentimise korral valija eesnimi
:result.PersonalCode: Eduka autentimise korral valija isikukood
:result.Status: Päringu staatus - ``POLL`` viitab vajadusele päringut korrata, ``OK``
                viitab edukale autentimisele. Vastuse muud väljad sisaldavad
                infot vaid siis kui väärtus on ``OK``.
:result.Surname: Eduka autentimise korral valija perekonnanimi


.. literalinclude:: ../../common/examples/mid.rpc.authenticatestatus.response.json
   :language: json
   :linenos:

.. literalinclude:: ../../common/examples/mid.rpc.authenticatestatus2.response.json
   :language: json
   :linenos:

Võimalikud veateated päringu ``RPC.AuthenticateStatus`` korral.

:BAD_REQUEST: Vigane päring.
:INTERNAL_SERVER_ERROR: Viga serveri sisemises töös.
:MID_BAD_CERTIFICATE: Viga valija Mobiil-ID isikutuvastussertifikaadiga.
:MID_NOT_USER: Telefoninumber ei kuulu Mobiil-ID kliendile.
:MID_OPERATOR: Probleem valija mobiiltelefoni SIM kaardiga, mille lahendamiseks
               tuleb pöörduda mobiilioperaatori poole.
:MID_ABSENT: Valija mobiiltelefon ei ole kättesaadav.
:MID_CANCELED: Valija katkestas Mobiil-ID seansi.
:MID_EXPIRED: Mobiil-ID seanss on aegunud.
:MID_GENERAL: Viga Mobiil-ID teenuse töös.
:VOTING_END: Hääletusperiood on lõppenud.


Hääle allkirjastamine
*********************

Valijarakendus teeb päringu ``RPC.GetCertificate`` allkirjastamissertifikaadi
hankimiseks.

:params.AuthMethod: Toetatud ainult autentimismeetod ``ticket``.
:params.AuthToken: Mobiil-ID autentimistõend.
:params.OS: Operatsioonisüsteem, millel valijarakendust kasutatakse.
:params.PhoneNo: Hääle allkirjastaja telefoninumber

.. literalinclude:: ../../common/examples/mid.rpc.getcertificate.query.json
   :language: json
   :linenos:


:result.Certificate: Allkirjastamissertifikaat X509-vormingus

.. literalinclude:: ../../common/examples/mid.rpc.getcertificate.response.json
   :language: json
   :linenos:

Võimalikud veateated päringu ``RPC.GetCertificate`` korral.

:BAD_REQUEST: Vigane päring.
:INTERNAL_SERVER_ERROR: Viga serveri sisemises töös.
:MID_BAD_CERTIFICATE: Viga valija Mobiil-ID allkirjastamissertifikaadiga.
:MID_GENERAL: Viga Mobiil-ID teenuse töös.
:MID_NOT_USER: Telefoninumber ei kuulu Mobiil-ID kliendile.
:VOTING_END: Hääletusperiood on lõppenud.


Valijarakendus teeb päringu ``RPC.Sign`` hääle allkirjastamise algatamiseks.
Mobiil-ID kontrollkoodi arvutab valijarakendus andmevälja ``Hash`` väärtusest.

:params.AuthMethod: Toetatud ainult autentimismeetod ``ticket``.
:params.AuthToken: Mobiil-ID autentimistõend.
:params.Hash: BASE64-kodeeritud elektroonilise hääle räsi
:params.HashType: Räsifunktsiooni nimi Mobiil-ID teenusele edastamiseks, kas
                  ``SHA256``, ``SHA384`` või  ``SHA512``
:params.OS: Operatsioonisüsteem, millel valijarakendust kasutatakse.
:params.PhoneNo: Hääle allkirjastaja telefoninumber

.. literalinclude:: ../../common/examples/mid.rpc.sign.query.json
   :language: json
   :linenos:

:result.SessionCode: Mobiil-ID seansiidentifikaator edasiste poll-päringute
                     jaoks.

.. literalinclude:: ../../common/examples/mid.rpc.sign.response.json
   :language: json
   :linenos:

Võimalikud veateated päringu ``RPC.Sign`` korral.

:BAD_REQUEST: Vigane päring.
:INTERNAL_SERVER_ERROR: Viga serveri sisemises töös.
:VOTING_END: Hääletusperiood on lõppenud.


Valijarakendus teeb päringu ``RPC.SignStatus`` allkirjastamisprotsessi seisundi
hindamiseks.

:params.OS: Operatsioonisüsteem, millel valijarakendust kasutatakse.
:params.SessionCode: Mobiil-ID seansiidentifikaator

.. literalinclude:: ../../common/examples/mid.rpc.signstatus.query.json
   :language: json
   :linenos:

:result.Signature: Juhul kui vastuse ``Status`` väli on ``OK``, BASE-64 kodeeritud
                   PKCS1-vormingus signatuur, vastasel juhul ``null``.
:result.Algorithm: Juhul kui vastuse ``Status`` väli on ``OK``, Mobiil-ID teenuse
                   poolt tagastatud signatuuri algoritm. Võimalikud väärtused on
                   ``SHA256WithECEncryption``, ``SHA256WithRSAEncryption``,
                   ``SHA384WithECEncryption``, ``SHA384WithRSAEncryption``,
                   ``SHA512WithECEncryption`` ja ``SHA512WithRSAEncryption``.
:result.Status: Päringu staatus - ``POLL`` viitab vajadusele päringut korrata, ``OK``
                viitab edukale allkirjastamisele. Vastuse muud väljad sisaldavad
                infot vaid siis kui väärtus on ``OK``.

.. literalinclude:: ../../common/examples/mid.rpc.signstatus.response.json
   :language: json
   :linenos:

.. literalinclude:: ../../common/examples/mid.rpc.signstatus2.response.json
   :language: json
   :linenos:

Võimalikud veateated päringu ``RPC.SignStatus`` korral.

:BAD_REQUEST: Vigane päring.
:INTERNAL_SERVER_ERROR: Viga serveri sisemises töös.
:MID_ABSENT: Valija mobiiltelefon ei ole kättesaadav.
:MID_BAD_CERTIFICATE: Viga valija Mobiil-ID allkirjastamissertifikaadiga.
:MID_OPERATOR: Probleem valija mobiiltelefoni SIM kaardiga, mille lahendamiseks
               tuleb pöörduda mobiilioperaatori poole.
:MID_CANCELED: Valija katkestas Mobiil-ID seansi.
:MID_EXPIRED: Mobiil-ID seanss on aegunud.
:MID_GENERAL: Viga Mobiil-ID teenuse töös.
:VOTING_END: Hääletusperiood on lõppenud.


Hääle kontrollimine
-------------------

Kontrollrakendus teeb päringu ``RPC.Verify`` allkirjastatud hääle ning häält
kvalifitseerivate tõendite allalaadimiseks kogumisteenusest.

:params.OS: Operatsioonisüsteem, millel kontrollrakendust kasutatakse.
:params.VoteID: QR-koodi vahendusel valijarakendusest saadud hääle
                identifikaator talletusteenuses.

.. literalinclude:: ../../common/examples/ver.rpc.verify.query.json
   :language: json
   :linenos:

:result.Qualification.ocsp:
:result.Qualification.tspreg:
    Vaata peatükki hääle verifitseerimisest


:result.Type: Allkirjastatud hääle vorming. Hetkel on ainus toetatud väärtus
              ``bdoc``.
:result.Vote: BASE64-kodeeritud hääl ``SignedVote`` eelpoolmääratud vormingus
              (:ref:`signed-vote`).

.. literalinclude:: ../../common/examples/ver.rpc.verify.response.json
   :language: json
   :linenos:

Võimalikud veateated päringu ``RPC.Verify`` korral.

:BAD_REQUEST: Vigane päring.
:INTERNAL_SERVER_ERROR: Viga serveri sisemises töös.
:VOTING_END: Hääletusperiood on lõppenud.
