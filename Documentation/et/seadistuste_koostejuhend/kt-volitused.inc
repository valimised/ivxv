..  IVXV kogumisteenuse haldusteenuse kirjeldus

.. _kt-management:

Kogumisteenuse volituste kirjeldamine
-------------------------------------

Kasutajatele volituste määramine käib süsteemis kirjeldatud rollide kaudu.
Igale rollile on määratud komplekt õigusi ja kasutajal on kõik volitused, mis
talle seotud rollide kaudu on määratud.

Volituste määramise korraldus määrab ühele kasutajale tema rollid süsteemis.

Volitused koostatakse JSON-vormingus failina, millega määratakse:

#. Korralduse sisu (``action=user-permissions``);

#. Volitatud kasutaja *Common Name* väli tema ID-kaardilt (väli ``cn``);

#. Kasutaja rollide nimekiri komadega eraldatud nimekirjana (väli ``roles``).

Faili vorming:

.. code-block:: JavaScript

   {
       "action": "user-permissions",
       "cn": "<kasutaja-CN">,
       "roles": "roll1[,roll2]"
   }


.. _rollid:

Rollid
******

Kogumisteenuses on järgnevad rollid:

#. **Kogumisteenuse haldur** (``admin``) on ette nähtud kogumisteenuse
   tehniliseks haldamiseks;

#. **Valimiste haldur** (``election-conf-manager``) on ette nähtud valimiste
   seadistuste kehtestamiseks;

#. **Vaataja** (``viewer``) on ette nähtud haldusteenuse kaudu väljastatavate
   andmete vaatamiseks;

#. **Õigusteta kasutaja** (``none``). See roll on ette nähtud kasutajakonto
   kirje hoidmiseks olukorras, kus kasutajale pole ühtegi teist rolli määratud
   (näiteks pärast lisamist või pärast kõigist teistest rollidest eemaldamist).

.. list-table::
   Rollide ja volituste maatriks
   :header-rows: 1

   *  -
      - admin
      - election- conf- manager
      - viewer
      - none

   *  - Üldseisundi ja statistika vaatamine
      - ✓
      - ✓
      - ✓
      - `-`

   *  - Valimiste seadistuste rakendamine
      - ✓
      - ✓
      - `-`
      - `-`

   *  - E-urni allalaadimine
      - ✓
      - ✓
      - `-`
      - `-`

   *  - Kasutajate haldus
      - ✓
      - `-`
      - `-`
      - `-`

   *  - Tehnilise seadistuse rakendamine
      - ✓
      - `-`
      - `-`
      - `-`

   *  - Logide vaatamine
      - ✓
      - `-`
      - `-`
      - `-`


Volituste reeglid
*****************

* Kasutaja võib olla mitmes rollis korraga;

* Roll annab kasutajale rolliga seotud õigused, ükski roll õigusi ära ei võta.

.. vim:syntax=rst:
