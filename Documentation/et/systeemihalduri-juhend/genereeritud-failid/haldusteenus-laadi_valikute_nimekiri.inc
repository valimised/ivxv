.. Elektroonilise hääletamise infosüsteem IVXV

.. None

.. SEE FAIL ON GENEREERITUD AUTOMAATTESTI POOLT:
.. * steps/test_util/report.py

..  Feature: Kogumisteenuse käivitamine kõrgkäideldava konfiguratsiooniga
             features/collector-ha-setup.feature:6
   Scenario: Valikute nimekirja rakendamine
             features/collector-ha-setup.feature:437
       Step: Nimekirjade haldur laadib valikute nimekirja haldusteenusesse
             features/collector-ha-setup.feature:483

.. code-block:: text

   ivxv-admin@admin $ ivxv-cmd-load choices /output/voting/HA-SETUP/config/choices.bdoc
   INFO: Config file is signed by: TRUMM,KAPTEN,33333333333 2019-01-25T14:37:28Z
   INFO: User TRUMM,KAPTEN,33333333333 with role "election-conf-manager" is authorized to execute "choices" commands
   INFO: Using signature "TRUMM,KAPTEN,33333333333 2019-01-25T14:37:28Z" as config file version
   INFO: Config file version is "TRUMM,KAPTEN,33333333333 2019-01-25T14:37:28Z"
   INFO: Loading command "choices list" from file /output/voting/HA-SETUP/config/choices.bdoc
   command_file:INFO: Loading command file "/output/voting/HA-SETUP/config/choices.bdoc" (choices list)
   command_file:INFO: Validating choices list
   command_file:INFO: Files in choices list package are valid
   INFO: Choices list file loaded successfully
   INFO: Choices list file is registered in management service