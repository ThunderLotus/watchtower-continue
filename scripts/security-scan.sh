#!/bin/bash

# Watchtower 安全扫描脚本
# 用于检查依赖中的安全漏洞和过期依赖

set -e

echo "🔍 开始 Watchtower 安全扫描..."
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 1. 检查 Go 版本
echo -e "${BLUE}📋 检查 Go 版本...${NC}"
go version
echo ""

# 2. 检查依赖漏洞
echo -e "${BLUE}🔒 检查依赖漏洞 (govulncheck)...${NC}"
if ! command -v govulncheck &> /dev/null; then
    echo -e "${YELLOW}安装 govulncheck...${NC}"
    go install golang.org/x/vuln/cmd/govulncheck@latest
fi
govulncheck ./... || echo -e "${RED}发现安全漏洞！${NC}"
echo ""

# 3. 检查代码安全问题
echo -e "${BLUE}🛡️ 检查代码安全问题 (gosec)...${NC}"
if ! command -v gosec &> /dev/null; then
    echo -e "${YELLOW}安装 gosec...${NC}"
    go install github.com/securego/gosec/v2/cmd/gosec@latest
fi
gosec ./... || echo -e "${YELLOW}发现代码安全问题，请检查输出${NC}"
echo ""

# 4. 检查过期依赖
echo -e "${BLUE}📦 检查过期依赖...${NC}"
if ! command -v go-mod-outdated &> /dev/null; then
    echo -e "${YELLOW}安装 go-mod-outdated...${NC}"
    go install github.com/psampaz/go-mod-outdated@latest
fi
go list -u -m -json all | go-mod-outdated -update -direct || echo -e "${YELLOW}有过期依赖需要更新${NC}"
echo ""

# 5. 依赖统计
echo -e "${BLUE}📊 依赖统计${NC}"
TOTAL_DEPS=$(go list -m all | wc -l)
DIRECT_DEPS=$(go list -m -f '{{if not .Indirect}}{{.Path}}{{end}}' all | wc -l)
echo "总依赖数: $TOTAL_DEPS"
echo "直接依赖数: $DIRECT_DEPS"
echo ""

# 6. 检查许可证
echo -e "${BLUE}📄 检查许可证...${NC}"
if ! command -v go-licenses &> /dev/null; then
    echo -e "${YELLOW}安装 go-licenses...${NC}"
    go install github.com/google/go-licenses@latest
fi
echo "生成许可证报告..."
go-licenses save ./... --save_path=./licenses 2>/dev/null || echo -e "${YELLOW}许可证检查完成${NC}"
echo ""

echo -e "${GREEN}✅ 安全扫描完成！${NC}"
echo ""
echo "📝 建议："
echo "1. 定期运行此脚本检查安全问题"
echo "2. 及时更新发现漏洞的依赖"
echo "3. 关注 GitHub Actions 的安全扫描结果"
echo "4. 定期审查第三方依赖的许可证合规性"