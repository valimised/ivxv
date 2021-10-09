PLANTUML=env -u DISPLAY plantuml

all: management-service-components.png collector-status.png service-status-changes.png service-status.png upload-command.png voter-list-status.png

%.png: %.pu
	$(PLANTUML) $<
