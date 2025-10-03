Krahhi toimumise indikaatorid
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Haldusveebiliides.
Vali **Teenused** vasakult menüüst. Kui teenus on krahhinud,
siis tema **Seisund** on **Tõrge**.

Halduskäsurealiides.
Logi sisse **ivxv-admin** kasutajana ja päri teenuste statistikat::

   $ sudo su - ivxv-admin
   $ ivxv-status

Kui teenus on krahhinud, siis tema **State** on **FAILURE**.


Krahhist taastumise eeldused
^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Krahhist taastumiseks piisab kui teenuste CA TLS privaatvõti ja sertifikaat on
alles, mida saab kasutada teenuse varuvõtmete ja -sertifikaatide
genereerimiseks. Samas, saab ka enne valimisi teenuse varuvõtmed ja
-sertifikaadid ette genereerida.

Kui kõikide teenuste CA on sama, siis tasub valija- ja kontrollrakenduste
seadistusse panna hoopis CA sertifikaat. Juhul kui CA on teenustel erinev,
siis peab panema kõik teenuste TLS sertifikaadid rakenduste seadistusse
vahetult enne valimisi.


Krahhi võimalik mõju hääletamistulemusele
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Sõltuvalt teenusest võivad tagajärjed olla järgmised:

   * kui krahhivad nimekirja-, Mobiil-ID, vahendus-, seansi-, Smart-ID,
     hääletamis-, Web-eID teenused, siis valijal ei õnnestu hääletada

   * kui krahhivad järjekorra- ja x-tee teenused, siis VIS ei saa pärida
     andmeid x-tee kaudu

   * kui krahhib verifitseerimisteenus, siis valijal ei õnnestu oma häält
     verifitseerida


Krahhist taastumise sammud
^^^^^^^^^^^^^^^^^^^^^^^^^^

.. seealso::

   * Talletusteenust antud juhendis ei käsitleta, selleks
     on eraldi juhend :ref:`recovery-storageservice`

Krahhi saab liigitada järgmiselt:

   #. Teenus on terve, ning vajab ainult taaskäivitamist, v.a x-tee teenus

   #. Teenuse andmed (TLS privaatvõti, sertifikaat jne) on rikutud, v.a
      vahendus- ja x-tee teenused

   #. Teenuse lähtekoodis on viga, v.a x-tee teenus

   #. X-tee teenuse taastumine, sh. krahh ja lähtekoodi viga

   #. Teenuse masin on krahhinud, v.a vahendus- ja x-tee teenused

   #. Vahendusteenuse masin on krahhinud

Teenus on terve, ning vajab ainult taaskäivitamist, v.a x-tee teenus
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

   #. SSH logi sisse haldus masinasse **ivxv-admin** kasutajana::

         $ sudo su - ivxv-admin

   #. Taaskäivita teenus, näiteks nimekirjateenust::

         $ ivxv-service restart choices@choices1.ivxv.ee

Teenuse andmed (TLS privaatvõti, sertifikaat jne) on rikutud, v.a vahendus- ja x-tee teenused
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

   #. Vaheta välja krahhinud teenuse **id** tehnilises seadistuses
      varuteenuse **id** vastu, mille TLS privaatvõti ja sertifikaat on
      genereeritud

   #. Allkirjasta tehniline seadistus

   #. Laadi seadistus läbi haldusveebiliidese **Seadistused**

   #. SSH logi sisse haldusmasinasse **ivxv-admin** kasutajana::

         $ sudo su - ivxv-admin

   #. Laadi TLS privaatvõti ja sertifikaat, näiteks nimekirjateenus::

         $ ivxv-secret-load --service=choices@choices4.ivxv.ee tls-key /opt/ivxv-install/service_definition/choices@choices4.ivxv.ee.key
         $ ivxv-secret-load --service=choices@choices4.ivxv.ee tls-cert /opt/ivxv-install/service_definition/choices@choices4.ivxv.ee.pem

   #. Kui tegu on nimekirja-, Mobiil-ID, Smart-ID, hääletamis-, Web-eID või
      verifitseerimisteenusega siis tuleb üles laadida ka küpsiste
      allkirjatamiseks privaatvõti::

         $ ivxv-secret-load mid-token-key /opt/ivxv-install/service_definition/ticket.key

   #. Kui tegu on hääletamisteenusega siis tuleb üles laadida TSP vastuse
      räsi allkirjastamiseks privaatvõti::

         $ ivxv-secret-load tsp-regkey /opt/ivxv-install/service_definition/tspreg.key

   #. Käivita teenused::

         $ ivxv-config-apply --type=election

