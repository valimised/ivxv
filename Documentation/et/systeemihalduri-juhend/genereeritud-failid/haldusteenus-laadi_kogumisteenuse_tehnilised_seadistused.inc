.. Elektroonilise hääletamise infosüsteem IVXV

.. None

.. SEE FAIL ON GENEREERITUD AUTOMAATTESTI POOLT:
.. * steps/test_util/report.py

..  Feature: Kogumisteenuse paigaldamine kõrgkäideldava konfiguratsiooniga
             features/collector-ha-setup.feature:6
   Scenario: Kogumisteenuse seadistuste rakendamine
             features/collector-ha-setup.feature:492
       Step: Haldur laadib haldusteenusesse kogumisteenuse tehnilised seadistused
             features/collector-ha-setup.feature:502

.. code-block:: text

   ivxv-admin@admin $ ivxv-cmd-load technical /output/voting/HA-SETUP/config/technical.bdoc
   INFO: Config file is signed by: ROPKA,KIVIVALVUR,32608320001 2017-08-24T15:39:05Z
   INFO: User ROPKA,KIVIVALVUR,32608320001 with role "admin" is authorized to execute "technical" commands
   INFO: Using signature "ROPKA,KIVIVALVUR,32608320001 2017-08-24T15:39:05Z" as config file version
   INFO: Loading command technical from file /output/voting/HA-SETUP/config/technical.bdoc
   command_file:INFO: Loading command file "/output/voting/HA-SETUP/config/technical.bdoc" (collectors technical configuration)
   command_file:INFO: Command file loaded
   command_file:INFO: Command file successfully extracted
   command_file:INFO: Reading files from command file
   command_file:INFO: Validating collectors technical configuration
   command_file:INFO: Files in collectors technical configuration package are valid
   INFO: Technical config file loaded successfully