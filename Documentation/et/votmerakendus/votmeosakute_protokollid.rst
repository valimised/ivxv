..  IVXV tehniline dokumentatsioon

Eritüübiliste võtmeosakute protokollide liidestamine
====================================================

Võtmeosakute genereerimise protokolli liides
--------------------------------------------

Klass `ee.ivxv.key.protocol.GenerationProtocol` defineerib liidese, mida
ElGamali või RSA võtmeosakute genereerimise protokoll peab täitma. Liides on
järgnev::

    public interface GenerationProtocol {
        byte[] generateKey() throws ProtocolException;
    }

Eritüübiliste võtmeosakute genereerimiseks peab implementeerima `generateKey()`
meetodi, mis tagastab kodeeritud avaliku võtme X.509 sertifikaadina DER
formaadis. Protokolli klass peab olema paki `ee.ivxv.key.protocol.generation`
alampakis.

Protokolli parameetrid tuleb määrata protokolli klassiinstantsi
initsialiseerimise ajal.

Dekrüpteerimise protokolli liides
---------------------------------

Dekrüpteerimise ja võtmeosakute genereerimise protokoll ei pea olema üksüheses
seoses, st. võtmeosakute genereerimise protokollile võib vastata mitu
dekrüpteerimise protokolli. Seega on dekrüpteerimise protokolli liides
defineeritud sõltumatult võtmeosakute genereerimise liidesest. Protokoll peab
implementeerima `ee.ivxv.key.protocol.DecryptionProtocol` liidese::

    public interface DecryptionProtocol {
        ElGamalDecryptionProof decryptMessage(byte[] msg) throws ProtocolException;
    }

Meetod `decryptMessage()` võtab sisendina krüptogrammi DER formaadis ning
väljastab `ElGamalDecryptionProof` instantsi. Juhul kui protokoll ei toeta
lugemistõendi väljastamist, siis on vastavad väljad väärtustatud tühiväärtusega
(`null`).

Analoogselt võtmeosakute genereerimise protokollile tuleb protokolli
parameetrid määrata klassiinstantsi initsialiseerimise ajal.

Allkirjastamise protokolli liides
---------------------------------

Lisaks dekrüpteerimise protokollile saab implementeerida ka allkirjastamise
protokolli. Sarnaselt dekrüpteerimisprotokollile võib ühele võtmegenereerimise
meetodile vastata mitu allkirjastamise protokolli. Protokoll peab
implementeerima `ee.ivxv.key.protocol.SigningProtocol` liidese::

    public interface SigningProtocol {
        byte[] sign(byte[] msg) throws ProtocolException;
    }

Meetod `sign()` võtab sisendina sõnumi, mida soovitakse allkirjastada, ja
väljastab RSA-PSS allkirja järgnevate parameetritega:

.. _RSA-PSS parameetrid:

- sõnumi räsifunktsioon: SHA2-256
- maski genereerimise funktsioon: MGF1, maski räsifunktsioon SHA2-256 ja maski
  pikkus 32 baiti
- soola pikkus: 32 baiti
- sababait: `0xbc`

Analoogselt võtmeosakute genereerimise protokollile tuleb protokolli
parameetrid määrata klassiinstantsi initsialiseerimise ajal.

Toetatud protokollid
--------------------

Hetkel on teostatud järgnevad võtmeosakute genereerimise protokollid:

* `ee.ivxv.key.protocol.generation.desmedt.DesmedtGeneration`: Võtmeosakud on
  sellised, et oleks võimalik kasutada [DF89]_ hajutatud
  dekrüpteerimisprotokolli. Võtmeosakud salvestatakse otse PKCS15 liidest
  toetavale pääsmikule. Klassiinstantsi initsialiseerimise ajal saab anda
  järgnevaid argumente:

  + `PKCS15Card[] cards`: järjend objektides mis implementeerivad PKCS15Card
    liidest (nt. kiipkaardid või tarkvaralised pääsmikud).
  + `ElGamalParameters params`: ElGamali krüptosüsteemi parameetrid.
  + `ThresholdParameters tparams`: läviskeemi parameetrid.
  + `Rnd rnd`: juhuslikkuse sisend võtmeosakute genereerimisel
  + `byte[] cardShareAID`: võtmeosaku ligipääsuidentifikaator PKCS15
    pääsmikul. Defineerib, millise ligipääsutunnusega pääseb võtmeosakule
    ligi.
  + `byte[] cardShareName`: võtmeosaku identifikaator PKCS15 pääsmikul.

