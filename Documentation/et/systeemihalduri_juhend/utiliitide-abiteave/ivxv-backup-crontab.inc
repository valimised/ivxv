.. Elektroonilise hääletamise infosüsteem IVXV

.. Käsureautiliidi ivxv-backup-crontab kasutusjuhend

.. SEE FAIL ON GENEREERITUD AUTOMAATTESTI POOLT:
.. * features/steps/autodoc.py
.. * features/autodoc-utilities.feature

.. _ivxv-backup-crontab:

ivxv-backup-crontab
```````````````````

:command:`ivxv-backup-crontab --help`:

.. code-block:: text

   Generate crontab for IVXV backup automation.

   This utility must be called as editor by crontab utility:

       $ env VISUAL=ivxv-backup-crontab crontab -e

   Usage: ivxv-backup-crontab <filename>
