#!/bin/bash

# BMKG Scraper Runner Script
# This script sets up and runs the Go-based BMKG scraper in WSL

set -e  # Exit on error

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print colored messages
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}  BMKG Scraper - Data Collection Tool${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
}

# Get the script directory (works in WSL)
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

print_header

# Step 1: Check if Go is installed
print_info "Checking Go installation..."
if ! command -v go &> /dev/null; then
    print_error "Go is not installed!"
    echo ""
    echo "Please install Go 1.21 or higher:"
    echo "  Ubuntu/Debian: sudo apt update && sudo apt install golang-go"
    echo "  Or download from: https://go.dev/dl/"
    echo ""
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}')
print_success "Go is installed: $GO_VERSION"

# Step 2: Check if go.mod exists
print_info "Checking go.mod file..."
if [ ! -f "go.mod" ]; then
    print_error "go.mod not found!"
    echo ""
    echo "Creating go.mod file..."
    cat > go.mod << 'EOF'
module ingestion/scraper-go

go 1.21

require golang.org/x/time v0.5.0
EOF
    print_success "go.mod created"
else
    print_success "go.mod found"
fi

# Step 3: Install/update dependencies
print_info "Installing dependencies..."
if go mod download; then
    print_success "Dependencies downloaded"
else
    print_error "Failed to download dependencies"
    exit 1
fi

print_info "Tidying up dependencies..."
if go mod tidy; then
    print_success "Dependencies tidied"
else
    print_warning "go mod tidy had issues, but continuing..."
fi

# Step 4: Create output directories
print_info "Creating output directories..."

# Convert Windows path to WSL path if needed
if [[ "$SCRIPT_DIR" == /mnt/* ]]; then
    # Already in WSL path format
    DATA_DIR="$SCRIPT_DIR/../../data/raw"
else
    # Might be Windows path, use relative path
    DATA_DIR="../../data/raw"
fi

mkdir -p "$DATA_DIR"
print_success "Output directory ready: $DATA_DIR"

# Step 5: Build the scraper
print_info "Building scraper..."
cd cmd/scraper

if go build -o scraper main.go; then
    print_success "Scraper built successfully"
else
    print_error "Build failed!"
    exit 1
fi

# Step 6: Run the scraper
print_info "Running scraper..."
echo ""
echo -e "${BLUE}----------------------------------------${NC}"
echo -e "${BLUE}  Scraping Data...${NC}"
echo -e "${BLUE}----------------------------------------${NC}"
echo ""

START_TIME=$(date +%s)

if ./scraper; then
    END_TIME=$(date +%s)
    DURATION=$((END_TIME - START_TIME))
    
    echo ""
    echo -e "${BLUE}----------------------------------------${NC}"
    print_success "Scraping completed in ${DURATION} seconds"
    echo -e "${BLUE}----------------------------------------${NC}"
else
    print_error "Scraper execution failed!"
    exit 1
fi

# Step 7: Show output files
echo ""
print_info "Checking output files..."
cd "$SCRIPT_DIR"

OUTPUT_FILES=(
    "../../data/raw/tokopedia.json"
    "../../data/raw/bmkg_weather_forecast.json"
    "../../data/raw/bmkg_latest_earthquake.json"
    "../../data/raw/bmkg_recent_earthquakes.json"
)

echo ""
echo -e "${GREEN}Output Files:${NC}"
echo "----------------------------------------"

FILES_FOUND=0
for file in "${OUTPUT_FILES[@]}"; do
    if [ -f "$file" ]; then
        FILE_SIZE=$(du -h "$file" | cut -f1)
        FILE_NAME=$(basename "$file")
        echo -e "  ${GREEN}✓${NC} $FILE_NAME (${FILE_SIZE})"
        FILES_FOUND=$((FILES_FOUND + 1))
    else
        FILE_NAME=$(basename "$file")
        echo -e "  ${RED}✗${NC} $FILE_NAME (not found)"
    fi
done

echo "----------------------------------------"
echo ""

if [ $FILES_FOUND -eq ${#OUTPUT_FILES[@]} ]; then
    print_success "All $FILES_FOUND output files created successfully!"
elif [ $FILES_FOUND -gt 0 ]; then
    print_warning "$FILES_FOUND of ${#OUTPUT_FILES[@]} files created"
else
    print_error "No output files were created"
    exit 1
fi

# Step 8: Show file locations
echo ""
print_info "Data saved to: $(cd ../../data/raw && pwd)"
echo ""

# Step 9: Summary
print_header
echo -e "${GREEN}Scraping Summary:${NC}"
echo "  • Weather forecast data collected"
echo "  • Latest earthquake data collected"
echo "  • Recent earthquakes data collected"
echo "  • Tokopedia product data collected"
echo ""
echo -e "${BLUE}Next steps:${NC}"
echo "  • Review the JSON files in data/raw/"
echo "  • Process the data with your pipeline"
echo "  • Schedule this script with cron for regular updates"
echo ""
print_success "All done! 🎉"
echo ""
