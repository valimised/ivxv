.. Elektroonilise hääletamise infosüsteem IVXV

.. None

.. SEE FAIL ON GENEREERITUD AUTOMAATTESTI POOLT:
.. * steps/test_util/report.py

..  Feature: Kogumisteenuse käivitamine kõrgkäideldava konfiguratsiooniga
             features/collector-ha-setup.feature:6
   Scenario: Korduvhääletamine ringkonna vahetuse järel
             features/collector-ha-setup.feature:563
       Step: Nimekirjade haldur laadib valijate nimekirja #4 haldusteenusesse
             features/collector-ha-setup.feature:572

.. code-block:: text

   ivxv-admin@admin $ ivxv-cmd-load voters /output/voting/HA-SETUP/config/voters-4.asice
   INFO: Config file is signed by: TRUMM,KAPTEN,33333333333 2019-02-11T11:10:50Z
   INFO: User TRUMM,KAPTEN,33333333333 with role "election-conf-manager" is authorized to execute "voters" commands
   INFO: Using signature "TRUMM,KAPTEN,33333333333 2019-02-11T11:10:50Z" as config file version
   INFO: Config file version is "TRUMM,KAPTEN,33333333333 2019-02-11T11:10:50Z"
   INFO: Loading command "voters list" from file /output/voting/HA-SETUP/config/voters-4.asice
   command_file:INFO: Loading command file "/output/voting/HA-SETUP/config/voters-4.asice" (voters list)
   command_file:INFO: Validating voters list
   command_file:INFO: Files in voters list package are valid
   INFO: Some voting lists are already loaded, executing consistency checks: ivxv-config-validate --choices=/etc/ivxv/choices.bdoc --voters=/output/voting/HA-SETUP/config/voters-4.asice --voters=/etc/ivxv/voters01.bdoc --voters=/etc/ivxv/voters02.bdoc
   INFO: Validating choices list file /etc/ivxv/choices.bdoc
   ...
   command_file:INFO: Validating voters list
   command_file:INFO: Files in voters list package are valid
   INFO: Validating voters list file /etc/ivxv/voters02.bdoc
   command_file:INFO: Loading command file "/etc/ivxv/voters02.bdoc" (voters list)
   command_file:INFO: Validating voters list
   command_file:INFO: Files in voters list package are valid
   INFO: Config files are valid
   INFO: Voters list file loaded successfully
   INFO: Voters list file is registered in management service