.. Elektroonilise hääletamise infosüsteem IVXV

.. None

.. SEE FAIL ON GENEREERITUD AUTOMAATTESTI POOLT:
.. * steps/test_util/report.py

..  Feature: Kogumisteenuse käivitamine kõrgkäideldava konfiguratsiooniga
             features/collector-ha-setup.feature:6
   Scenario: Mikroteenuse seiskamine ja käivitamine
             features/collector-ha-setup.feature:431
       Step: Haldur käivitab teenuse "voting@voting2.ivxv.ee"
             features/collector-ha-setup.feature:437

.. code-block:: text

   ivxv-admin@admin $ ivxv-service restart voting@voting2.ivxv.ee
   INFO: Restarting service voting@voting2.ivxv.ee
   SERVICE voting@voting2.ivxv.ee: Restarting service
   SERVICE voting@voting2.ivxv.ee: Registering service state as "CONFIGURED" in management database (last state: "FAILURE")
   SERVICE voting@voting2.ivxv.ee: Service restarted successfully
   INFO: Service voting@voting2.ivxv.ee restarted