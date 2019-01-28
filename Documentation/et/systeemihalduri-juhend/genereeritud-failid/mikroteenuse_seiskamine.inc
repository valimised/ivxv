.. Elektroonilise hääletamise infosüsteem IVXV

.. None

.. SEE FAIL ON GENEREERITUD AUTOMAATTESTI POOLT:
.. * steps/test_util/report.py

..  Feature: Kogumisteenuse käivitamine kõrgkäideldava konfiguratsiooniga
             features/collector-ha-setup.feature:6
   Scenario: Mikroteenuse seiskamine ja käivitamine
             features/collector-ha-setup.feature:427
       Step: Haldur seiskab teenuse "voting@voting2.ivxv.ee"
             features/collector-ha-setup.feature:430

.. code-block:: text

   ivxv-admin@admin $ ivxv-service stop voting@voting2.ivxv.ee
   INFO: Stopping service voting@voting2.ivxv.ee
   SERVICE voting@voting2.ivxv.ee: Stopping service
   SERVICE voting@voting2.ivxv.ee: Service stopped successfully
   INFO: Service voting@voting2.ivxv.ee stopped