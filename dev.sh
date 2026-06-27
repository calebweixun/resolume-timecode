#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
RED='\033[0;31m'
NC='\033[0m'

# Defaults
HTTP_PORT="${HTTP_PORT:-8080}"
OSC_OUT_PORT="${OSC_OUT_PORT:-7001}"
OSC_IN_PORT="${OSC_IN_PORT:-7000}"
OSC_ADDR="${OSC_ADDR:-127.0.0.1}"

# Find Go binary — check common install locations
find_go() {
  for candidate in \
    "$(command -v go 2>/dev/null)" \
    "$HOME/sdk/go/bin/go" \
    "$HOME/.local/share/go/bin/go" \
    /usr/local/go/bin/go \
    /opt/homebrew/bin/go; do
    [ -x "$candidate" ] && echo "$candidate" && return 0
  done
  return 1
}

GO="$(find_go)" || { echo -e "${RED}Error: Go not found. Install Go from https://go.dev/dl/${NC}"; exit 1; }

usage() {
  echo "Usage: $0 [options]"
  echo ""
  echo "Options:"
  echo "  build         Build only, don't run"
  echo "  run           Build and run the app (default)"
  echo "  send-osc      Send a test OSC message to simulate Resolume"
  echo "  open          Open browser to web client"
  echo "  clean         Remove build artifacts"
  echo "  -h, --help    Show this help"
  echo ""
  echo "Environment variables:"
  echo "  HTTP_PORT     HTTP server port (default: 8080)"
  echo "  OSC_OUT_PORT  OSC output port (default: 7001)"
  echo "  OSC_IN_PORT   OSC input port (default: 7000)"
  echo "  OSC_ADDR      OSC host address (default: 127.0.0.1)"
}

build() {
  echo -e "${CYAN}Building resolume-timecode... (using $GO)${NC}"
  "$GO" build -o ./resolume-timecode-dev . 2>&1
  echo -e "${GREEN}Build successful: ./resolume-timecode-dev${NC}"
}

run_app() {
  build
  echo ""
  echo -e "${YELLOW}Starting app...${NC}"
  echo -e "  Web client:  ${CYAN}http://localhost:${HTTP_PORT}${NC}"
  echo -e "  OSC in:      ${CYAN}UDP :${OSC_OUT_PORT}${NC} (receives from Resolume)"
  echo -e "  OSC test:    run ${CYAN}./dev.sh send-osc${NC} in another terminal"
  echo ""
  ./resolume-timecode-dev
}

send_osc() {
  echo -e "${CYAN}Sending test OSC messages to UDP :${OSC_OUT_PORT}...${NC}"
  echo -e "${YELLOW}(simulating Resolume clip at 50% with 60s duration)${NC}"

  # Use Python if available (much simpler for OSC)
  if command -v python3 &>/dev/null; then
    python3 - <<PYEOF
import socket, struct, time

def encode_osc_message(address, *args):
    def pad4(b):
        return b + b'\x00' * ((4 - len(b) % 4) % 4)

    data = pad4(address.encode() + b'\x00')
    type_tag = b',' + b''.join(
        b'f' if isinstance(a, float) else b's' if isinstance(a, str) else b'i'
        for a in args
    ) + b'\x00'
    data += pad4(type_tag)

    for arg in args:
        if isinstance(arg, float):
            data += struct.pack('>f', arg)
        elif isinstance(arg, str):
            data += pad4(arg.encode() + b'\x00')
        elif isinstance(arg, int):
            data += struct.pack('>i', arg)
    return data

sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
port = $OSC_OUT_PORT

print(f"Sending OSC messages to UDP :{port}")
print("Simulating clip: 'Test Clip', 60s duration, counting down from 50%...")
print()

# First send clip name
msg = encode_osc_message("/composition/selectedclip/name", "Test Clip")
sock.sendto(msg, ("127.0.0.1", port))
print("  /composition/selectedclip/name = 'Test Clip'")
time.sleep(0.05)

# Send duration
msg = encode_osc_message("/composition/selectedclip/video/duration", 60.0)
sock.sendto(msg, ("127.0.0.1", port))
print("  /composition/selectedclip/video/duration = 60.0")
time.sleep(0.05)

# Send position updates (simulating playback)
for i in range(20):
    pos = 0.5 + (i * 0.01)  # 50% -> 70%
    msg = encode_osc_message("/composition/selectedclip/video/position", pos)
    sock.sendto(msg, ("127.0.0.1", port))
    print(f"  /composition/selectedclip/video/position = {pos:.3f}  (remaining ~{60*(1-pos):.1f}s)")
    time.sleep(0.1)

print()
print("Done.")
PYEOF
  else
    echo -e "${RED}python3 not found. Install python3 or use a dedicated OSC tool like 'oscsend'.${NC}"
    echo ""
    echo "Manual test with oscsend:"
    echo "  oscsend osc.udp://localhost:${OSC_IN_PORT} /composition/selectedclip/video/position f 0.5"
  fi
}

open_browser() {
  echo -e "${CYAN}Opening http://localhost:${HTTP_PORT} in browser...${NC}"
  if command -v open &>/dev/null; then
    open "http://localhost:${HTTP_PORT}"
  elif command -v xdg-open &>/dev/null; then
    xdg-open "http://localhost:${HTTP_PORT}"
  else
    echo "Visit: http://localhost:${HTTP_PORT}"
  fi
}

clean() {
  rm -f ./resolume-timecode-dev
  echo -e "${GREEN}Cleaned.${NC}"
}

case "${1:-run}" in
  build)      build ;;
  run)        run_app ;;
  send-osc)   send_osc ;;
  open)       open_browser ;;
  clean)      clean ;;
  -h|--help)  usage ;;
  *)
    echo -e "${RED}Unknown command: $1${NC}"
    usage
    exit 1
    ;;
esac
