.. Elektroonilise hääletamise infosüsteem IVXV

.. Käsureautiliidi ivxv-update-packages kasutusjuhend

.. SEE FAIL ON GENEREERITUD AUTOMAATTESTI POOLT:
.. * features/steps/autodoc.py
.. * features/autodoc-utilities.feature

.. _ivxv-update-packages:

ivxv-update-packages
````````````````````

:command:`ivxv-update-packages --help`:

.. code-block:: text

   Update service packages in IVXV service hosts.

   This utility checks versions of software packages in service hosts
   and installs new versions if required.

   Usage: ivxv-update-packages [--force]

   Options:
       --force     Update even package version does not require update
