.. Elektroonilise hääletamise infosüsteem IVXV

.. Käsureautiliidi ivxv-export-votes kasutusjuhend

.. SEE FAIL ON GENEREERITUD AUTOMAATTESTI POOLT:
.. * features/steps/autodoc.py
.. * features/autodoc-utilities.feature

.. _ivxv-export-votes:

ivxv-export-votes
`````````````````

:command:`ivxv-export-votes --help`:

.. code-block:: text

   Export collected votes.

   This utility copies current ballot box from voting service to backup
   service and consolidates collected votes from all ballot box backup files.

   Usage: ivxv-export-votes [--consolidate] <output-file>

   Options:
       --consolidate   Consolidate all ballot box backup files
