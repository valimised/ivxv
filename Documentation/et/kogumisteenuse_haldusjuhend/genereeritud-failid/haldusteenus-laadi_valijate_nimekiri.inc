.. Elektroonilise hääletamise infosüsteem IVXV

.. SEE FAIL ON GENEREERITUD AUTOMAATTESTI POOLT:
.. * steps/test_util/report.py

..  Feature: Kogumisteenuse käivitamine kõrgkäideldava konfiguratsiooniga
             features/end-to-end-10-collector-ha-setup.feature:7
   Scenario: Valijate algnimekirja rakendamine
             features/end-to-end-10-collector-ha-setup.feature:643
       Step: Nimekirjade haldur laadib valijate algnimekirja haldusteenusesse
             features/end-to-end-10-collector-ha-setup.feature:667

.. code-block:: text

   ivxv-admin@admin $ ivxv-cmd-load voters /output/voting/HA-SETUP/config/voters-0.asice
   INFO: Config file is signed by: NÕID,VÄIKE,33333333333 2021-10-02T19:46:00Z
   INFO: User NÕID,VÄIKE,33333333333 with role 'election-conf-manager' is authorized to execute 'voters' commands
   INFO: Using signature 'NÕID,VÄIKE,33333333333 2021-10-02T19:46:00Z' as config file version
   INFO: Config file version is 'NÕID,VÄIKE,33333333333 2021-10-02T19:46:00Z'
   INFO: Loading command 'voters list' from file '/output/voting/HA-SETUP/config/voters-0.asice'
   command_file:INFO: Loading command file '/output/voting/HA-SETUP/config/voters-0.asice' (voters list)
   command_file:INFO: Validating voters list
   command_file:INFO: Files in voters list package are valid
   INFO: Some voting lists are already loaded, executing consistency checks: ivxv-config-validate --election=/etc/ivxv/election.bdoc --choices=/etc/ivxv/choices.bdoc --districts=/etc/ivxv/districts.bdoc --voters=/output/voting/HA-SETUP/config/voters-0.asice
   INFO: Validating elections configuration file '/etc/ivxv/election.bdoc'
   ...
   command_file:INFO: Files in voters list package are valid
   INFO: Config files are valid
   INFO: Detected election ID 'HA-SETUP' from election config
   INFO: Election ID consistency check succeeded
   INFO: Checking districts and choices lists consistency
   INFO: Checking districts and voter list changeset #0 consistency
   INFO: Voting lists consistency check succeeded
   INFO: Voters list file loaded successfully
   INFO: Voters list file is registered in management service