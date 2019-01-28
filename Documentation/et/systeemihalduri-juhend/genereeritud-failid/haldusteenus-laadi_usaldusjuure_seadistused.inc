.. Elektroonilise hääletamise infosüsteem IVXV

.. None

.. SEE FAIL ON GENEREERITUD AUTOMAATTESTI POOLT:
.. * steps/test_util/report.py

..  Feature: Kogumisteenuse käivitamine kõrgkäideldava konfiguratsiooniga
             features/collector-ha-setup.feature:6
   Scenario: Kogumisteenuse tehniliste seadistuste rakendamine
             features/collector-ha-setup.feature:366
       Step: Haldur laadib haldusteenusesse usaldusjuure seadistused
             features/collector-ha-setup.feature:374

.. code-block:: text

   ivxv-admin@admin $ ivxv-cmd-load trust /output/voting/HA-SETUP/config/HA-SETUP.trust.bdoc
   command_file:INFO: Loading command file "/output/voting/HA-SETUP/config/HA-SETUP.trust.bdoc" (trust root configuration)
   command_file:INFO: Validating trust root configuration
   command_file:INFO: Files in trust root configuration package are valid
   INFO: Config file is signed by: ORAV,IVAN,30809010001 2019-01-25T14:30:11Z
   INFO: User ORAV,IVAN,30809010001 with role "admin" is authorized to execute "trust" commands
   INFO: Using signature "ORAV,IVAN,30809010001 2019-01-25T14:30:11Z" as config file version
   INFO: Config file version is "ORAV,IVAN,30809010001 2019-01-25T14:30:11Z"
   INFO: Loading command "trust root configuration" from file /output/voting/HA-SETUP/config/HA-SETUP.trust.bdoc
   command_file:INFO: Loading command file "/output/voting/HA-SETUP/config/HA-SETUP.trust.bdoc" (trust root configuration)
   command_file:INFO: Validating trust root configuration
   command_file:INFO: Files in trust root configuration package are valid
   INFO: Resetting collector management database
   db:INFO: Initializing management database /var/lib/ivxv/db/ivxv-management.db
   Removing crontab (if exist)
   no crontab for ivxv-admin
   INFO: Trust root configuration file loaded successfully
   INFO: Resetting user permissions
   lib:INFO: Creating Apache Web Server user permission file /var/lib/ivxv/admin-ui-permissions/ORAV,IVAN,30809010001-admin
   INFO: Trust root configuration file is registered in management service