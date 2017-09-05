.. Elektroonilise hääletamise infosüsteem IVXV

.. None

.. SEE FAIL ON GENEREERITUD AUTOMAATTESTI POOLT:
.. * steps/test_util/report.py

..  Feature: Kogumisteenuse paigaldamine kõrgkäideldava konfiguratsiooniga
             features/collector-ha-setup.feature:6
   Scenario: Kogumisteenuse seadistuste rakendamine
             features/collector-ha-setup.feature:492
       Step: Haldur laadib haldusteenusesse usaldusjuure seadistused
             features/collector-ha-setup.feature:493

.. code-block:: text

   ivxv-admin@admin $ ivxv-cmd-load trust /output/voting/HA-SETUP/config/trust.bdoc
   command_file:INFO: Loading command file "/output/voting/HA-SETUP/config/trust.bdoc" (trust root configuration)
   command_file:INFO: Command file loaded
   command_file:INFO: Command file successfully extracted
   command_file:INFO: Reading files from command file
   command_file:INFO: Validating trust root configuration
   command_file:INFO: Files in trust root configuration package are valid
   INFO: Config file is signed by: ORAV,IVAN,30809010001 2017-08-24T15:39:05Z
   INFO: User ORAV,IVAN,30809010001 with role "admin" is authorized to execute "trust" commands
   INFO: Using signature "ORAV,IVAN,30809010001 2017-08-24T15:39:05Z" as config file version
   INFO: Loading command trust from file /output/voting/HA-SETUP/config/trust.bdoc
   ...
   command_file:INFO: Command file successfully extracted
   command_file:INFO: Reading files from command file
   command_file:INFO: Validating trust root configuration
   command_file:INFO: Files in trust root configuration package are valid
   INFO: Resetting collector management database
   db:INFO: Initializing management database /var/lib/ivxv/db/ivxv-management.db
   INFO: Trust config file loaded successfully
   INFO: Resetting user permissions
   lib:INFO: Creating Apache Web Server user permission file /var/lib/ivxv/admin-ui-permissions/ORAV,IVAN,30809010001-admin