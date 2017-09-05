.. Elektroonilise hääletamise infosüsteem IVXV

.. Käsureautiliidi ivxv-status kasutusjuhend

.. SEE FAIL ON GENEREERITUD AUTOMAATTESTI POOLT:
.. * features/steps/autodoc.py
.. * features/autodoc-utilities.feature

.. _ivxv-status:

ivxv-status
```````````

:command:`ivxv-status --help`:

.. code-block:: text

   Output IVXV collector state.

   Usage: ivxv-status [--json] [--filter=<filter-type> ...]
                      [--service=<service-id> ...]

   Options:
       --json                  Output full data in JSON format.
                               Note: filters have no effect in JSON output.
       --filter=<filter-type>  Filter output by section. Possible values are:
                                 * collector - collector status;
                                 * election - election data;
                                 * config - versions of loaded config;
                                 * list - versions of loades lists;
                                 * service - service information;
                                 * ext - external service information;
                                 * storage - storage information;
                                 * smart - output only relevant data;
                                 * all - output all data;
                               [Default: smart].
       --service=<service-id>  Filter output by service ID.
                               Note: This filter conflicts other section
                               filters than "smart" or "service".
