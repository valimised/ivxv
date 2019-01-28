.. Elektroonilise hääletamise infosüsteem IVXV

.. None

.. SEE FAIL ON GENEREERITUD AUTOMAATTESTI POOLT:
.. * steps/test_util/report.py

..  Feature: Kogumisteenuse käivitamine kõrgkäideldava konfiguratsiooniga
             features/collector-ha-setup.feature:6
   Scenario: Valijate nimekirjade rakendamine
             features/collector-ha-setup.feature:489
       Step: Nimekirjade haldur eemaldab rakendamata valijate nimekirja haldusteenusest
             features/collector-ha-setup.feature:524

.. code-block:: text

   ivxv-admin@admin $ ivxv-cmd-remove voters
   INFO: Removing voters list "TRUMM,KAPTEN,33333333333 2019-01-25T14:37:53Z"