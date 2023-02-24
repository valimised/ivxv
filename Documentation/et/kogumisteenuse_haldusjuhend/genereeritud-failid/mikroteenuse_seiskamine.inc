.. Elektroonilise hääletamise infosüsteem IVXV

.. SEE FAIL ON GENEREERITUD AUTOMAATTESTI POOLT:
.. * steps/test_util/report.py

..  Feature: Kogumisteenuse käivitamine kõrgkäideldava konfiguratsiooniga
             features/end-to-end-10-collector-ha-setup.feature:7
   Scenario: Mikroteenuse seiskamine ja käivitamine
             features/end-to-end-10-collector-ha-setup.feature:512
       Step: Haldur seiskab teenuse "voting@voting2.ivxv.ee"
             features/end-to-end-10-collector-ha-setup.feature:515

.. code-block:: text

   ivxv-admin@admin $ ivxv-service stop voting@voting2.ivxv.ee
   INFO: Stopping service voting@voting2.ivxv.ee
   SERVICE voting@voting2.ivxv.ee: Stopping service
   SERVICE voting@voting2.ivxv.ee: Service stopped successfully
   INFO: Service voting@voting2.ivxv.ee stopped