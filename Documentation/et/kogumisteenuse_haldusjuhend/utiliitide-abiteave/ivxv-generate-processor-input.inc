.. Elektroonilise hääletamise infosüsteem IVXV

.. Käsureautiliidi ivxv-generate-processor-input kasutusjuhend

.. SEE FAIL ON GENEREERITUD AUTOMAATTESTI POOLT:
.. * features/steps/autodoc.py
.. * features/autodoc-utilities.feature

.. _ivxv-generate-processor-input:

ivxv-generate-processor-input
`````````````````````````````

:command:`ivxv-generate-processor-input --help`:

.. code-block:: text

   Generate input for processor application.

   This utility generates ZIP container with data files
   for processor application to validate ballot box:

       1. District list;
       2. Voter lists;
       3. Validation key for vote registration requests;
       4. Configuration for processor application.

   Usage: ivxv-generate-processor-input <output-file>
