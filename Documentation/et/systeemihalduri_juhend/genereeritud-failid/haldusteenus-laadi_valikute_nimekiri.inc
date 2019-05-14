.. Elektroonilise hääletamise infosüsteem IVXV

.. None

.. SEE FAIL ON GENEREERITUD AUTOMAATTESTI POOLT:
.. * steps/test_util/report.py

..  Feature: Kogumisteenuse käivitamine kõrgkäideldava konfiguratsiooniga
             features/collector-ha-setup.feature:6
   Scenario: Valikute nimekirja rakendamine
             features/collector-ha-setup.feature:443
       Step: Nimekirjade haldur laadib valikute nimekirja haldusteenusesse
             features/collector-ha-setup.feature:489

.. code-block:: text

   ivxv-admin@admin $ ivxv-cmd-load choices /output/voting/HA-SETUP/config/choices.asice
   INFO: Config file is signed by: TRUMM,KAPTEN,33333333333 2019-05-07T10:52:05Z
   INFO: User TRUMM,KAPTEN,33333333333 with role "election-conf-manager" is authorized to execute "choices" commands
   INFO: Using signature "TRUMM,KAPTEN,33333333333 2019-05-07T10:52:05Z" as config file version
   INFO: Config file version is "TRUMM,KAPTEN,33333333333 2019-05-07T10:52:05Z"
   INFO: Loading command "choices list" from file /output/voting/HA-SETUP/config/choices.asice
   command_file:INFO: Loading command file "/output/voting/HA-SETUP/config/choices.asice" (choices list)
   command_file:INFO: Validating choices list
   command_file:INFO: Files in choices list package are valid
   INFO: Choices list file loaded successfully
   INFO: Choices list file is registered in management service