..  IVXV kogumisteenuse haldusjuhend

.. _haldusteenus:

Haldusteenus
============

Haldusteenus on kogumisteenuse haldamiseks mõeldud lahendus. Haldusteenus
paigaldatakse eraldiseisvasse masinasse ja selle kaudu toimub kogumisteenuse
juhtimine paigaldusest kuni seiskamiseni.

Haldusteenuse funktsioonid on:

#. Kogumisteenuse alamteenuste haldamine:

   #. Seadistuste ja valimisnimekirjade laadimine;

   #. Alamteenuste paigaldus selleks ettevalmistatud masinatesse;

   #. Alamteenustele seadistuste ja nimekirjade rakendamine;

#. E-urni koostamine töötlemiseks;

#. Valimiste üldstatistika jälgimine;

#. E-urni ja logide korrapärane varundamine;

#. Kogumisteenuse seisundi seire;

Haldusteenus suhtleb hallatavate teenustega üle `SSH
<https://en.wikipedia.org/wiki/Secure_Shell>`_-kanali. Suhtluse algatab alati
haldusteenus. Usaldus teenusmasinate vastu luuakse süsteemihalduri abiga pärast
teenuseid majutavate masinate paigaldamist.

Teenust majutava masina paigaldamise järel loob haldur haldusteenusele
ligipääsu teenusmasina juurkontole, et haldusteenusel oleks võimalik teenuse
tarkvara paigaldada.  Pärast viimase teenuse paigaldamist teenuseid majutavasse
masinasse eemaldab haldusteenus ligipääsu juurkontole.


Haldusteenuse koosseis
----------------------

Haldusteenuse kasutajaliides koosneb kahest osast:

#. Haldamise põhifunktsionaalsus on teostatud :ref:`käsureautiliitide
   <utiliidid>` abil;

#. Graafiline kasutajaliides on veebipõhine liides, mille funktsionaalsuse
   tagavad käsureautiliidid.

   .. seealso::

      Graafilise kasutajaliidese kasutusjuhend asub dokumendis
      ``IVXV kogumisteenuse haldusliidese kasutusjuhend``.

Lisaks töötavad deemonprotsessid:

#. Veebiserver graafilise kasutajaliidese jaoks;

#. Haldusdeemon veebiserveri poolt vahendatud päringute käivitamiseks;

#. Agentdeemon teenuste seisundi jälgimiseks.
