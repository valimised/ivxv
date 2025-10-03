Krahhi toimumise indikaatorid
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Haldusveebiliides.
Vali **Teenused** vasakult menüüst. Kui talletusteenus on krahhinud,
siis tema **Seisund** on **Tõrge**.

Halduskäsurealiides.
Logi sisse **ivxv-admin** kasutajana ja päri teenuste statistikat::

   $ sudo su - ivxv-admin
   $ ivxv-status

Kui talletusteenus on krahhinud, siis tema **State** on **FAILURE**.


Krahhist taastumise eeldused
^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Krahhi taastumiseks piisab kui teenuste CA TLS privaatvõti ja sertifikaat on
alles, mida saab kasutada teenuse varuvõtmete ja -sertifikaatide
genereerimiseks. Samas, saab ka enne valimisi teenuse varuvõtmed ja
-sertifikaadid ette genereerida.

Kui kõikide teenuste CA on sama, siis tasub valija- ja kontrollrakenduste
seadistusse panna hoopis CA TLS sertifikaadi. Juhul kui CA on teenustel erinev,
siis peab panema kõik teenuste TLS sertifikaadid rakenduste seadistusse
vahetult enne valimisi.


Krahhi võimalik mõju hääletamistulemusele
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Kõik päringud mis saadetakse vastu krahhinud talletusteenuse
masinat ebaõnnestuvad. Valija ei saa autentida, ega hääletada.


Krahhist taastumise sammud
^^^^^^^^^^^^^^^^^^^^^^^^^^

Krahhi saab liigitada järgmiselt:

   #. Talletusteenus on terve, ning vajab ainult taaskäivitamist

   #. Talletusteenuse andmed (võti, sertifikaat jne) on rikutud või masina
      krahh

   #. Talletusteenuse lähtekoodis on viga

Talletusteenus on terve, ning vajab ainult taaskäivitamist
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

.. seealso::

   * Kasuta :ref:`recovery-stateless` juhendis oleva
     **Teenus on terve, ning vajab ainult taaskäivitamist, v.a x-tee teenus**
     peatükki

Talletusteenuse andmed (võti, sertifikaat jne) on rikutud või masina krahh
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

.. seealso::

   * Kasuta :ref:`recovery-stateless` juhendis oleva
     **Teenuse masin on krahhinud, v.a vahendus- ja x-tee teenused**
     peatükki


.. note::

   Punktis 10 tuleb asendada vana talletusteenuse **id** ja **address** väljad:

   .. code-block:: yaml

     - id: zone1
       services:
         storage:
           - id:          storage@storage1.ivxv.ee
             address:     ivxv1:2379
             peeraddress: ivxv1:2380

   uute vastu:

   .. code-block:: yaml

      - id: zone1
        services:
          storage:
            - id:          storage@storage4.ivxv.ee
              address:     ivxv4:2379
              peeraddress: ivxv4:2380

   Punktis 12 tuleb uuendada ka talletusteenuse TLS privaatvõti::

      $ ivxv-secret-load --service=storage@storage4.ivxv.ee tls-key /opt/ivxv-install/service_definition/storage@storage4.ivxv.ee.key

   ja sertifikaat:::

      $ ivxv-secret-load --service=storage@storage4.ivxv.ee tls-cert /opt/ivxv-install/service_definition/storage@storage4.ivxv.ee.pem

Veendu, et asendamine läks sujuvalt läbi:

   #. SSH logi uue koguja masinasse *root* kasutajana::

      $ sudo su

   #. Päri andmebaasi klastri seisu, näiteks, kui klaster koosneb
      *ivxv4*, *ivxv2*, *ivxv3* masinatest::

      $ cd /var/lib/ivxv/service/storage@storage4.ivxv.ee
      $ env ETCDCTL_API=3 etcdctl --endpoints=ivxv4:2379,ivxv2:2379,ivxv3:2379 --cert tls.pem --key tls.key --cacert ca.pem endpoint status -w table

Talletusteenuse lähtekoodis on viga
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Kui talletusteenuse lähtekoodis on viga, siis seda saab
asendada uuema versiooniga, olemasoleva andmebaasi
klastrit rikumata.

.. seealso::

   * Kasuta :ref:`recovery-stateless` juhendis oleva
     **Teenuse lähtekoodis on viga, v.a x-tee teenus**
     peatükki