Teenuse lähtekoodis on viga, v.a x-tee teenus
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

   #. Kopeeri uus *.deb* pakk haldusmasinasse

   #. SSH logi sisse haldusmasinasse **sudo** kasutajana::

         $ sudo su

   #. Kopeeri *.deb* pakk */etc/ivxv/debs/* kausta

   #. Logi **sudo** kasutajalt välja::

         $ exit

   #. SSH logi sisse haldusmasinasse **ivxv-admin** kasutajana::

         $ sudo su - ivxv-admin

   #. Uuenda teenuse *.deb* pakki kõikides masinates, näiteks nimekirjateenus::

         $ ivxv-update-packages --service=choices --package=/etc/ivxv/debs/ivxv-choices_1.9.12~dev_amd64.deb

   #. Taaskäivita teenus igal hostil, näiteks taaskäivitame nimekirjateenuse
      kolmeõlgse seadistuse korral::

         $ ivxv-service restart choices@choices1.ivxv.ee
         $ ivxv-service restart choices@choices2.ivxv.ee
         $ ivxv-service restart choices@choices3.ivxv.ee

X-tee teenuse taastumine, sh. krahh ja lähtekoodi viga
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

   #. Esmalt tasub proovida taaskäivitada systemd x-tee teenust::

         $ sudo systemctl restart xroad-service

   #. Ebaõnnestumise korral tuleb tarnida uus *.deb* pakk, ning
      vana pakk eemaldada::

         $ sudo apt purge xroad-service

   #. Kui uus *.deb* pakk on masinasse kopeeritud, tuleb see paigaldada::

         $ sudo apt install ./xroad-service_*.deb

   #. Kasuta x-tee teenuse juhendit, et paigaldada TLS sertifikaadid õigesse
      kaustadesse

   #. Käivita systemd x-tee teenus::

         $ sudo systemctl start xroad-service

