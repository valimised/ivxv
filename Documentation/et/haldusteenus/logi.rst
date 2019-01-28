..  IVXV kogumisteenuse haldusliidese kasutusjuhend

Haldussündmuste logi jälgimine
==============================

Kogumisteenuse sündmuste logi sirvimise leht avaneb menüüvalikust
``Logiraamat``.

Logis on järgmised väljad:

#. ``Aeg`` - sündmuse registreerimise aeg;

#. ``Teenus`` - teenuse identifikaator;

#. ``Tase`` - logisüdmuse tase (``INFO`` või ``ERROR``);

#. ``Sündmus`` - sündmuse liigi idnetifikaator;

#. ``Kirjeldus`` - sündmuse tekstiline kirjeldus.

Logi on võimalik filtreerida ja välja väärtuse järgi sortida.

Logisündmused
-------------

Kogumisteenuse olekud:

:COLLECTOR_INIT:
   Kogumisteenuse lähtestamine (käsuga :command:`ivxv-collector-init`);

:COLLECTOR_RESET:
   Kogumisteenuse seadistuste lähtestamine (usaldusjuure laadimine);

:COLLECTOR_STATE_CHANGE:
   Kogumisteenuse olekumuutus;

Korralduste laadimine:

:CMD_LOAD:
   Korraldusfaili haldusteenusesse laadimine;

:CMD_LOADED:
   Edukalt laaditud korraldusfaili registreerimine haldusteenuses;

:CMD_REMOVED:
   Korraldusfaili eemaldamine haldusteenusest;

Kasutajaõiguste muutused:

:PERMISSION_SET:
   Kasutajale õiguse määramine;

:PERMISSION_RESET:
   Kasutaja õiguste lähtestamine;

Hääletusperioodi registreerimine:

:SET_ELECTION_TIME:
   Hääletusperioodi algus- ja lõpuaegade registreerimine;

Mikroteenuste haldus:

:SERVICE_REGISTER:
   Teenuse registreerimine haldusteenuses;

:SERVICE_CONFIG_APPLY:
   Seadistuse rakendamine teenusele;

:SERVICE_STATE_CHANGE:
   Teenuse olekumuutus;

:SECRET_INSTALL:
   Saladuse laadimine teenusele.
