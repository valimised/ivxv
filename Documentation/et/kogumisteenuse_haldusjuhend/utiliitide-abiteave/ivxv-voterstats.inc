.. Elektroonilise hääletamise infosüsteem IVXV

.. Käsureautiliidi ivxv-voterstats kasutusjuhend

.. SEE FAIL ON GENEREERITUD AUTOMAATTESTI POOLT:
.. * features/steps/autodoc.py
.. * features/autodoc-utilities.feature

.. _ivxv-voterstats:

ivxv-voterstats
```````````````

:command:`ivxv-voterstats --help`:

.. code-block:: text

   Import voter stats from voting service and export common stats to VIS.

   Usage: ivxv-voterstats <TYPE> [--action=<action>] [--file=<file>]
               [--service-id=<service_id>] [--log-level=<level>]

   Options:
       <TYPE>                      Stats type "common" or "detail".
       --action=<action>           Limit actions for "common" stats type.
                                   Possible values are "import" and "export".
                                   [Default: all]
       --file=<file>               Path to stats file.
       --service-id=<service_id>   Voting service ID [Default: random].
       --log-level=<level>         Logging level [Default: INFO].
