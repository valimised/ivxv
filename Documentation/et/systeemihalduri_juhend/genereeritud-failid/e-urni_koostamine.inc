.. Elektroonilise hääletamise infosüsteem IVXV

.. None

.. SEE FAIL ON GENEREERITUD AUTOMAATTESTI POOLT:
.. * steps/test_util/report.py

..  Feature: Kogumisteenuse käivitamine kõrgkäideldava konfiguratsiooniga
             features/collector-ha-setup.feature:6
   Scenario: Valimiste kokkuvõtete tegemine
             features/collector-ha-setup.feature:592
       Step: Haldur ekspordib haldusteenusest kogutud hääled
             features/collector-ha-setup.feature:610

.. code-block:: text

   ivxv-admin@admin $ ivxv-export-votes /output/voting/HA-SETUP/exported-votes.zip
   INFO: Creating backup copy from current ballot box
   SERVICE backup@backup.ivxv.ee: Copying list of known SSH hosts to service host
   # Preparing ballot box backup file in voting service voting@voting2.ivxv.ee
   # Creating ballot box backup file ballot-box-20190507_1054.zip
   Exporting votes: 0
   Exporting votes: 1
   Exporting votes: 2
   Exporting votes: 3
   Exporting votes: 4
   Exporting votes: 5
   Exporting votes: 6
   # Copying backup file ballot-box-20190507_1054.zip to backup service
   # Removing backup file ballot-box-20190507_1054.zip from voting service
   INFO: Copying ballot box to management service
   SERVICE backup@backup.ivxv.ee: Copying consolidated ballot box from service host
   INFO: Collected votes archive is written to /output/voting/HA-SETUP/exported-votes.zip