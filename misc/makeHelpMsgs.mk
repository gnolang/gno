BLANK :=
SPACE := $(BLANK) $(BLANK)
HASH  := \#
COMMA := ,

DIR_OFFSET_OPT                    = $(if $(filter $(PWD),$(CURDIR)),,-C $(patsubst $(patsubst %/,%,$(PWD))/%,%,$(CURDIR)))
MAKE_SUBDIRS                      = $(patsubst %/Makefile,%,$(wildcard */Makefile))
MAX_LIST_CHARS                    = $(lastword $(sort $(foreach d,$(1),$(shell echo $(d) | sed -e 's/././g'))))
BASH_GET_TARGET_LINES             = cat Makefile | grep '^[a-z][^:]*:' | grep -v '$(HASH).*@LEGACY'
MAX_TARGET_CHARS                  = $(lastword $(sort $(shell $(BASH_GET_TARGET_LINES) | sed -e 's/:.*$$//' $(if $(1),-e 's/%/$(call MAX_LIST_CHARS,$(1))/',) -e 's/././g')))
SED_EXTRACT_TARGET_AND_COMMENT    = $\
    -e 's/:[^$(HASH)]*$(HASH) */$(HASH) /' $\
    -e 's/:[^$(HASH)]*$$//' $\
    -e 's/$(HASH)/$(subst .,$(SPACE),$(call MAX_TARGET_CHARS,$(1)))   $(HASH)/' $\
    -e 's/^\($(call MAX_TARGET_CHARS,$(1))...\) *$(HASH)/\1<--/' $\
    -e 's/^/  /'
BASH_DISPLAY_TARGETS_AND_COMMENTS = $\
    ( $\
        $(BASH_GET_TARGET_LINES) | $\
            $(if $(1),grep -v '%' |,) $\
            sed $\
                $(call SED_EXTRACT_TARGET_AND_COMMENT,$(1)) $(if $(1),; $\
        for d in $(patsubst %/,%,$(1)) ; do $\
            desc="$$( $\
                head -1 $$d/README.md | $\
                    sed -E $\
                        -e 's/^ *$(HASH)$(HASH)* *//' $\
                        -e 's/^ *('"$$d"'|`'"$$d"'`) *(--*|:) *//' $\
                        -e 's/^(..*)$$/ (\1)/' || $\
                echo > /dev/null $\
            )" ; $\
            $(BASH_GET_TARGET_LINES) | $\
                grep '\%' | $\
                sed $\
                    -e 's$(COMMA)\%$(COMMA)'"$$d$(COMMA)g" $\
                    $(call SED_EXTRACT_TARGET_AND_COMMENT,$(1)) $\
                    -e "s/$$/$$desc/" ; $\
        done,) $\
    ) | $\
    sort
BASH_DISPLAY_SUB_MAKES            = $\
    $(if $\
        $(MAKE_SUBDIRS),$\
        echo ; echo "Sub-directories with make targets:" ; $\
        for d in $(sort $(MAKE_SUBDIRS)); do $\
            echo '    '"$$(grep -q '^help *:' $$d/Makefile && echo '*  ' || echo '   ') $\
                $(if $(filter $(MAKE),$(shell which make)),make,$(MAKE)) $(if $(DIR_OFFSET_OPT),$(DIR_OFFSET_OPT)/,-C$(SPACE))$$d" ; $\
        done ; $\
        grep -q '^help *:' $(patsubst %,%/Makefile,$(MAKE_SUBDIRS)) && $\
            echo && echo '       * Is documented with a `help` target.' || echo > /dev/null,$\
        $(HASH) do nothing $\
    )

