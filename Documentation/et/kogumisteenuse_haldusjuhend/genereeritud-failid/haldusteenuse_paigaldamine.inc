.. Elektroonilise hääletamise infosüsteem IVXV

.. SEE FAIL ON GENEREERITUD AUTOMAATTESTI POOLT:
.. * steps/test_util/report.py

..  Feature: Kogumisteenuse käivitamine kõrgkäideldava konfiguratsiooniga
             features/end-to-end-10-collector-ha-setup.feature:7
   Scenario: Kogumisteenuse haldusteenuse paigaldus
             features/end-to-end-10-collector-ha-setup.feature:370
       Step: Haldur paigaldab haldusteenuse
             features/end-to-end-10-collector-ha-setup.feature:376

.. code-block:: text

   root@admin # dpkg -i /etc/ivxv/debs/ivxv-common_1.7.8_all.deb /etc/ivxv/debs/ivxv-admin_1.7.8_amd64.deb
   Selecting previously unselected package ivxv-common.
   (Andmebaasi lugemine ... 13710 files and directories currently installed.)
   Preparing to unpack .../debs/ivxv-common_1.7.8_all.deb ...
   Unpacking ivxv-common (1.7.8) ...
   Selecting previously unselected package ivxv-admin.
   Preparing to unpack .../ivxv-admin_1.7.8_amd64.deb ...
   Unpacking ivxv-admin (1.7.8) ...
   Paki ivxv-common (1.7.8) paikasättimine ...
   # Adding user group 'ivxv'
   Adding group `ivxv' (GID 109) ...
   ...
     systemctl restart apache2
   # Starting Apache web server
   # Restarting rsyslog log server
   Created symlink /etc/systemd/system/multi-user.target.wants/ivxv-admin.service → /lib/systemd/system/ivxv-admin.service.
   Created symlink /etc/systemd/system/multi-user.target.wants/ivxv-admin-agent.service → /lib/systemd/system/ivxv-admin-agent.service.
   /usr/lib/python3/dist-packages/schematics/validate.py:121: SyntaxWarning: "is" with a literal. Did you mean "=="?
     if not kwargs or kwargs.pop('context', 0) is 0:
   Processing triggers for rsyslog (8.2001.0-1ubuntu1.1) ...
   invoke-rc.d: WARNING: No init system and policy-rc.d missing! Defaulting to block.
