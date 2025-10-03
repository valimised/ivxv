Krahhi toimumise indikaatorid
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Haldusveebiliides.
Vali **Teenused** vasakult menüüst. Kui talletusteenused on krahhinud,
siis tema **Seisund** on **Tõrge**.

Halduskäsurealiides.
Logi sisse **ivxv-admin** kasutajana ja päri teenuste statistikat::

   $ sudo su - ivxv-admin
   $ ivxv-status

Kui talletusteenused on krahhinud, siis tema **State** on **FAILURE**.


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

Kui krahhib talletusteenuse klaster, siis kindlasti tegu on andmekaoga.
Andmekao suurus sõltub varundusteenusest ehk kui tihti e-urni varundati.

Lisaks, juba hääletanud valijal võivad tekkida segadus, kui teist korda
hääletades, kuvatakse valijarakenduses "Teie pole veel hääletanud"
tervitustekst.


Krahhist taastumise sammud
^^^^^^^^^^^^^^^^^^^^^^^^^^

.. seealso::

   * Tee :ref:`recovery-storageservice` juhendis oleva
     **Teenuse andmed (võti, sertifikaat jne) on rikutud või masina krahh**
     peatükki protseduurid läbi iga uue koguja masina jaoks.

.. note::

   Punktis 10 tuleb välja vahetada kõik vanad **id** ja **address** kirjed
   uute vastu, lisaks tuleb muuta ka talletusteenuse **bootstrap**
   seadistust:

   .. code-block:: yaml

      bootstrap:
      - storage@storage1.ivxv.ee
      - storage@storage2.ivxv.ee
      - storage@storage3.ivxv.ee

   selliselt:

   .. code-block:: yaml

      bootstrap:
      - storage@storage4.ivxv.ee
      - storage@storage5.ivxv.ee
      - storage@storage6.ivxv.ee

Nüüd tuleb taastada koguja krahhieelne seisund, v.a kaotsi läinud hääled:

   #. SSH logi sisse haldusmasinasse **ivxv-admin** kasutajana::

         $ sudo su - ivxv-admin

   #. Laadi ringkondade nimekiri::

         $ ivxv-config-apply --type=districts

   #. Laadi valikute nimekiri::

         $ ivxv-config-apply --type=choices

   #. Laadi nii valijate alg-, kui ka kõik vahepealsed muudatusnimekirjad::

         $ ivxv-config-apply --type=voters

.. warning::
   Ära unusta, et valimiste lõpus tuleb hääli alla laadida koos varundatud
   häältega, et vähemalt mingis ulatuses tagada e-urni terviklikkus::

      $ ivxv-export-votes --consolidate
