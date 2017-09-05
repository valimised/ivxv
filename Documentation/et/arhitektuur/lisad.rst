..  IVXV arhitektuur

Lisad
=====

Kogumisteenuse ehitamine paigatud Go standardteegiga
----------------------------------------------------

Eestis on ringluses mitmeid ID-kaarte ja Digi-ID-sid, mille sertifikaadid
sisaldavad valesti kodeeritud RSA avalikku võtit. Go standardteek keeldub
sellised vigaseid sertifikaate vastu võtmast. Samas on nende arv liiga suur,
et selle vastu mitte midagi ette võtta.

Lahenduseks tuleb IVXV kogumisteenuse alamteenused kompileerida kasutades
paigatud Go standardteeki. Tarnega on kaasas ``ivxv-golang`` pakk, mis
sisaldab paikasid sellist laadi vigaste sertifikaatide lubamiseks ja vahendeid
nende rakendamiseks.

Paigatud standardteegi ehitamine peaks toimuma samas keskkonnas, kus ka IVXV
kogumisteenus ehitatakse, st Ubuntu 16.04.

Esimese asjana tuleb paigaldada kõik ``ivxv-golang`` kaustas olevas
``README.rst`` failis loetletud sõltuvused. Seejärel ``make`` käsu andmisel
laetakse Ubuntu hoidlatest alla kõige uuem Go 1.7 lähtekood, paigatakse,
ehitatakse ning testitakse vigaste sertifikaatide kasutust. Õnnestumise korral
on ``source/`` alamkaustas muuhulgas kaks vajalikku .deb pakki:

- ``golang-1.7-src_1.7.1-2ubuntu1_amd64.deb`` ja
- ``golang-1.7-go_1.7.1-2ubuntu1_amd64.deb``.

Need tuleb paigaldada IVXV kogumisteenust ehitavasse arvutisse enne IVXV
kompileerimist: siis kasutatakse valmendamise käigus paigatud Go
standardteeki.
