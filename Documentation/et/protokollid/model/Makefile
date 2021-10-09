PLANTUML=env -u DISPLAY plantuml

all: phase1.png phase2.png phase3.png

phase1.png: sequence.pu
	$(PLANTUML) -DPHASE_1 $<
	mv sequence.png $@

phase2.png: sequence.pu
	$(PLANTUML) -DPHASE_2 $<
	mv sequence.png $@

phase3.png: sequence.pu
	$(PLANTUML) -DPHASE_3 $<
	mv sequence.png $@

%.png: %.pu
	$(PLANTUML) $<
