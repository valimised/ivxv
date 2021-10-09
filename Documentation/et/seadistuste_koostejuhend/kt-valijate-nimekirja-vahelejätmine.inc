..  IVXV seadistuste koostamise juhend

.. _kt-voter-list-skip:

Valijate nimekirja vahelejätmine
--------------------------------

Valijate nimekirja vahelejätmine tähendab haldusteenuses registreerinud
vigase valijate nimekirja asendamist tühja nimekirjaga.

Protseduur on vajalik olukorras, kus haldusteenus on Valimiste Infosüsteemist
alla laadinud ja haldusteenuses registreerinud valijate muudatusnimekirja, mida
pole võimalik kogumisteenusele rakendada (nimekiri on vigane või pole kooskõlas
teiste kogumisteenuste seadistustega).

Valijate nimekirja vahelejätmine korralduse koostab valimiste korraldaja.

Seadistusfaili nimi peab alati lõppema stringiga :file:`.skip.yaml`.

:election:

        Kohustuslik väli.
        Valimise unikaalne identifikaator.

:skip_voter_list:

        Kohustuslik väli.
        Vahele jäetava valijate muudatusnimekirja versioon.

:changeset:

        Kohustuslik väli.
        Vahele jäetava valijate muudatusnimekirja järjekorranumber.

Näide
*****

:file:`example.voters.skip.yaml`:

.. literalinclude:: config-examples/example.voters.skip.yaml
   :language: yaml
   :linenos:

.. vim:syntax=rst:
