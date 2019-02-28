.. Elektroonilise hääletamise infosüsteem IVXV

.. Haldusteenuse paigaldamine:

.. SEE FAIL ON GENEREERITUD AUTOMAATTESTI POOLT:
.. * steps/test_util/report.py

..  Feature: Kogumisteenuse käivitamine kõrgkäideldava konfiguratsiooniga
             features/collector-ha-setup.feature:6
   Scenario: Kogumisteenuse haldusteenuse paigaldus
             features/collector-ha-setup.feature:340
       Step: Haldur paigaldab haldusteenuse
             features/collector-ha-setup.feature:346

.. code-block:: text

   root@admin # dpkg -i /etc/ivxv/debs/ivxv-common_1.4.3_all.deb /etc/ivxv/debs/ivxv-admin_1.4.3_amd64.deb
   Selecting previously unselected package ivxv-common.
   (Andmebaasi lugemine ... 12741 files and directories currently installed.)
   Preparing to unpack .../debs/ivxv-common_1.4.3_all.deb ...
   Unpacking ivxv-common (1.4.3) ...
   Selecting previously unselected package ivxv-admin.
   Preparing to unpack .../ivxv-admin_1.4.3_amd64.deb ...
   Unpacking ivxv-admin (1.4.3) ...
   Paki ivxv-common (1.4.3) paikasättimine ...
   # Adding user group 'ivxv'
   Adding group `ivxv' (GID 108) ...
   ...
   Generating DH parameters, 2048 bit long safe prime, generator 2
   This is going to take a long time
   .....+..........................................................................................................................++*++*
   # Starting Apache web server
   # Restarting rsyslog log server
   Created symlink /etc/systemd/system/multi-user.target.wants/ivxv-admin.service → /lib/systemd/system/ivxv-admin.service.
   Created symlink /etc/systemd/system/multi-user.target.wants/ivxv-admin-agent.service → /lib/systemd/system/ivxv-admin-agent.service.
   Processing triggers for rsyslog (8.32.0-1ubuntu4) ...
   invoke-rc.d: WARNING: No init system and policy-rc.d missing! Defaulting to block.