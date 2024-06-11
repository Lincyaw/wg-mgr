# 定义目标文件名
TARGET=vpn-tool

# 定义源文件
SRC=.

# 定义编译器
GOCMD=go

# 定义目标操作系统和架构
PLATFORMS=windows/amd64 windows/arm64 linux/amd64 linux/arm64 darwin/amd64 darwin/arm64

# 定义文件后缀
EXTENSIONS=.exe .exe "_linux" "_linux" "_darwin" "_darwin"

# 定义构建目录
BUILDDIR=build

# 默认目标
.PHONY: all
all: $(PLATFORMS)

# 获取文件后缀
EXTENSIONS_LIST := $(foreach ext,$(EXTENSIONS),$(ext))

# 编译每个平台的目标
$(PLATFORMS):
	@mkdir -p $(BUILDDIR)
	OS=$(word 1,$(subst /, ,$@)) && \
	ARCH=$(word 2,$(subst /, ,$@)) && \
	INDEX=$(shell echo $(PLATFORMS) | tr ' ' '\n' | grep -n $@ | cut -d: -f1) && \
	EXT=$$(echo $(EXTENSIONS_LIST) | cut -d' ' -f$$INDEX) && \
	OUTPUT=$(BUILDDIR)/$(TARGET)_$$OS_$$ARCH$$EXT && \
	echo "Building for $$OS/$$ARCH: $$OUTPUT" && \
	GOOS=$$OS GOARCH=$$ARCH $(GOCMD) build -o $$OUTPUT $(SRC)

# 清理构建目录
.PHONY: clean
clean:
	@echo "Cleaning build directory..."
	@rm -rf $(BUILDDIR)

# 帮助信息
.PHONY: help
help:
	@echo "Usage:"
	@echo "  make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all       Build all targets"
	@echo "  clean     Clean build directory"
	@echo "  help      Show this help message"
