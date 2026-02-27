#!/bin/bash

echo "Starting Core Procurement Service in separate terminals..."

PROJECT_DIR="$(cd "$(dirname "$0")" && pwd)"

echo "Opening auth-identity-service terminal..."
gnome-terminal -- bash -c "cd '$PROJECT_DIR/services/auth-identity-service' && go run main.go; exec bash"

echo "Opening inventory-service terminal..."
gnome-terminal -- bash -c "cd '$PROJECT_DIR/services/inventory-service' && go run main.go; exec bash"

echo "Opening purchase-service terminal..."
gnome-terminal -- bash -c "cd '$PROJECT_DIR/services/purchase-service' && go run main.go; exec bash"

echo "Opening approval-service terminal..."
gnome-terminal -- bash -c "cd '$PROJECT_DIR/services/approval-service' && go run main.go; exec bash"

echo ""
