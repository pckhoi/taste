SHELL = /bin/bash
.DEFAULT_GOAL = all

GO := go
MD5 := md5
BUILD_DIR := build
MD5_DIR := $(BUILD_DIR)/md5
OS_ARCHS := linux-amd64 darwin-amd64 darwin-arm64
BINARIES := $(foreach osarch,$(OS_ARCHS),$(BUILD_DIR)/$(osarch)/bin/taste)
TAR_FILES := $(foreach osarch,$(OS_ARCHS),$(BUILD_DIR)/$(osarch).tar.gz)

.PHONY: all clean
all: $(TAR_FILES)
clean:
	rm -rf $(BUILD_DIR)

define binary_rule =
echo "\$$(BUILD_DIR)/$(2)-$(1)/bin/taste: \$$(MD5_DIR)/go.sum.md5 taste/VERSION \$$(SOURCES)" >> $(3) && \
echo -e "\t@-mkdir -p \$$(dir \$$@) 2>/dev/null" >> $(3) && \
(if [ "$(2)" == "linux" ]; then \
  echo -e "\tenv GOARCH=amd64 GOOS=linux go build -a -o \$$@ github.com/pckhoi/taste/taste" >> $(3); \
else \
  echo -e "\tGOARCH=$(1) GOOS=$(2) go build -a -o \$$@ github.com/pckhoi/taste/taste" >> $(3); \
fi) && \
echo "" >> $(3)

endef

$(BUILD_DIR)/taste.d: | $(BUILD_DIR)
	echo "SOURCES =" > $@
	echo "$$($(GO) list -deps github.com/pckhoi/taste/taste | \
		grep github.com/pckhoi/taste/ | \
		sed -r -e 's/github.com\/pckhoi\/taste\/(.+)/\1/g' | \
		xargs -n 1 -I {} find {} -maxdepth 1 -name '*.go' \! -name '*_test.go' -print | \
		sed -r -e 's/(.+)/$(subst /,\/,SOURCES += $(MD5_DIR))\/\1.md5/g')" >> $@
	echo "" >> $@
	$(foreach osarch,$(OS_ARCHS),$(call binary_rule,$(word 2,$(subst -, ,$(osarch))),$(word 1,$(subst -, ,$(osarch))),$@))

define license_rule =
$(BUILD_DIR)/$(1)/LICENSE: LICENSE
	cp $$< $$@

endef

define tar_rule =
$(BUILD_DIR)/$(1).tar.gz: $(BUILD_DIR)/$(1)/bin/taste $(BUILD_DIR)/$(1)/LICENSE
	cd $(BUILD_DIR) && \
	tar -czvf $$(notdir $$@) $(1)
endef

# calculate md5
$(MD5_DIR)/%.md5: % | $(MD5_DIR)
	@-mkdir -p $(dir $@) 2>/dev/null
	$(if $(filter-out $(shell cat $@ 2>/dev/null),$(shell $(MD5) $<)),$(MD5) $< > $@)

$(eval $(call program_mk_file_rule,$(prog)))

$(foreach osarch,$(OS_ARCHS),$(eval $(call license_rule,$(osarch))))

$(foreach osarch,$(OS_ARCHS),$(eval $(call tar_rule,$(osarch))))

$(BUILD_DIR): ; @-mkdir $@ 2>/dev/null
$(MD5_DIR): | $(BUILD_DIR) ; @-mkdir $@ 2>/dev/null

include $(BUILD_DIR)/taste.d