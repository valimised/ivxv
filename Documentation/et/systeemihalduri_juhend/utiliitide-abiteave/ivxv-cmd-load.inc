.. Elektroonilise hääletamise infosüsteem IVXV

.. Käsureautiliidi ivxv-cmd-load kasutusjuhend

.. SEE FAIL ON GENEREERITUD AUTOMAATTESTI POOLT:
.. * features/steps/autodoc.py
.. * features/autodoc-utilities.feature

.. _ivxv-cmd-load:

ivxv-cmd-load
`````````````

:command:`ivxv-cmd-load --help`:

.. code-block:: text

   Load command to IVXV Collector Management Service.

   Usage: ivxv-cmd-load [--autoapply] [--show-version] <type> FILE

   Options:
       <type>              Command type. Possible values are:
                           - election: election config
                           - technical: collector technical config
                           - trust: trust root config
                           - choices: choices list
                           - districts: districts list
                           - voters: voters list
                           - user: user account and role(s)
       --autoapply         Apply command file automatically (by Agent Daemon).
       --show-version      Output config file version and exit.
