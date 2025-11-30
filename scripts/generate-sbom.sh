#!/usr/bin/env bash
# Generate Software Bill of Materials (SBOM) in SPDX and CycloneDX formats
# SBOM provides transparency into software components for security and compliance
set -eo pipefail

# Source standard output functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/check-output.sh
source "${SCRIPT_DIR}/lib/check-output.sh"

# Output directory
OUTPUT_DIR="${SBOM_OUTPUT_DIR:-reports/sbom}"
PROJECT_NAME="${PROJECT_NAME:-ckeletin-go}"
VERSION="${VERSION:-$(git describe --tags --always 2>/dev/null || echo 'dev')}"

# Format selection
FORMAT="${1:-all}"  # all, spdx, cyclonedx

usage() {
    echo "Usage: $0 [format]"
    echo ""
    echo "Formats:"
    echo "  all        Generate both SPDX and CycloneDX (default)"
    echo "  spdx       Generate SPDX format only"
    echo "  cyclonedx  Generate CycloneDX format only"
    echo ""
    echo "Environment variables:"
    echo "  SBOM_OUTPUT_DIR  Output directory (default: reports/sbom)"
    echo "  PROJECT_NAME     Project name in SBOM (default: ckeletin-go)"
    echo "  VERSION          Version to include (default: git describe)"
    echo ""
    echo "Examples:"
    echo "  $0              # Generate all formats"
    echo "  $0 spdx         # Generate SPDX only"
    echo "  $0 cyclonedx    # Generate CycloneDX only"
}

# Check for syft installation
check_syft() {
    if ! command -v syft &> /dev/null; then
        echo "Error: syft is not installed"
        echo ""
        echo "Install with:"
        echo "  brew install syft                    # macOS"
        echo "  curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sh -s -- -b /usr/local/bin"
        echo ""
        echo "Or run: task setup  # Installs all dev tools"
        exit 1
    fi
}

generate_spdx() {
    echo "Generating SPDX SBOM..."

    # SPDX JSON format (machine-readable, most common)
    syft dir:. \
        --output spdx-json="${OUTPUT_DIR}/${PROJECT_NAME}-${VERSION}.spdx.json" \
        --source-name "${PROJECT_NAME}" \
        --source-version "${VERSION}" \
        2>/dev/null

    echo "  Created: ${OUTPUT_DIR}/${PROJECT_NAME}-${VERSION}.spdx.json"

    # Also generate human-readable tag-value format
    syft dir:. \
        --output spdx="${OUTPUT_DIR}/${PROJECT_NAME}-${VERSION}.spdx" \
        --source-name "${PROJECT_NAME}" \
        --source-version "${VERSION}" \
        2>/dev/null

    echo "  Created: ${OUTPUT_DIR}/${PROJECT_NAME}-${VERSION}.spdx"
}

generate_cyclonedx() {
    echo "Generating CycloneDX SBOM..."

    # CycloneDX JSON format
    syft dir:. \
        --output cyclonedx-json="${OUTPUT_DIR}/${PROJECT_NAME}-${VERSION}.cdx.json" \
        --source-name "${PROJECT_NAME}" \
        --source-version "${VERSION}" \
        2>/dev/null

    echo "  Created: ${OUTPUT_DIR}/${PROJECT_NAME}-${VERSION}.cdx.json"

    # CycloneDX XML format (some tools prefer XML)
    syft dir:. \
        --output cyclonedx-xml="${OUTPUT_DIR}/${PROJECT_NAME}-${VERSION}.cdx.xml" \
        --source-name "${PROJECT_NAME}" \
        --source-version "${VERSION}" \
        2>/dev/null

    echo "  Created: ${OUTPUT_DIR}/${PROJECT_NAME}-${VERSION}.cdx.xml"
}

main() {
    case "$FORMAT" in
        -h|--help|help)
            usage
            exit 0
            ;;
        all|spdx|cyclonedx)
            ;;
        *)
            echo "Error: Unknown format '$FORMAT'"
            usage
            exit 1
            ;;
    esac

    check_syft

    check_header "Generating SBOM (Software Bill of Materials)"
    echo "  Project: ${PROJECT_NAME}"
    echo "  Version: ${VERSION}"
    echo "  Output:  ${OUTPUT_DIR}/"
    echo ""

    # Create output directory
    mkdir -p "$OUTPUT_DIR"

    # Generate requested formats
    case "$FORMAT" in
        all)
            generate_spdx
            generate_cyclonedx
            ;;
        spdx)
            generate_spdx
            ;;
        cyclonedx)
            generate_cyclonedx
            ;;
    esac

    echo ""
    check_success "SBOM generation complete"
    echo ""
    echo "Generated files:"
    ls -la "$OUTPUT_DIR/"* 2>/dev/null | awk '{print "  " $9 " (" $5 " bytes)"}'
    echo ""
    echo "Use these SBOMs for:"
    echo "  - Security audits and vulnerability scanning"
    echo "  - License compliance verification"
    echo "  - Supply chain transparency"
    echo "  - Enterprise/government compliance requirements"
}

main "$@"
