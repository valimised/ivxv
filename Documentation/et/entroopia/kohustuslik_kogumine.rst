Kohustusliku entroopia kogumine
===============================

Seadistades erinevad entroopiaallikad võtmerakenduses, lähtestatakse
vastavalt allika tüübile vastav implementatsioon, mis tagab entroopia
tagastamise vastavalt spetsifikatsioonile. Vajades rakenduses entroopiat,
loetakse defineeritud allikatest sisend, kombineeritakse see (vt. peatükki
`Entroopia kogumine`_) krüptograafilise kombineerijaga ning kasutatakse
kombineerijast tulevat väljundit.

Vaikimisi loetakse entroopiaallikatest sisendit mitteblokeeruvalt, st. kui
allikast ei ole võimalik saada nõutud arv baite, tagastatakse saadaolev arv
baite. Nii võib tekkida olukord, kus rakenduse kasutamisel vahetult pärast
süsteemi käivitamist pole saadaval piisaval palju entroopiat, kuid kuna
kombineerija on alati mitteblokeeruv, siis genereeritakse krüptomaterjal selle
vähese entroopia põhjal.

Täiendavalt on olemas kombineerija blokeeruv variant, mis garanteerib et
entroopiaallikatest on reaalselt saadud nii palju entroopiat kui kombineerijast
küsitakse. Samas, selle kombineerija variandi kasutamisel on probleemiks, et
vaja võib minna massiivne kogus entroopiat, mis võib takistada rakenduste
normaalset tööd.

Kuna kombineerija kasutab väljundi genereerimiseks kogu eelnevat
entroopiaallikatest saadud sisendit, siis on võimalik kasutada võtmerakenduse
töös kombineeritud lähenemist. Kombineeritud lähenemise korral loetakse
võtmerakenduse töö alguses blokeeruvalt piisavalt palju entroopiat, et saavutada
nõutav turvatase. Hilisema töö käigus loetakse allikatest entroopiat
mitteblokeeruvalt.

Tehnilise poole pealt vastavad entroopiaallikad `ee.ivxv.crypto.rnd.Rnd`
liidesele, selles liideses on `read` mitteblokeeruv ja `mustRead` blokeeruv
meetod. Lähtestades võtmegenereerimise protokolli või kutsudes välja
protkolli meetodi `generateKey()` (vt. peatükki `Eritüübiliste võtmeosakute
protokollide liidestamine`_), peab protokolli implementatsioon vajaduse korral
kutsuma välja meetodi `Rnd.mustRead`, et garanteerida piisava entroopia
olemasolu.
