.. Elektroonilise hääletamise infosüsteem IVXV

.. Käsureautiliidi ivxv-config-validate kasutusjuhend

.. SEE FAIL ON GENEREERITUD AUTOMAATTESTI POOLT:
.. * features/steps/autodoc.py
.. * features/autodoc-utilities.feature

.. _ivxv-config-validate:

ivxv-config-validate
````````````````````

:command:`ivxv-config-validate --help`:

.. code-block:: text

   Validate IVXV collector config files.

   Validate single config files. Also validate voting lists consistency if
   multiple lists are provided.

   Usage:
       ivxv-config-validate [--plain] [--trust=<trust-file>]
           [--technical=<technical-file>] [--election=<election-file>]
           [--choices=<choices-file>] [--districts=<districts-file>]
           [--voters=<voters-file> ...]

   Options:
       --plain     Validate plain config file (Default: BDOC container)