* `ee.ivxv.key.protocol.generation.shoup.ShoupGeneration`: Võtmeosakud on
  sellised, et oleks võimalik kasutada [Shoup00]_ põhinevat hajutatud
  allkirjastamisprotokolli. Võtmeosakud salvestatakse otse PKCS15 liidest
  toetavale pääsmikule. Klassiinstantsi initsialiseerimise ajal saab anda
  järgnevaid argumente:

  + `PKCS15Card[] cards`: järjend objektidest, mis implementeerivad PKCS15Card
    liidest (nt. kiipkaardid või tarkvaralised pääsmikud).
  + `int modLen`: RSA võtme pikkus bittides
  + `ThresholdParameters tparams`: läviskeemi parameetrid.
  + `Rnd rnd`: juhuslikkuse sisend võtmeosakute genereerimisel
  + `byte[] cardShareAID`: võtmeosaku ligipääsuidentifikaator PKCS15
    pääsmikul. Defineerib, millise ligipääsutunnusega pääseb võtmeosakule
    ligi.
  + `byte[] cardShareName`: võtmeosaku identifikaator PKCS15 pääsmikul.

On teostatud järgnevad dekrüpteerimise protokollid:

* `ee.ivxv.key.protocol.decryption.recover.RecoverDecryption`: Loetakse PKCS15
  liidest toetavatelt pääsmikelt võtmeosakud, rekonstrueeritakse nende abil
  operatiivmälus salajane võti ning teostatakse dekrüpteerimine lugemistõendiga.
  Klassiinstantsi initsialiseerimise ajal saab anda järgnevaid argumente:

  + `PKCS15Card[] cards`: järjend objektides mis implementeerivad PKCS15Card
    liidest (nt. kiipkaardid või tarkvaralised pääsmikud).
  + `ThresholdParameters tparams`: läviskeemi parameetrid.
  + `byte[] cardShareAID`: võtmeosaku ligipääsuidentifikaator PKCS15
    pääsmikul. Defineerib, millise ligipääsutunnusega pääseb võtmeosakule
    ligi.
  + `byte[] cardShareName`: võtmeosaku identifikaator PKCS15 pääsmikul.

On teostatud järgnevad allkirjastamise protokollid:

* `ee.ivxv.key.protocol.signing.shoup.ShoupSigning`: Loetakse PKCS15 liidest
  toetavatelt pääsmikelt võtmeosakud, konstrueeritakse mälus allkirjastamise
  osakud ilma võtit rekonstrueerimata ning kombineeritakse allkirjastamise
  osakud RSA-PSS allkirjaks. Klassiinstantsi initsialiseerimise ajal saab anda
  järgnevaid argumente:

  + `PKCS15Card[] cards`: järjend objektides mis implementeerivad PKCS15Card
    liidest (nt. kiipkaardid või tarkvaralised pääsmikud).
  + `ThresholdParameters tparams`: läviskeemi parameetrid.
  + `Rnd rnd`: juhuslikkuse sisend RSA-PSS allkirja soola genereerimisel.
  + `byte[] cardShareAID`: võtmeosaku ligipääsuidentifikaator PKCS15
    pääsmikul. Defineerib, millise ligipääsutunnusega pääseb võtmeosakule
    ligi.
  + `byte[] cardShareName`: võtmeosaku identifikaator PKCS15 pääsmikul.

Protokollide võtmerakendusega liidestamine
------------------------------------------

Järgnev kirjeldus käib nii võtmeosakute genereerimise ja dekrüpteerimise
protokollide kohta.

.. note:: Praegune kirjeldus on üldine. Kui konfi- ja argumentide parsimine on
   lõplikult välja töötatud ning protokollid võtmerakendusega liidestatud, siis
   tuleks järgnevat lõiku täiendada.

Liidestamaks uut protokolli võtmerakendusega, tuleb kõigepealt teostada vastava
protokolli liidest täitev klass. Võtmerakendus peab töö alguses seadistuse
töötlemise käigus aru saama kas käsureaargumentidest või seadistusfailist,
millist protokolli soovitakse kasutada. Seejärel tuleb vastava klassi staatilise
meetodi abil ülejäänud käsureaargumentide või seadistusfaili abil
initsialiseerida uus protokolliklassi instants. Seejärel tuleb genereerida võti
või dekrüpteerida sõnum.

Toetatud protokollide kirjeldused
---------------------------------