Teenuse masin on krahhinud, v.a vahendus- ja x-tee teenused
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

   #. Paigalda värske Ubuntu masin

      .. important::
         Palun täpsusta üle, millist Ubuntu versiooni parasjagu
         kasutatakse

   #. Paigalda teenuse masinasse baastarkvara pakid APT
      pakihoidlast::

         $ sudo apt-get install acl locales unzip zstd --yes

   #. Nüüd tuleb haldus-, koguja-, varundus-, logi-, vahendus-, logi
      monitooringu masinate **ip** ja DNS **nimi** lisada uue teenuse masina
      */etc/hosts* faili

   #. Seejärel tuleb haldus-, koguja-, varundus-, logi-, vahendus masinates
      kustutada krahhinud teenuse masina **ip** ja DNS **nimi** ning lisada
      uus

      .. attention::

         See on kõige olulisem samm, kuna siin peab valima sellise DNS
         **nime**, mille tulevikus lisame tehnilistesse seadistustesse.

         Reegel on lihtne, et iga uue teenuse masina DNS **nimi** peab olema
         unikaalne ja ei tohi olla sama, mis krahhinud teenuste masinatele
         olid varem omistatud. Haldusteenuse andmebaasis on registreeritud
         kõigi süsteemis leiduvate masinate DNS **nimed**, isegi krahhinud
         isendite omad. Haldusteenuse eripära seisneb selles, et see ei
         paigalda tarkvara masinatele, mille DNS **nimi** on juba
         haldusteenuse andmebaasis registreeritud.

   #. SSH logi sisse haldusmasinasse **ivxv-admin** kasutajana::

         $ sudo su - ivxv-admin

   #. Kuna haldusteenus paigaldab vajaliku taristu kõikidele süsteemi masinatele
      automaatselt, siis on vaja vältida SSH ühenduse loomisel *prompt* sõnumi
      *Are you sure you want to continue connecting (yes/no)? yes* kinnitamist.

      Kustuta vana teenusmasina SSH avalik võti */etc/ssh/ssh_known_hosts*
      failist, ning lisa uus, näiteks, kui uue teenuse masina DNS **nimi** on
      *ivxv4*::

         $ ssh-keyscan -t ecdsa ivxv4 | grep -v ^# >> /etc/ssh/ssh_known_hosts

   #. SSH logi sisse uue teenuse masinasse **root** kasutajana::

         $ sudo su

   #. Kuna haldusteenus edastab teenuse logid logimonitori, lisaks
      *syslog* *RELP* protokollile, ka üle SSH protokolli, siis on vaja samuti
      vältida SSH ühenduse loomisel *prompt* sõnumi kinnitamist, näiteks,
      monitooringu masina *logmonitor* DNS **address** korral::

         $ ssh-keyscan -t ecdsa logmonitor | grep -v ^# >> /etc/ssh/ssh_known_hosts

   #. Kuna haldusteenus peab paigaldama tarkvara APT pakihoidlast, mis
      on võimatu juurkasutaja õiguseid omandamata, pead kopeerima haldusmasina
      SSH avaliku võtme, mis asub
      */var/lib/ivxv/user/ivxv-admin/.ssh/id_ed25519.pub* failis, ning lisama
      selle uue teenuse masina */root/.ssh/authorized_keys* faili

   #. Nüüd tuleb tehnilistes seadistustes välja vahetada teenuste **id** ja
      **nimed** (**address**), näiteks:

      .. code-block:: yaml

         - id: zone1
            services:
               mid:
                  - id:         mid@mid1.ivxv.ee
                   address:     ivxv1:4441
               choices:
                  - id:         choices@choices1.ivxv.ee
                   address:     ivxv1:4442
               voting:
                  - id:         voting@voting1.ivxv.ee
                   address:     ivxv1:4443
               verification:
                  - id:         verification@verification1.ivxv.ee
                   address:     ivxv1:4444
               smartid:
                  - id:         smartid@smartid1.ivxv.ee
                   address:     ivxv1:4445
               votesorder:
                  - id:         votesorder@votesorder1.ivxv.ee
                   address:     ivxv1:4446
               webeid:
                  - id:         webeid@webeid1.ivxv.ee
                   address:     ivxv1:4447
                   origin:      https://ivxv1.test.ivxv.ee:443
               sessionstatus:
                  - id:         sessionstatus@sessionstatus1.ivxv.ee
                   address:     ivxv1:4448

      uute **id** ja **address** vastu:

      .. code-block:: yaml

         - id: zone1
            services:
               mid:
                  - id:         mid@mid4.ivxv.ee
                   address:     ivxv4:4441
               choices:
                  - id:         choices@choices4.ivxv.ee
                   address:     ivxv4:4442
               voting:
                  - id:         voting@voting4.ivxv.ee
                   address:     ivxv4:4443
               verification:
                  - id:         verification@verification4.ivxv.ee
                   address:     ivxv4:4444
               smartid:
                  - id:         smartid@smartid4.ivxv.ee
                   address:     ivxv4:4445
               votesorder:
                  - id:         votesorder@votesorder4.ivxv.ee
                   address:     ivxv4:4446
               webeid:
                  - id:         webeid@webeid4.ivxv.ee
                   address:     ivxv4:4447
                   origin:      https://ivxv4.test.ivxv.ee:443
               sessionstatus:
                 - id:          sessionstatus@sessionstatus4.ivxv.ee
                   address:     ivxv4:4448

      .. attention::
         Pane tähele, et on muutunud nii **id** kui ka **address**.
         Vahendus- ja talletusteenuste **id** ja **address** jäävad samaks
         **siis ja ainult siis**, kui need teenused ei olnud krahhinud masina
         koosseisus, vastasel juhul tuleb sarnaselt muuta ka nende teenuste
         **id** ja **address**.

   #. Allkirjasta uus tehniline seadistus. Laadi see läbi
      haldusveebiliidese, vasakult valides **Seadusted**.

   #. SSH logi sisse haldusmasinasse **ivxv-admin** kasutajana::

         $ sudo su - ivxv-admin

   #. Laadi kõik teenuste TLS privaatvõtmed ja sertifikaadid, näiteks::

         $ ivxv-secret-load --service=choices@choices4.ivxv.ee tls-key /opt/ivxv-install/service_definition/choices@choices4.ivxv.ee.key
         $ ivxv-secret-load --service=voting@voting4.ivxv.ee tls-key /opt/ivxv-install/service_definition/voting@voting4.ivxv.ee.key
         $ ivxv-secret-load --service=mid@mid4.ivxv.ee tls-key /opt/ivxv-install/service_definition/mid@mid4.ivxv.ee.key
         $ ivxv-secret-load --service=smartid@smartid4.ivxv.ee tls-key /opt/ivxv-install/service_definition/smartid@smartid4.ivxv.ee.key
         $ ivxv-secret-load --service=sessionstatus@sessionstatus4.ivxv.ee tls-key /opt/ivxv-install/service_definition/sessionstatus@sessionstatus4.ivxv.ee.key
         $ ivxv-secret-load --service=verification@verification4.ivxv.ee tls-key /opt/ivxv-install/service_definition/verification@verification4.ivxv.ee.key
         $ ivxv-secret-load --service=votesorder@votesorder4.ivxv.ee tls-key /opt/ivxv-install/service_definition/votesorder@votesorder4.ivxv.ee.key
         $ ivxv-secret-load --service=webeid@webeid4.ivxv.ee tls-key /opt/ivxv-install/service_definition/webeid@webeid4.ivxv.ee.key
         $ ivxv-secret-load --service=choices@choices4.ivxv.ee tls-cert /opt/ivxv-install/service_definition/choices@choices4.ivxv.ee.pem
         $ ivxv-secret-load --service=voting@voting4.ivxv.ee tls-cert /opt/ivxv-install/service_definition/voting@voting4.ivxv.ee.pem
         $ ivxv-secret-load --service=mid@mid4.ivxv.ee tls-cert /opt/ivxv-install/service_definition/mid@mid4.ivxv.ee.pem
         $ ivxv-secret-load --service=smartid@smartid4.ivxv.ee tls-cert /opt/ivxv-install/service_definition/smartid@smartid4.ivxv.ee.pem
         $ ivxv-secret-load --service=sessionstatus@sessionstatus4.ivxv.ee tls-cert /opt/ivxv-install/service_definition/sessionstatus@sessionstatus4.ivxv.ee.pem
         $ ivxv-secret-load --service=verification@verification4.ivxv.ee tls-cert /opt/ivxv-install/service_definition/verification@verification4.ivxv.ee.pem
         $ ivxv-secret-load --service=votesorder@votesorder4.ivxv.ee tls-cert /opt/ivxv-install/service_definition/votesorder@votesorder4.ivxv.ee.pem
         $ ivxv-secret-load --service=webeid@webeid4.ivxv.ee tls-cert /opt/ivxv-install/service_definition/webeid@webeid4.ivxv.ee.pem

   #. Lisa küpsiste signeerimisvõti::

      $ ivxv-secret-load mid-token-key /opt/ivxv-install/service_definition/ticket.key

   #. Lisa hääletamisteenuse TSP räsi signeerimisvõti::

      $ ivxv-secret-load tsp-regkey /opt/ivxv-install/service_definition/tspreg.key

   #. Käivita teenused::

      $ ivxv-config-apply --type=election

Vahendusteenuse masin on krahhinud
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

.. seealso::

   * Kasuta käesoleva juhendi
     **Teenuste masin on krahhinud, v.a vahendus- ja x-tee teenused**
     peatükki, v.a punktid 6, 10-12.

.. note::

   Punktis 8 tuleb ainult asendada vana vahendusteenuse **id** ja **address**:

   .. code-block:: yaml

     - id: zone1
       services:
         proxy:
           - id:          proxy@proxy1.ivxv.ee
             address:     proxy1:443

   uute vastu:

   .. code-block:: yaml

     - id: zone1
       services:
         proxy:
           - id:          proxy@proxy4.ivxv.ee
             address:     proxy4:443
