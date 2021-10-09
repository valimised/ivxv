.. Elektroonilise hääletamise infosüsteem IVXV

.. Käsureautiliidi ivxv-voting-sessions kasutusjuhend

.. SEE FAIL ON GENEREERITUD AUTOMAATTESTI POOLT:
.. * features/steps/autodoc.py
.. * features/autodoc-utilities.feature

.. _ivxv-voting-sessions:

ivxv-voting-sessions
````````````````````

:command:`ivxv-voting-sessions --help`:

.. code-block:: text

   Import list of voting sessions from Log Monitor.

   Session data is in CSV format.

   Usage: ivxv-voting-sessions (vote | verify) <output_file> [--anonymize]
               [--log-level=<level>]

   Options:
       <output_file>               Write sessions to file.
       --anonymize                 Anonymize session data
                                   (IP addresses and ID codes).
       --log-level=<level>         Logging level [Default: INFO].
