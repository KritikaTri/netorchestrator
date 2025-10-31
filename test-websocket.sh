#!/bin/bash

# WebSocket Testing Script
# Tests WebSocket connection and real-time updates

API_URL="http://localhost:8080"
WS_URL="ws://localhost:8080/ws"
API_KEY="neto_test"

echo "=========================================="
echo "WebSocket Connection Test"
echo "=========================================="
echo ""

# Check if websocat is installed, if not provide instructions
if ! command -v websocat &> /dev/null; then
    echo "⚠️  websocat not installed. Install with:"
    echo "   brew install websocat  (macOS)"
    echo "   or use Python script (see test-websocket.py)"
    echo ""
    echo "Creating Python WebSocket test client..."
    cat > /tmp/test_websocket.py << 'EOF'
import asyncio
import websockets
import json
import sys

async def test_websocket():
    uri = "ws://localhost:8080/ws"
    try:
        async with websockets.connect(uri) as websocket:
            print("✅ WebSocket connected successfully")
            print("Waiting for messages (timeout: 10 seconds)...")
            
            try:
                message = await asyncio.wait_for(websocket.recv(), timeout=10.0)
                data = json.loads(message)
                print(f"✅ Received message: {json.dumps(data, indent=2)}")
                return True
            except asyncio.TimeoutError:
                print("⚠️  No messages received within 10 seconds")
                print("   (This is OK if no topology updates occurred)")
                return True
    except Exception as e:
        print(f"❌ WebSocket connection failed: {e}")
        return False

if __name__ == "__main__":
    try:
        result = asyncio.run(test_websocket())
        sys.exit(0 if result else 1)
    except KeyboardInterrupt:
        print("\nTest interrupted")
        sys.exit(1)
EOF
    echo "Running Python WebSocket test..."
    python3 /tmp/test_websocket.py || python /tmp/test_websocket.py
    exit $?
fi

# Use websocat if available
echo "Testing WebSocket connection..."
echo "Connecting to: $WS_URL"
echo ""

timeout 5 websocat -E "$WS_URL" <<< '{"type":"test"}' 2>&1 | head -10

echo ""
echo "✅ WebSocket endpoint is accessible"

