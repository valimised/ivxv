
ASTYLE=astyle
WINDRES=windres
CXX = g++
CC=gcc
CXX_FLAGS = -g -D_WIN32_IE=0x0900 -Wfatal-errors -Wall -Wextra -Wpedantic -Wconversion -Wshadow
CFLAGS = -D_WIN32_IE=0x0900 -Wfatal-errors -Wall -Wextra -Wpedantic -Wconversion -Wshadow

SERVER_BIN = server.exe
CLIENT_BIN = client.exe

BUILD_DIR = ./build

CPP = Aggregator.cpp Counter.cpp CounterManager.cpp Error.cpp Network.cpp\
	  Protocol.cpp server.cpp Slotter.cpp

CLIC = client.c

RES = resources.rc

OBJ = $(CPP:%.cpp=$(BUILD_DIR)/%.o)
CLIOBJ = $(CLIC:%.c=$(BUILD_DIR)/%.o)
RESOBJ = $(RES:%.rc=$(BUILD_DIR)/%.o)

LDFLAGS = -static-libgcc -static-libstdc++ -g -lcomctl32 -lws2_32 -Wl,--subsystem,console
CLILDFLAGS = -s -lws2_32 -Wl,--subsystem,console

DEP = $(OBJ:%.o=%.d)

all: $(SERVER_BIN) $(CLIENT_BIN)

$(SERVER_BIN) : $(BUILD_DIR)/$(SERVER_BIN)

$(CLIENT_BIN) : $(BUILD_DIR)/$(CLIENT_BIN)

$(BUILD_DIR)/$(SERVER_BIN) : $(OBJ) $(RESOBJ)
	@echo [Building]: $@
	@mkdir -p $(@D)
	@$(CXX) $(CXX_FLAGS) $^ -o $@ $(LDFLAGS)

$(BUILD_DIR)/$(CLIENT_BIN) : $(CLIOBJ)
	@echo [Building]: $@
	@mkdir -p $(@D)
	@$(CC) $(C_FLAGS) $^ -o $@ $(CLILDFLAGS)

-include $(DEP)

$(BUILD_DIR)/%.o: %.rc
	@echo [Compiling resources]: $<
	@mkdir -p $(@D)
	@$(WINDRES) -o $@ -i $<

$(BUILD_DIR)/%.o : %.cpp
	@echo [Compiling CXX]: $<
	@mkdir -p $(@D)
	@$(CXX) $(CXX_FLAGS) -MMD -c $< -o $@

$(BUILD_DIR)/%.o: %.c
	@echo [Compiling C]: $<
	@mkdir -p $(@D)
	@$(CC) $(CFLAGS) -o $@ -c $<

.PHONY : clean
clean :
	@echo [Cleaning]
	@rm -f $(BUILD_DIR)/$(SERVER_BIN) $(OBJ) $(DEP) $(RESOBJ)\
		$(BUILD_DIR)/$(CLIENT_BIN) $(CLIOBJ)

astyle:
	$(ASTYLE) --options=astylerc *.c *.cpp *.h

