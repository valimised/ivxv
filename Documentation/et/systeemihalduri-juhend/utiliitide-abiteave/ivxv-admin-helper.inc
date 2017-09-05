.. Elektroonilise hääletamise infosüsteem IVXV

.. Käsureautiliidi ivxv-admin-helper kasutusjuhend

.. SEE FAIL ON GENEREERITUD AUTOMAATTESTI POOLT:
.. * features/steps/autodoc.py
.. * features/autodoc-utilities.feature

.. _ivxv-admin-helper:

ivxv-admin-helper
`````````````````

:command:`ivxv-admin-helper --help`:

.. code-block:: text

   Usage:
       ivxv-admin-helper configure-etcd-apt-source
           Configure APT source for etcd

       ivxv-admin-helper create-ssh-access <account-name>
           Create management service access to account in service host

       ivxv-admin-helper init-host
           Initialize service host

       ivxv-admin-helper init-service <service-id>
           Initialize service data directory

       ivxv-admin-helper install-pkg <package-filename>
           Install IVXV package with dependencies

       ivxv-admin-helper remove-admin-root-access
           Remove management service access to service host root account

       ivxv-admin-helper rsyslog-config-apply
           Apply rsyslog config file for IVXV logging
