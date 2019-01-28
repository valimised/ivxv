PLANTUML=env -u DISPLAY plantuml

all: example.png processing.png collector_microservices.png \
	collector_extension.png monitoring.png app_modules.png key.png audit.png
	$(MAKE) -C management-service -j

clean:
	$(MAKE) -C management-service $@

collector_extension.png: development.pu
	$(PLANTUML) -DCOLLECTOR_EXTENSION $<
	mv development.png collector_extension.png

collector_microservices.png: development.pu
	$(PLANTUML) -DCOLLECTOR_MICROSERVICES $<
	mv development.png collector_microservices.png

processing.png: development.pu
	$(PLANTUML) -DPROCESSING_APPLICATION $<
	mv development.png processing.png

key.png: development.pu
	$(PLANTUML) -DKEY_APPLICATION $<
	mv development.png key.png

audit.png: development.pu
	$(PLANTUML) -DAUDIT_APPLICATION $<
	mv development.png audit.png

monitoring.png: development.pu
	$(PLANTUML) -DMONITORING $<
	mv development.png monitoring.png

app_modules.png: development.pu
	$(PLANTUML) -DAPP_MODULES $<
	mv development.png app_modules.png

%.png: %.pu
	$(PLANTUML) $<
