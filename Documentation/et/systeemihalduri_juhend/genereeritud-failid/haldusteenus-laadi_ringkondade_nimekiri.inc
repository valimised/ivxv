.. Elektroonilise hääletamise infosüsteem IVXV

.. None

.. SEE FAIL ON GENEREERITUD AUTOMAATTESTI POOLT:
.. * steps/test_util/report.py

..  Feature: Kogumisteenuse käivitamine kõrgkäideldava konfiguratsiooniga
             features/collector-ha-setup.feature:6
   Scenario: Hääletamise nimekirjade rakendamine
             features/collector-ha-setup.feature:425
       Step: Nimekirjade haldur laadib ringkondade nimekirja haldusteenusesse
             features/collector-ha-setup.feature:513

.. code-block:: text

   ivxv-admin@admin $ ivxv-cmd-load districts /output/voting/HA-SETUP/config/districts.bdoc
   INFO: Config file is signed by: TRUMM,KAPTEN,33333333333 2018-11-30T10:12:38Z
   INFO: User TRUMM,KAPTEN,33333333333 with role "election-conf-manager" is authorized to execute "districts" commands
   INFO: Using signature "TRUMM,KAPTEN,33333333333 2018-11-30T10:12:38Z" as config file version
   INFO: Loading command "districts list" from file /output/voting/HA-SETUP/config/districts.bdoc
   command_file:INFO: Loading command file "/output/voting/HA-SETUP/config/districts.bdoc" (districts list)
   command_file:INFO: Validating districts list
   command_file:INFO: Files in districts list package are valid
   INFO: Some voting lists are already loaded, executing consistency checks: ivxv-config-validate --choices=/etc/ivxv/choices.bdoc --districts=/output/voting/HA-SETUP/config/districts.bdoc
   INFO: Validating choices list file /etc/ivxv/choices.bdoc
   command_file:INFO: Loading command file "/etc/ivxv/choices.bdoc" (choices list)
   ...
   command_file:INFO: Loading command file "/output/voting/HA-SETUP/config/districts.bdoc" (districts list)
   command_file:INFO: Validating districts list
   command_file:INFO: Files in districts list package are valid
   INFO: Config files are valid
   INFO: Checking districts and choices lists consistency
   INFO: Voting lists consistency check succeeded
   INFO: Writing simplified district list to /var/lib/ivxv/admin-ui-data/districts.json
   INFO: Districts list file loaded successfully
   INFO: Districts list file is registered in management service