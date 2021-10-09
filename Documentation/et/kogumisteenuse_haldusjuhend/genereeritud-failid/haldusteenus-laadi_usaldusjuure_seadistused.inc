.. Elektroonilise hääletamise infosüsteem IVXV

.. SEE FAIL ON GENEREERITUD AUTOMAATTESTI POOLT:
.. * steps/test_util/report.py

..  Feature: Kogumisteenuse käivitamine kõrgkäideldava konfiguratsiooniga
             features/end-to-end-10-collector-ha-setup.feature:7
   Scenario: Kogumisteenuse tehniliste seadistuste rakendamine
             features/end-to-end-10-collector-ha-setup.feature:405
       Step: Haldur laadib haldusteenusesse usaldusjuure seadistused
             features/end-to-end-10-collector-ha-setup.feature:413

.. code-block:: text

   ivxv-admin@admin $ ivxv-cmd-load trust /output/voting/HA-SETUP/config/HA-SETUP.trust.asice
   command_file:INFO: Loading command file '/output/voting/HA-SETUP/config/HA-SETUP.trust.asice' (trust root configuration)
   command_file:INFO: Validating trust root configuration
   command_file:INFO: Files in trust root configuration package are valid
   INFO: Config file is signed by: ORAV,IVAN,30809010001 2021-10-02T19:39:58Z
   INFO: User ORAV,IVAN,30809010001 with role 'admin' is authorized to execute 'trust' commands
   INFO: Using signature 'ORAV,IVAN,30809010001 2021-10-02T19:39:58Z' as config file version
   INFO: Config file version is 'ORAV,IVAN,30809010001 2021-10-02T19:39:58Z'
   INFO: Loading command 'trust root configuration' from file '/output/voting/HA-SETUP/config/HA-SETUP.trust.asice'
   command_file:INFO: Loading command file '/output/voting/HA-SETUP/config/HA-SETUP.trust.asice' (trust root configuration)
   command_file:INFO: Validating trust root configuration
   command_file:INFO: Files in trust root configuration package are valid
   INFO: Resetting collector management database
   db:INFO: Initializing management database '/var/lib/ivxv/db/ivxv-management.db'
   Removing crontab (if exist)
   no crontab for ivxv-admin
   INFO: Trust root configuration file loaded successfully
   INFO: Resetting user permissions
   INFO: Trust root configuration file is registered in management service