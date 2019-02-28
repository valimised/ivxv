.. Elektroonilise hääletamise infosüsteem IVXV

.. None

.. SEE FAIL ON GENEREERITUD AUTOMAATTESTI POOLT:
.. * steps/test_util/report.py

..  Feature: Kogumisteenuse käivitamine kõrgkäideldava konfiguratsiooniga
             features/collector-ha-setup.feature:6
   Scenario: Valimiste kokkuvõtete tegemine
             features/collector-ha-setup.feature:589
       Step: Haldur ekspordib haldusteenusest kogutud hääled
             features/collector-ha-setup.feature:605

.. code-block:: text

   ivxv-admin@admin $ ivxv-consolidate-votes /output/voting/HA-SETUP/exported-votes.zip
   INFO: Creating backup copy from current ballot box
   Identity added: /var/lib/ivxv/user/ivxv-admin/.ssh/id_ed25519 (ivxv-admin@admin)
   SERVICE backup@backup.ivxv.ee: Copying list of known SSH hosts to service host
   # Preparing ballot box backup file in voting service voting@voting2.ivxv.ee
   # Creating ballot box backup file ballot-box-20190211_1112.zip
   Exporting votes: 0
   Exporting votes: 1
   Exporting votes: 2
   Exporting votes: 3
   Exporting votes: 4
   ...
   /var/backups/ivxv/ballot-box/ballot-box-20190211_1112.zip:  95% 23/24 votes/41212125078/20190211111017108+0000.ocsp
   added "/var/backups/ivxv/ballot-box/ballot-box-20190211_1112.zip" file "votes/41212125078/20190211111017108+0000.ocsp"
   /var/backups/ivxv/ballot-box/ballot-box-20190211_1112.zip: 100% 24/24 votes/41212125078/20190211111017108+0000.tspreg
   added "/var/backups/ivxv/ballot-box/ballot-box-20190211_1112.zip" file "votes/41212125078/20190211111017108+0000.tspreg"
   INFO: Copying ballot box to management service
   SERVICE backup@backup.ivxv.ee: Copying consolidated ballot box from service host
   INFO: Removing consolidated ballot box from backup service
   '/var/lib/ivxv/ballot-box-consolidated-20190211_1112.zip' eemaldatud
   INFO: Collected votes archive is written to /output/voting/HA-SETUP/exported-votes.zip