..  IVXV API

======
Go API
======

Go lähtekoodi API nägemiseks kasutatakse laiendatud standardvahendite hulgas
olevat `godoc <https://godoc.org/golang.org/x/tools/cmd/godoc>`_ tööriista.

Tööriista kasutamise eelduseks on, et keskkonnamuutuja ``GOPATH`` viitab IVXV
Go pakkidele. Kui IVXV juurkaustas käivitada käsk ``make gopath``, siis
trükitakse ekraanile sõne, mille peab seadma selle keskkonnamuutuja väärtuseks.

Seejärel käivitab käsk ``godoc -http=:6060`` kohalikus arvutis veebiserveri,
mis kuulab pordi 6060 peal. Minnes veebilehitsejaga aadressile
http://localhost:6060/pkg/ivxv.ee näeb IVXV Go lähtekoodi dokumentatsiooni.
