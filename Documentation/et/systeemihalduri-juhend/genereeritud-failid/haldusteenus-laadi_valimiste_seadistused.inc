.. Elektroonilise hääletamise infosüsteem IVXV

.. None

.. SEE FAIL ON GENEREERITUD AUTOMAATTESTI POOLT:
.. * steps/test_util/report.py

..  Feature: Kogumisteenuse käivitamine kõrgkäideldava konfiguratsiooniga
             features/collector-ha-setup.feature:6
   Scenario: Lugemisperioodi alustamine
             features/collector-ha-setup.feature:578
       Step: Haldur määrab kogumisteenuse uueks perioodiks "lõpetatud"
             features/collector-ha-setup.feature:581

.. code-block:: text

   ivxv-admin@admin $ ivxv-cmd-load election /output/voting/HA-SETUP/config/HA-SETUP.election.bdoc
   INFO: Config file is signed by: TRUMM,KAPTEN,33333333333 2019-01-25T14:39:18Z
   INFO: User TRUMM,KAPTEN,33333333333 with role "election-conf-manager" is authorized to execute "election" commands
   INFO: Using signature "TRUMM,KAPTEN,33333333333 2019-01-25T14:39:18Z" as config file version
   INFO: Config file version is "TRUMM,KAPTEN,33333333333 2019-01-25T14:39:18Z"
   INFO: Loading command "elections configuration" from file /output/voting/HA-SETUP/config/HA-SETUP.election.bdoc
   command_file:INFO: Loading command file "/output/voting/HA-SETUP/config/HA-SETUP.election.bdoc" (elections configuration)
   command_file:INFO: Validating elections configuration
   command_file:INFO: Files in elections configuration package are valid
   INFO: Elections configuration file loaded successfully
   command_file:INFO: Loading command file "/output/voting/HA-SETUP/config/HA-SETUP.election.bdoc" (elections configuration)
   INFO: Elections configuration file is registered in management service