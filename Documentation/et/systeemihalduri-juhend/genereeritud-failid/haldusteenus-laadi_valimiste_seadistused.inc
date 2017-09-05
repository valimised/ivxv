.. Elektroonilise hääletamise infosüsteem IVXV

.. None

.. SEE FAIL ON GENEREERITUD AUTOMAATTESTI POOLT:
.. * steps/test_util/report.py

..  Feature: Kogumisteenuse paigaldamine kõrgkäideldava konfiguratsiooniga
             features/collector-ha-setup.feature:6
   Scenario: Kogumisteenuse seadistuste rakendamine
             features/collector-ha-setup.feature:492
       Step: Haldur laadib haldusteenusesse valimiste seadistused
             features/collector-ha-setup.feature:520

.. code-block:: text

   ivxv-admin@admin $ ivxv-cmd-load election /output/voting/HA-SETUP/config/election.bdoc
   INFO: Config file is signed by: TRUMM,KAPTEN,33333333333 2017-08-24T15:39:05Z
   INFO: User TRUMM,KAPTEN,33333333333 with role "election-conf-manager" is authorized to execute "election" commands
   INFO: Using signature "TRUMM,KAPTEN,33333333333 2017-08-24T15:39:05Z" as config file version
   INFO: Loading command election from file /output/voting/HA-SETUP/config/election.bdoc
   command_file:INFO: Loading command file "/output/voting/HA-SETUP/config/election.bdoc" (elections configuration)
   command_file:INFO: Command file loaded
   command_file:INFO: Command file successfully extracted
   command_file:INFO: Reading files from command file
   command_file:INFO: Validating elections configuration
   command_file:INFO: Files in elections configuration package are valid
   INFO: Election config file loaded successfully