Shamiri saladuse jagamise skeem
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Olgu meil salajane väärtus :math:`s = a_0` ja soovime seda jagada :math:`n`
osapoole vahel selliselt, et vähemalt :math:`t` osapoolt saaksid selle saladuse
rekonstrueerida.  Selleks valime koefitsiendid :math:`a_1` kuni :math:`a_{t-1}`
ning vaatame polünoomi muutuja :math:`x` suhtes:

.. math::
    P(x) = a_{t-1}  x^{t-1} + .. + a_1  x + a_0

Olgu :math:`x_1` kuni :math:`x_n` nullist erinevad unikaalsed väärtused
(üldiselt :math:`1` kuni :math:`n`), sellisel juhul saame osakud :math:`s_i =
P(x_i)` ning salajase väärtuse :math:`s = P(0)`.

Geomeetriliselt vaadates on :math:`P(x)` polünoom ning osakud punktid sellel
polünoomil. Põhikoolimatemaatikast teame, et :math:`t-1` järku polünoomi
joonistamiseks piisab meile :math:`t` punktist (sirge jaoks kaks punkti,
parabooli jaoks kolm punkti jne.). Salajane väärtus on selle polünoomi väärtus
y-telje lõikepunktis.

Vaadates rekonstrueerimist arvuliselt, mitte geomeetriliselt, saame kasutades
Lagrange interpoleerimise meetodit. Tähist :math:`\prod` kasutame me mitme
liikmega korrutise tähistamiseks ja tähist :math:`\sum` kasutame me mitme
liikmega summa tähistamiseks.

Nüüd, tähistame lisaks :math:`t` osapoolt, kes osalevad salajase väärtuse
rekonstrueerimisel tähisega :math:`U`. Lagrange interpoleerimise valem ütleb:

.. math::
    \overline{P}(x) = \sum\limits_{j \in U} s_j \frac{\prod\limits_{i \in U, j \neq i}x-x_i}{\prod\limits_{i \in U, j \neq i}x_j-x_i}

Tõepoolest: fikseerime :math:`j` - paneme tähele, et kui :math:`x = x_j`, siis
murru väärtus on :math:`1` (kuna lugejas ja nimetajas olevad kordajad taandavad
üksteist) ja kui :math:`x \neq x_j`, kuid :math:`x = x_k`, mingi muu :math:`k
\in U` korral, siis murd on :math:`0` (kuna lugejas on :math:`x_k - x_i = 0`
mingi :math:`i \in U` korral). Seega:

.. math::
    \overline{P}(x_j) = s_j + 0 \sum_{i \in U, i \neq j} s_i = s_j = P(x_j)

Kuna osapooled teavad väärtuseid :math:`s_i = P(x_i)` (osakud), siis
kombineerides ning korrutades need läbi baaspolünoomiga

.. math::
    L(U,x,j) = \frac{\prod\limits_{i \in U, j \neq i}x - x_i}{\prod\limits_{i \in U, j \neq i}x_j - x_i}

ja fikseerides :math:`x = 0`, saame jagatud saladuse.

`ee.ivxv.key.protocol.generation.desmedt.DesmedtGeneration`
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Arvestades, et ElGamali võtmeparameetriks on rühm :math:`G` koos generaatoriga
:math:`g`, siis salajaseks võtmeks valitakse :math:`x`, mis on ülimalt
:math:`ord(g)`, st. :math:`g` multiplikatiivne järk rühmas :math:`G`. Vastavaks
avalikuks võtmeks võetakse väärtus :math:`y = g^x`. Salajase võtme komplektiks
saab väärtus :math:`(G, g, x)` ja avaliku võtme komplektiks väärtus :math:`(G,
g, y)`. Rühm :math:`G` valitakse selliselt, et tema järk on mingi algarv
:math:`p` selliselt, et kehtib :math:`p = 2  q + 1`, kus :math:`q` on samuti
algarv. Selliselt juhul kirjeldab :math:`G` väärtust algarv :math:`p`.

Algebrast teame, et kui :math:`G` järk on :math:`2q + 1`, siis iga selle rühma
elemendi järk on kas :math:`1`, :math:`2`, :math:`q` või :math:`2q`. Me oleme
huvitatud selliselt generaatorist, mille järk on :math:`q` ja mis on ruutjääk,
kuna see genereerib piisavalt suure alamrühma, mille kõik elemendid on
ruutjäägid. Vastasel juhul võib toimuda ühe biti lekkimine krüpteeritud sõnumi
kohta. Sellise generaatori leidmiseks vaatame me rühma suvalisi elemente ning
kontrollime tema järku ning ruutjäägilisust kuni leiame sobiva elemendi. Sellise
elemendi määrame generaatoriks.

