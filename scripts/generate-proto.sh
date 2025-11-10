#!/bin/bash

set -e

echo "=================================="
echo "Device Bridge - Proto Generation"
echo "=================================="
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if protoc is installed
if ! command -v protoc &> /dev/null; then
    echo -e "${RED}Error: protoc not found${NC}"
    echo ""
    echo "Please install Protocol Buffer Compiler:"
    echo ""
    echo "  macOS:   brew install protobuf"
    echo "  Ubuntu:  sudo apt-get install -y protobuf-compiler"
    echo "  Windows: Download from https://github.com/protocolbuffers/protobuf/releases"
    echo ""
    exit 1
fi

# Check protoc version
PROTOC_VERSION=$(protoc --version | awk '{print $2}')
echo -e "${BLUE}Found protoc version: ${PROTOC_VERSION}${NC}"

# Check Go plugins
echo ""
echo "Checking Go protobuf plugins..."

PLUGINS_OK=true

if ! command -v protoc-gen-go &> /dev/null; then
    echo -e "${YELLOW}Warning: protoc-gen-go not found${NC}"
    PLUGINS_OK=false
fi

if ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo -e "${YELLOW}Warning: protoc-gen-go-grpc not found${NC}"
    PLUGINS_OK=false
fi

if ! command -v protoc-gen-grpc-gateway &> /dev/null; then
    echo -e "${YELLOW}Warning: protoc-gen-grpc-gateway not found${NC}"
    PLUGINS_OK=false
fi

if ! command -v protoc-gen-openapiv2 &> /dev/null; then
    echo -e "${YELLOW}Warning: protoc-gen-openapiv2 not found${NC}"
    PLUGINS_OK=false
fi

if [ "$PLUGINS_OK" = false ]; then
    echo ""
    echo "Installing Go protobuf plugins..."
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
    go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest

    # Add to PATH
    export PATH="$PATH:$(go env GOPATH)/bin"

    echo -e "${GREEN}Plugins installed${NC}"
fi

# Generate protobuf code
echo ""
echo -e "${BLUE}Generating protobuf code...${NC}"

cd "$(dirname "$0")/.."

# Create output directory for OpenAPI
mkdir -p docs/api

# Find protoc include path
PROTOC_INCLUDE=""
if [ -d "$HOME/.local/include" ]; then
    PROTOC_INCLUDE="-I$HOME/.local/include"
elif [ -d "/usr/local/include" ]; then
    PROTOC_INCLUDE="-I/usr/local/include"
elif [ -d "/usr/include" ]; then
    PROTOC_INCLUDE="-I/usr/include"
fi

protoc \
    -I. \
    $PROTOC_INCLUDE \
    --go_out=. \
    --go_opt=paths=source_relative \
    --go-grpc_out=. \
    --go-grpc_opt=paths=source_relative \
    --grpc-gateway_out=. \
    --grpc-gateway_opt=paths=source_relative \
    --grpc-gateway_opt=generate_unbound_methods=true \
    --openapiv2_out=docs/api \
    --openapiv2_opt=allow_merge=true \
    --openapiv2_opt=merge_file_name=devicebridge \
    proto/devicebridge/v1/*.proto

if [ $? -eq 0 ]; then
    echo ""
    echo -e "${GREEN}✓ Protobuf code generated successfully${NC}"
    echo ""
    echo "Generated files:"
    echo "  Proto files:"
    ls -lh proto/devicebridge/v1/*.pb.go 2>/dev/null || echo "    (files will appear after first successful generation)"
    echo ""
    echo "  OpenAPI spec:"
    ls -lh docs/api/*.swagger.json 2>/dev/null || echo "    (file will appear after first successful generation)"
    echo ""
else
    echo ""
    echo -e "${RED}✗ Failed to generate protobuf code${NC}"
    exit 1
fi

echo "=================================="
echo -e "${GREEN}Done!${NC}"
echo "=================================="
