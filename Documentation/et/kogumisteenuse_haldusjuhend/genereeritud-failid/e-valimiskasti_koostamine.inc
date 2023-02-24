.. Elektroonilise hääletamise infosüsteem IVXV

.. SEE FAIL ON GENEREERITUD AUTOMAATTESTI POOLT:
.. * steps/test_util/report.py

..  Feature: Kogumisteenuse käivitamine kõrgkäideldava konfiguratsiooniga
             features/end-to-end-10-collector-ha-setup.feature:7
   Scenario: Valimiste kokkuvõtete tegemine
             features/end-to-end-10-collector-ha-setup.feature:892
       Step: Haldur ekspordib haldusteenusest kogutud hääled
             features/end-to-end-10-collector-ha-setup.feature:970

.. code-block:: text

   ivxv-admin@admin $ ivxv-export-votes /output/voting/HA-SETUP/exported-votes.zip
   INFO: Creating backup copy from current ballot box
   SERVICE backup@backup.ivxv.ee: Copying list of known SSH hosts to service host
   # Preparing ballot box backup file in voting service voting@voting1.ivxv.ee
   # Creating ballot box backup file ballot-box-20211228_1155.zip
   Exporting votes: 0
   Exporting votes: 1
   Exporting votes: 2
   Exporting votes: 3
   Exporting votes: 4
   Exporting votes: 5
   Exporting votes: 6
   Exporting votes: 7
   Exporting votes: 8
   # Copying backup file ballot-box-20211228_1155.zip to backup service
   # Removing backup file ballot-box-20211228_1155.zip from voting service
   INFO: Copying ballot box to management service
   SERVICE backup@backup.ivxv.ee: Copying ballot box from service host
   INFO: Collected votes archive is written to '/output/voting/HA-SETUP/exported-votes.zip'