Instants genereerib juhusliku :math:`0<x<q` salajaseks võtmeks, jagab selle
Shamiri ühissalastuse abil argumentidena antud osapoolte vahel. Iga salajase
võtme osak kodeeritakse kui ühissalastamata salajase võtme komplekt.

Seejärel arvutatakse :math:`y = g^x` ning tagastatakse kodeeritud avaliku võtme
komplekt.

`ee.ivxv.key.protocol.generation.shoup.ShoupGeneration`
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

RSA võtmepaar genereeritakse järgnevalt: genereeritakse kaks algarvu :math:`p`
ja :math:`q` bitipikkusega :math:`\textit{modLen}/2` ning võetakse :math:`n =
pq`. Avalik võti :math:`e` võetakse selliselt et :math:`\gcd(e, \phi(n)) = 1`,
kuid antud protokollis on :math:`e` fikseeritud :math:`e=65537`. Seega tuleb
valida :math:`p` ja :math:`q` nii pikalt kui see tingimus kehtib. Salajane võti
:math:`d` võetakse selliselt, et :math:`de \equiv 1 \pmod{\phi(n)}`, kus
:math:`\phi` on Euleri :math:`\phi`.

Arv :math:`\phi(n)` näitab, kui paljud arvudest :math:`1 \leq m < n` on sellised
et :math:`\gcd(m,n) = 1`, kus :math:`\gcd(a,b)` on kahe arvu :math:`a` ja
:math:`b` suurim ühistegur. On ilmne, et kui :math:`p` on algarv, siis
:math:`\phi(p) = p-1`. Lisaks on lihtne näidata, et kui :math:`p` ja :math:`q`
on algarvud, siis :math:`\phi(pq) = \phi(p)\phi(q)`.

Euleri teoreem ütleb, et kui :math:`a` ja :math:`n` on ühistegurita, siis:

.. math::
    a^{\phi(n)} \equiv 1 \pmod{n}

Seega, kui sõnumi :math:`m` allkirjastamiseks tehakse :math:`s \equiv m^d
\pmod{n}`, siis verifitseerimiseks kontrollitakse kas :math:`s^e \equiv m
\pmod{n}`. Tõepoolest: :math:`(m^d)^e \equiv m^{de} \equiv m^{k\phi(n)+1} \equiv
m^{k\phi(n)}m \equiv 1^km \equiv m \pmod{n}`.

Salajane võti :math:`d` jagatakse Shamiri salastuse jagamisega osadeks, iga osa
kodeeritakse kui jagamata salajase võtme komponent ning salvestatakse
osapoolele. Avalik võtme komplekt kodeeritakse ning tagastatakse.

`ee.ivxv.key.protocol.decryption.recover.RecoverDecryption`
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Protokoll toimib, rekonstrueerides ElGamali võtme ning dekrüpteerides sellega
krüptogramme.

Täpsemalt, olgu :math:`U` indeksid kaartidest, mis moodustavad argumendiks antud
:math:`\mathit{cards}` muutuja. Instants loeb salajase võtme komplektid
kaartidelt, kontrollib võtmekomplektide terviklust (st. rühma :math:`G` ja
generaatori :math:`g` kirjelduse ühesust), dekodeerib igast komplektist salajase
võtme :math:`x_i`.

Seejärel arvutatakse salajane võti :math:`x` kasutades Lagrange
interpoleerimist:

.. math::
    x = P(0) = \sum\limits_{j\in U} s_j \frac{\prod\limits_{i\in U, j \neq i} -x_i}{\prod\limits_{i\in U, j \neq i} x_j-x_i}

Krüptogrammi :math:`c=(c_1,c_2)=(my^r,g^r)` dekrüpteerimiseks arvutatakse:

.. math::
    d = \frac{c_1}{c_{2}^x}

Dekrüpteerimise lugemistõendi jaoks valitakse juhuslik :math:`r` ning
konstrueeritakse järgnevad pühendumused:

.. math::
    a = c_{2}^r \\
    b = g^r

