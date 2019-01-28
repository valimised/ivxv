.. Elektroonilise hääletamise infosüsteem IVXV

.. Käsureautiliidi ivxv-admin-sudo kasutusjuhend

.. SEE FAIL ON GENEREERITUD AUTOMAATTESTI POOLT:
.. * features/steps/autodoc.py
.. * features/autodoc-utilities.feature

.. _ivxv-admin-sudo:

ivxv-admin-sudo
```````````````

:command:`ivxv-admin-sudo --help`:

.. code-block:: text

   Usage:
       ivxv-admin-sudo backup-ballot-box <voting-host> <service-id>
                                         <backup-filename>
           Backup ballot box (in backup service)

       ivxv-admin-sudo backup-log <log-host> <backup-filename>
           Backup log file (in backup service)

       ivxv-admin-sudo create-ssh-access <account-name>
           Create management service access to account in service host

       ivxv-admin-sudo init-host
           Initialize service host

       ivxv-admin-sudo init-service <service-id>
           Initialize service data directory.
           Value 'backup' is used for all backup services

       ivxv-admin-sudo install-pkg <package-filename>
           Install IVXV package with dependencies

       ivxv-admin-sudo prepare-ballot-box-backup <service-id> <backup-filename>
           Prepare votes backup file in voting service

       ivxv-admin-sudo remove-admin-root-access
           Remove management service access to service host root account

       ivxv-admin-sudo rsyslog-config-apply
           Apply rsyslog config file for IVXV logging
