.. Elektroonilise hääletamise infosüsteem IVXV

.. None

.. SEE FAIL ON GENEREERITUD AUTOMAATTESTI POOLT:
.. * steps/test_util/report.py

..  Feature: Kogumisteenuse käivitamine kõrgkäideldava konfiguratsiooniga
             features/collector-ha-setup.feature:6
   Scenario: Lugemisperioodi alustamine
             features/collector-ha-setup.feature:585
       Step: Haldur määrab kogumisteenuse uueks perioodiks "lõpetatud"
             features/collector-ha-setup.feature:588

.. code-block:: text

   ivxv-admin@admin $ ivxv-cmd-load election /output/voting/HA-SETUP/config/HA-SETUP.election.asice
   INFO: Config file is signed by: TRUMM,KAPTEN,33333333333 2019-05-07T10:54:00Z
   INFO: User TRUMM,KAPTEN,33333333333 with role "election-conf-manager" is authorized to execute "election" commands
   INFO: Using signature "TRUMM,KAPTEN,33333333333 2019-05-07T10:54:00Z" as config file version
   INFO: Config file version is "TRUMM,KAPTEN,33333333333 2019-05-07T10:54:00Z"
   INFO: Loading command "elections configuration" from file /output/voting/HA-SETUP/config/HA-SETUP.election.asice
   command_file:INFO: Loading command file "/output/voting/HA-SETUP/config/HA-SETUP.election.asice" (elections configuration)
   command_file:INFO: Validating elections configuration
   command_file:INFO: Files in elections configuration package are valid
   INFO: Elections configuration file loaded successfully
   command_file:INFO: Loading command file "/output/voting/HA-SETUP/config/HA-SETUP.election.asice" (elections configuration)
   INFO: Elections configuration file is registered in management service