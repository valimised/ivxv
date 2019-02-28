.. Elektroonilise hääletamise infosüsteem IVXV

.. None

.. SEE FAIL ON GENEREERITUD AUTOMAATTESTI POOLT:
.. * steps/test_util/report.py

..  Feature: Kogumisteenuse käivitamine kõrgkäideldava konfiguratsiooniga
             features/collector-ha-setup.feature:6
   Scenario: Valijate nimekirjade rakendamine
             features/collector-ha-setup.feature:493
       Step: Nimekirjade haldur eemaldab rakendamata valijate nimekirja haldusteenusest
             features/collector-ha-setup.feature:528

.. code-block:: text

   ivxv-admin@admin $ ivxv-cmd-remove voters
   INFO: Removing voters list "TRUMM,KAPTEN,33333333333 2019-02-11T11:09:38Z"