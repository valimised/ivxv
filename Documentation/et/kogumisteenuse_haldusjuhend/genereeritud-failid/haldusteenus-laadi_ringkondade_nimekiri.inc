.. Elektroonilise hääletamise infosüsteem IVXV

.. SEE FAIL ON GENEREERITUD AUTOMAATTESTI POOLT:
.. * steps/test_util/report.py

..  Feature: Kogumisteenuse käivitamine kõrgkäideldava konfiguratsiooniga
             features/end-to-end-10-collector-ha-setup.feature:7
   Scenario: Ringkondade nimekirja rakendamine
             features/end-to-end-10-collector-ha-setup.feature:587
       Step: Nimekirjade haldur laadib ringkondade nimekirja haldusteenusesse
             features/end-to-end-10-collector-ha-setup.feature:631

.. code-block:: text

   ivxv-admin@admin $ ivxv-cmd-load districts /output/voting/HA-SETUP/config/districts.asice
   INFO: Config file is signed by: NÕID,VÄIKE,33333333333 2021-10-02T19:45:51Z
   INFO: User NÕID,VÄIKE,33333333333 with role 'election-conf-manager' is authorized to execute 'districts' commands
   INFO: Using signature 'NÕID,VÄIKE,33333333333 2021-10-02T19:45:51Z' as config file version
   INFO: Config file version is 'NÕID,VÄIKE,33333333333 2021-10-02T19:45:51Z'
   INFO: Loading command 'districts list' from file '/output/voting/HA-SETUP/config/districts.asice'
   command_file:INFO: Loading command file '/output/voting/HA-SETUP/config/districts.asice' (districts list)
   command_file:INFO: Validating districts list
   command_file:INFO: Files in districts list package are valid
   INFO: Some voting lists are already loaded, executing consistency checks: ivxv-config-validate --election=/etc/ivxv/election.bdoc --choices=/etc/ivxv/choices.bdoc --districts=/output/voting/HA-SETUP/config/districts.asice
   INFO: Validating elections configuration file '/etc/ivxv/election.bdoc'
   ...
   command_file:INFO: Files in districts list package are valid
   INFO: Config files are valid
   INFO: Detected election ID 'HA-SETUP' from election config
   INFO: Election ID consistency check succeeded
   INFO: Checking districts and choices lists consistency
   INFO: Voting lists consistency check succeeded
   INFO: Writing simplified district list to '/var/lib/ivxv/admin-ui-data/districts.json'
   INFO: Districts list file loaded successfully
   INFO: Districts list file is registered in management service