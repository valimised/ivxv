Krahhi toimumise indikaatorid
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Haldusveebiliides.
Vali **Teenused** vasakult menüüst. Kui logikogumisteenus on krahhinud,
siis tema **Seisund** on **Tõrge**.

Halduskäsurealiides.
Logi sisse **ivxv-admin** kasutajana ja päri teenuste statistikat::

   $ sudo su - ivxv-admin
   $ ivxv-status

Kui logikogumisteenus on krahhinud, siis tema **State** on **FAILURE**.


Krahhist taastumise eeldused
^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Krahhist taastumine on alati võimalik.


Krahhi võimalik mõju hääletamistulemusele
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Hääletamistulemusele mõju puudub, küll aga krahhinud logikogumisteenusega kaob
ka osa varundatud logidest, mida enam taastada ei ole võimalik. Samas, need
logid on ikkagi endiselt alles koguja masinates, ning ka logimonitori
masinas. Ainus koht kus avaldub logikogumisteenuse krahh on varundusteenuse
masin, kus jääbki puudu osa varundatud logidest.


Krahhist taastumise sammud
^^^^^^^^^^^^^^^^^^^^^^^^^^
.. seealso::

   * Kasuta :ref:`recovery-stateless` juhendis oleva
     **Teenuse masin on krahhinud, v.a vahendus- ja x-tee teenused**
     peatükki, v.a sammud 10, 12-16.


.. note::

   Punktis 10 tuleb välja vahetada ainult logikogumisteenuse **id** ja
   **address** väljad:

   .. code-block:: yaml

     - id: zone0
       services:
         log:
           - id:          log@log1.ivxv.ee
             address:     log1:20514

   uute vastu:

   .. code-block:: yaml

     - id: zone0
       services:
         log:
           - id:          log@log2.ivxv.ee
             address:     log2:20514
