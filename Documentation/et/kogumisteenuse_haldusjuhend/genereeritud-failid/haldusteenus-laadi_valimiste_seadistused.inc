.. Elektroonilise hääletamise infosüsteem IVXV

.. SEE FAIL ON GENEREERITUD AUTOMAATTESTI POOLT:
.. * steps/test_util/report.py

..  Feature: Kogumisteenuse käivitamine kõrgkäideldava konfiguratsiooniga
             features/end-to-end-10-collector-ha-setup.feature:7
   Scenario: Lugemisperioodi alustamine
             features/end-to-end-10-collector-ha-setup.feature:874
       Step: Haldur määrab kogumisteenuse uueks perioodiks "lõpetatud"
             features/end-to-end-10-collector-ha-setup.feature:877

.. code-block:: text

   ivxv-admin@admin $ ivxv-cmd-load election /output/voting/HA-SETUP/config/HA-SETUP.election.asice
   INFO: Config file is signed by: NÕID,VÄIKE,33333333333 2021-10-02T19:48:37Z
   INFO: User NÕID,VÄIKE,33333333333 with role 'election-conf-manager' is authorized to execute 'election' commands
   INFO: Using signature 'NÕID,VÄIKE,33333333333 2021-10-02T19:48:37Z' as config file version
   INFO: Config file version is 'NÕID,VÄIKE,33333333333 2021-10-02T19:48:37Z'
   INFO: Loading command 'elections configuration' from file '/output/voting/HA-SETUP/config/HA-SETUP.election.asice'
   command_file:INFO: Loading command file '/output/voting/HA-SETUP/config/HA-SETUP.election.asice' (elections configuration)
   command_file:INFO: Validating elections configuration
   command_file:INFO: Files in elections configuration package are valid
   INFO: Elections configuration file loaded successfully
   command_file:INFO: Loading command file '/output/voting/HA-SETUP/config/HA-SETUP.election.asice' (elections configuration)
   INFO: Elections configuration file is registered in management service