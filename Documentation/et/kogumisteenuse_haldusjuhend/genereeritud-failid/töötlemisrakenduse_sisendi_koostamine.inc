.. Elektroonilise hääletamise infosüsteem IVXV

.. SEE FAIL ON GENEREERITUD AUTOMAATTESTI POOLT:
.. * steps/test_util/report.py

..  Feature: Kogumisteenuse käivitamine kõrgkäideldava konfiguratsiooniga
             features/end-to-end-10-collector-ha-setup.feature:7
   Scenario: Valimiste kokkuvõtete tegemine
             features/end-to-end-10-collector-ha-setup.feature:892
       Step: Haldur genereerib haldusteenuses töötlemisrakenduse sisendi aluse
             features/end-to-end-10-collector-ha-setup.feature:973

.. code-block:: text

   ivxv-admin@admin $ ivxv-generate-processor-input /output/voting/HA-SETUP/processor.cfg.zip
   INFO: Generating processor application config
   command_file:INFO: Loading command file '/etc/ivxv/election.bdoc' (elections configuration)
   INFO: Creating input file for processor application
   INFO: Preparing container structure in directory '/tmp/tmp6jswwi8i'
   INFO: Copying district list 'HA-SETUP.districts.json.bdoc'
   INFO: Copying voter list signing key 'voterfile.pub.key'
   INFO: Copying voter list #0 content '00.HA-SETUP.voters.utf'
   INFO: Copying voter list #0 signature '00.HA-SETUP.voters.sig'
   INFO: Copying voter list #1 content '01.HA-SETUP.voters.utf'
   INFO: Copying voter list #1 signature '01.HA-SETUP.voters.sig'
   ...
   INFO: Adding '02.HA-SETUP.voters.sig' to ZIP container
   INFO: Adding '02.HA-SETUP.voters.utf' to ZIP container
   INFO: Adding '03.HA-SETUP.voters.sig' to ZIP container
   INFO: Adding '03.HA-SETUP.voters.utf' to ZIP container
   INFO: Adding 'HA-SETUP.districts.json.bdoc' to ZIP container
   INFO: Adding 'HA-SETUP.processor.yaml' to ZIP container
   INFO: Adding 'ts.key' to ZIP container
   INFO: Adding 'voterfile.pub.key' to ZIP container
   INFO: Processor input is written to '/output/voting/HA-SETUP/processor.cfg.zip'
