
Krahhi toimumise indikaatorid
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Haldusveebiliides.
Vali **Teenused** vasakult menüüst. Kui varundusteenus on krahhinud,
siis tema **Seisund** on **Tõrge**.

Halduskäsurealiides.
Logi sisse **ivxv-admin** kasutajana ja päri teenuste statistikat::

   $ sudo su - ivxv-admin
   $ ivxv-status

Kui varundusteenus on krahhinud, siis tema **State** on **FAILURE**.


Krahhist taastumise eeldused
^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Krahhist taastumine on alati võimalik.


Krahhi võimalik mõju hääletamistulemusele
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Varundusteenuse krahh mõjutab hääletamistulemust (andmekadu), kui täiendavalt
krahhib ka talletusteenus.


Krahhist taastumise sammud
^^^^^^^^^^^^^^^^^^^^^^^^^^

.. seealso::

   * Kasuta :ref:`recovery-stateless` juhendis oleva
     **Teenuse masin on krahhinud, v.a vahendus- ja x-tee teenused**
     peatükki, v.a sammud 8, 10, 12-16.

.. note::

   Punktis 10 tuleb välja vahetada ainult varundusteenuse **id** ja **address**
   väljad:

   .. code-block:: yaml

     - id: zone0
       services:
         log:
           - id:          backup@backup1.ivxv.ee
             address:     backup1

   uute vastu:

   .. code-block:: yaml

     - id: zone0
       services:
         log:
           - id:          backup@backup2.ivxv.ee
             address:     backup2

SSH logi sisse haldusmasinasse **ivxv-admin** kasutajana::

    $ sudo su - ivxv-admin

Seadista SSH ühenduse loomiseks *prompt* sõnumi
*Are you sure you want to continue connecting (yes/no)? yes* kinnitamist
vältida, näiteks, kui varundusteenuse masina DNS **address** on *backup2*::

    $ ssh backup2
    $ echo "StrictHostKeyChecking no" >> /var/lib/ivxv/user/ivxv-admin/.ssh/config
