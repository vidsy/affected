#
# Install installs the tool
#

INSTALLFLAGS ?=
INSTALLCMD   ?= install

ifdef BINOUT
install: INSTALLFLAGS += -o $(BINOUT)
install: INSTALLCMD = build
endif
install:
	@go $(INSTALLCMD) $(INSTALLFLAGS) ./cmd/affected

build: INSTALLFLAGS += -o ./affected
build: INSTALLCMD = build
build: install

#
# Run affected
#

TOOLFLAGS ?=
ifdef FORMAT
run: TOOLFLAGS += -f $(FORMAT)
endif
ifdef COMMITA
run: TOOLFLAGS += -c $(COMMITA)
endif
ifdef COMMITB
run: TOOLFLAGS += -c $(COMMITB)
endif
run:
	@go run ./cmd/affected $(TOOLFLAGS)
