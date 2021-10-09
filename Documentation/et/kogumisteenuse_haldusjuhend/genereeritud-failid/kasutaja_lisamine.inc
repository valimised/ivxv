.. Elektroonilise hääletamise infosüsteem IVXV

.. SEE FAIL ON GENEREERITUD AUTOMAATTESTI POOLT:
.. * steps/test_util/report.py

..  Feature: Kogumisteenuse käivitamine kõrgkäideldava konfiguratsiooniga
             features/end-to-end-10-collector-ha-setup.feature:7
   Scenario: Kogumisteenuse tehniliste seadistuste rakendamine
             features/end-to-end-10-collector-ha-setup.feature:405
       Step: Haldur määrab haldusteenuses kasutajale "NÕID,VÄIKE,33333333333" rolli "election-conf-manager"
             features/end-to-end-10-collector-ha-setup.feature:472

.. code-block:: text

   ivxv-admin@admin $ ivxv-cmd-load user /output/voting/HA-SETUP/config/user-NÕID,VÄIKE,33333333333-election-conf-manager.asice
   INFO: Config file is signed by: ORAV,IVAN,30809010001 2021-10-02T19:44:27Z
   INFO: User ORAV,IVAN,30809010001 with role 'admin' is authorized to execute 'user' commands
   INFO: Using signature 'ORAV,IVAN,30809010001 2021-10-02T19:44:27Z' as config file version
   INFO: Config file version is 'ORAV,IVAN,30809010001 2021-10-02T19:44:27Z'
   INFO: Loading command 'user permissions configuration' from file '/output/voting/HA-SETUP/config/user-NÕID,VÄIKE,33333333333-election-conf-manager.asice'
   command_file:INFO: Loading command file '/output/voting/HA-SETUP/config/user-NÕID,VÄIKE,33333333333-election-conf-manager.asice' (user permissions configuration)
   command_file:INFO: Validating user permissions configuration
   command_file:INFO: Files in user permissions configuration package are valid
   INFO: User permissions configuration file loaded successfully
   INFO: Resetting user 'NÕID,VÄIKE,33333333333' permissions