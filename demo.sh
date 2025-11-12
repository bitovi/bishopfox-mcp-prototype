#!/bin/bash

# Demo script for the MCP TUI interface
# This script demonstrates the enhanced UI capabilities

echo "🚀 BishopFox MCP Prototype - TUI Demo"
echo "===================================="
echo ""
echo "This demo showcases the enhanced terminal user interface built with Charm Bracelet libraries."
echo ""
echo "Features demonstrated:"
echo "  🎨 Beautiful colors and styling"
echo "  💬 Interactive chat interface"
echo "  ⚡ Real-time loading animations"
echo "  🔄 Session management"
echo "  📱 Responsive design"
echo ""
echo "Prerequisites:"
echo "  ✅ Docker containers running (docker compose up)"
echo "  ✅ Server accessible at http://localhost:8100"
echo ""

# Check if server is running
echo "🔍 Checking server status..."
if curl -s http://localhost:8100/health >/dev/null 2>&1; then
    echo "✅ Server is running!"
else
    echo "❌ Server not responding. Please run 'docker compose up' first."
    echo ""
    echo "To start the demo:"
    echo "  1. Run: docker compose up"
    echo "  2. Wait for services to start"
    echo "  3. Run: ./demo.sh"
    exit 1
fi

echo ""
echo "🎯 Starting TUI interface..."
echo ""
echo "Try asking questions like:"
echo "  • 'What assets do we have?'"
echo "  • 'Show me critical vulnerabilities'"
echo "  • 'List all Windows servers'"
echo ""
echo "Press Ctrl+C to exit when you're done exploring!"
echo ""
echo "Starting in 3 seconds..."
sleep 1
echo "2..."
sleep 1
echo "1..."
sleep 1

# Launch the TUI
./querier-tui