Seejärel arvutatakse Fiat-Shamiri pretensioon järgnevalt, kus `H` on
räsifunktsioon `SHA2-256` ning `B2I` on meetod, mis teisendab baidijada
täisarvuks ühtlaselt vahemikus::

    K = H("DECRYPTION" || y || c || d || a || b)
    k = B2I(K, q)

Nüüd arvutatakse lugemistõendi vastus:

.. math::
    s = kx + r

Kogu lugemistõend on komplekt :math:`(a,b,s)`. Tagastatakse :math:`(d,(a,b,s))`.

`ee.ivxv.key.protocol.signing.shoup.ShoupSigning`
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Antud protokollis ei toimu võtme rekonstrueerimist.

Olgu :math:`U` indeksid kaartidest, mis moodustavad argumendiks antud
:math:`\mathit{cards}` muutuja. Klassiinstants loeb salajase võtme komplektid ja
kontrollib nende terviklust (st. mooduli ja avaliku võtme ühesus). Loetakse
mällu võtme moodul :math:`n` ja avalik võti :math:`e`. Lisaks dekodeeritakse ja
loetakse mällu salajased võtmed :math:`d_i`. Allkirja genereerimiseks sõnumile
:math:`m` rakendatakse sellele EMSA-PSS kodeerimist [RFC8017]_, kus on kasutusel
varasemalt defineeritud `RSA-PSS parameetrid`_, saades allkirjastamiseks sõnumi
:math:`M`.

Me tähistame tähisega :math:`n!` arvu :math:`n` faktoriaali, st. :math:`n! = 1
\cdot 2 \cdot 3 \cdot \ldots \cdot n`. Meenutame, et Lagrange interpolatsiooni
baaspolünoom oli:

.. math::
    L(U,x,j) = \frac{\prod\limits_{i\in U, j \neq i} x-x_i}{\prod\limits_{i\in U, j \neq i} x_j-x_i}

Defineerime modifitseeritud Lagrange baaspolünoomi järgnevalt:

.. math::
    L'(U,x,j) = n! \frac{\prod\limits_{i\in U, j \neq i} x-x_i}{\prod\limits_{i\in U, j \neq i} x_j- x_i}

Kuna me teame, et punktid :math:`1 \leq x_i,x_j \leq n`, siis
:math:`|x_j-x_i|<n`. Seega, korrutades Lagrange baaspolünoomi läbi :math:`n!`,
saame, et :math:`L'(U,j)` on alati täisarv.

.. warning: Kuna kehtib `|k|=|-k|`, siis võib mingitel juhtudel faktoriaalist
   tegurid ära taandada ja saada murdarvu. Me oleme eksperimentaalselt
   kontrollinud kõiki juhte kuni 15 osapoolega skeemideni ning siis ei teki
   murdarvu. Rohkemate osapoolte korral tuleb kontrollida murrulisust ja
   vajadusel muuta protokolli.

Allkirja konstrueerimiseks arvutame:

.. math::
    s = \prod\limits_{j\in U} {(M^{x_j})}^{L'(U,0,j)} = M^{\sum\limits_{j\in U} x_j L'(U,0,j)} = M^{n!d}

Kuna kasutasime modifitseeritud Lagrange interpoleerimist, siis võrreldes
tavalise RSA allkirjaga on see astendatud :math:`n!`-ga. Bezout' lemmast teame,
et :math:`x` ja :math:`y` korral leiduvad :math:`a` ja :math:`b` selliselt, et
:math:`ax+by=\gcd(x,y)`. Veel enam, selliseid :math:`a` ja :math:`b` väärtuseid
on võimalik leida laiendatud Eukleidese algoritmiga suurima ühisteguri
leidmiseks.

Kasutades Eukleidese laiendatud algoritmi, leitakse :math:`a` ja :math:`b`,
selliselt et :math:`ae+bn!=\gcd(e,n!)`. Kuna avalik võti :math:`e` on valitud
algarv, siis :math:`\gcd(e,n!)=1`. Arvutame:

.. math::
    \sigma = M^as^b

Arvestades, et :math:`de = 1 \pmod{\phi(n)}`, on see tõesti korrektne allkiri:

.. math::
    \sigma^e &= M^{ae}s^{be}      \\
             &= M^{ae}M^{n!dbe} \\
             &= M^{ae}M^{n!bde} \\
             &= M^{ae}M^{n!b}     \\
             &= M^{ae+bn!}         \\
             &= M^{\gcd(e,n!)}       \\
             &= M

Protokolli instants tagastab :math:`\sigma` allkirjana.
