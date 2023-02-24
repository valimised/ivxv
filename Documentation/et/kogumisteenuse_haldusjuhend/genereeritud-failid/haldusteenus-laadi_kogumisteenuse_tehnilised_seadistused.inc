.. Elektroonilise hääletamise infosüsteem IVXV

.. SEE FAIL ON GENEREERITUD AUTOMAATTESTI POOLT:
.. * steps/test_util/report.py

..  Feature: Kogumisteenuse käivitamine kõrgkäideldava konfiguratsiooniga
             features/end-to-end-10-collector-ha-setup.feature:7
   Scenario: Kogumisteenuse tehniliste seadistuste rakendamine
             features/end-to-end-10-collector-ha-setup.feature:404
       Step: Haldur laadib haldusteenusesse kogumisteenuse tehnilised seadistused
             features/end-to-end-10-collector-ha-setup.feature:438

.. code-block:: text

   ivxv-admin@admin $ ivxv-cmd-load technical /output/voting/HA-SETUP/config/HA-SETUP.technical.asice
   INFO: Config file is signed by: ROPKA,KIVIVALVUR,32608320001 2021-12-28T11:38:40Z
   INFO: User ROPKA,KIVIVALVUR,32608320001 with role 'admin' is authorized to execute 'technical' commands
   INFO: Using signature 'ROPKA,KIVIVALVUR,32608320001 2021-12-28T11:38:40Z' as config file version
   INFO: Config file version is 'ROPKA,KIVIVALVUR,32608320001 2021-12-28T11:38:40Z'
   INFO: Loading command 'collectors technical configuration' from file '/output/voting/HA-SETUP/config/HA-SETUP.technical.asice'
   command_file:INFO: Loading command file '/output/voting/HA-SETUP/config/HA-SETUP.technical.asice' (collectors technical configuration)
   command_file:INFO: Validating collectors technical configuration
   command_file:INFO: Files in collectors technical configuration package are valid
   INFO: Collectors technical configuration file loaded successfully
   INFO: Collectors technical configuration file is registered in management service
