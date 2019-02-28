.. Elektroonilise hääletamise infosüsteem IVXV

.. None

.. SEE FAIL ON GENEREERITUD AUTOMAATTESTI POOLT:
.. * steps/test_util/report.py

..  Feature: Kogumisteenuse käivitamine kõrgkäideldava konfiguratsiooniga
             features/collector-ha-setup.feature:6
   Scenario: Mikroteenuse seiskamine ja käivitamine
             features/collector-ha-setup.feature:431
       Step: Haldur tuvastab teenuse "voting@voting2.ivxv.ee" seisundi
             features/collector-ha-setup.feature:435

.. code-block:: text

   ivxv-admin@admin $ ivxv-service ping voting@voting2.ivxv.ee
   INFO: Pinging service voting@voting2.ivxv.ee
   SERVICE voting@voting2.ivxv.ee: Registering background info: Ping error: * ivxv-voting@voting@voting2.ivxv.ee.service - IVXV voting service
   SERVICE voting@voting2.ivxv.ee: ERROR: Pinging service failed
   ERROR: Failed to query service voting@voting2.ivxv.ee status