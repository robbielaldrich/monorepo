#!/usr/bin/env bash
# 4x AI upscale of a card scan. The anime model gives the crispest text/outlines on card art.
# usage: ./upscale.sh card.jpg card_4x.png [model]
set -euo pipefail
here="$(cd "$(dirname "$0")" && pwd)"
"$here/tools/realesrgan/realesrgan-ncnn-vulkan" -i "$1" -o "$2" -m "$here/tools/realesrgan/models" -n "${3:-realesrgan-x4plus-anime}" -s 4
