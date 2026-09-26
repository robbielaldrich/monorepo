#!/usr/bin/env bash
# One-time setup: Python venv + Real-ESRGAN (ncnn-vulkan build, macOS universal binary).
set -euo pipefail
cd "$(dirname "$0")"

python3 -m venv .venv
.venv/bin/pip install -q -r requirements.txt

if [ ! -x tools/realesrgan/realesrgan-ncnn-vulkan ]; then
  mkdir -p tools/realesrgan
  curl -sSL -o tools/realesrgan.zip \
    https://github.com/xinntao/Real-ESRGAN/releases/download/v0.2.5.0/realesrgan-ncnn-vulkan-20220424-macos.zip
  unzip -oq tools/realesrgan.zip -d tools/realesrgan
  rm tools/realesrgan.zip
  chmod +x tools/realesrgan/realesrgan-ncnn-vulkan
fi
echo "ready"
