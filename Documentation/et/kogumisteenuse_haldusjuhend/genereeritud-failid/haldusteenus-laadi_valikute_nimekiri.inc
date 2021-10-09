.. Elektroonilise hääletamise infosüsteem IVXV

.. SEE FAIL ON GENEREERITUD AUTOMAATTESTI POOLT:
.. * steps/test_util/report.py

..  Feature: Kogumisteenuse käivitamine kõrgkäideldava konfiguratsiooniga
             features/end-to-end-10-collector-ha-setup.feature:7
   Scenario: Valikute nimekirja rakendamine
             features/end-to-end-10-collector-ha-setup.feature:529
       Step: Nimekirjade haldur laadib valikute nimekirja haldusteenusesse
             features/end-to-end-10-collector-ha-setup.feature:575

.. code-block:: text

   ivxv-admin@admin $ ivxv-cmd-load choices /output/voting/HA-SETUP/config/choices.asice
   INFO: Config file is signed by: NÕID,VÄIKE,33333333333 2021-10-02T19:45:46Z
   INFO: User NÕID,VÄIKE,33333333333 with role 'election-conf-manager' is authorized to execute 'choices' commands
   INFO: Using signature 'NÕID,VÄIKE,33333333333 2021-10-02T19:45:46Z' as config file version
   INFO: Config file version is 'NÕID,VÄIKE,33333333333 2021-10-02T19:45:46Z'
   INFO: Loading command 'choices list' from file '/output/voting/HA-SETUP/config/choices.asice'
   command_file:INFO: Loading command file '/output/voting/HA-SETUP/config/choices.asice' (choices list)
   command_file:INFO: Validating choices list
   command_file:INFO: Files in choices list package are valid
   INFO: Choices list file loaded successfully
   INFO: Choices list file is registered in management service