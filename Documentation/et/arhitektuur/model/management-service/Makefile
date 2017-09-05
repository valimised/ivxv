PLANTUML=env -u DISPLAY plantuml

all: management-service-components.png collector-status.png service-status-changes.png service-status.png upload-command.png

clean:
	rm -f *.png

%.png: %.pu
	$(PLANTUML) $<
