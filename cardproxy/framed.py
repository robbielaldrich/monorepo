"""Replace the art of a card whose art sits in a plain rectangular window
(e.g. old-frame Magic cards) with a new image.

The art window is given in the ORIGINAL card's pixel coords (default: the 745x1040
Scryfall PNG of an 8th/Mirrodin-era frame, measured on Mesmeric Orb). If --hires is
given (an AI-upscaled copy of the card), the window is scaled and the art is pasted
at that resolution.

usage: python framed.py CARD ART OUT [--hires CARD_4X.png] [--window 60,120,683,577] [--top .125]
"""
import argparse
from PIL import Image, ImageOps

ap = argparse.ArgumentParser()
ap.add_argument("card"); ap.add_argument("art"); ap.add_argument("out")
ap.add_argument("--hires", help="upscaled copy of CARD to composite at")
ap.add_argument("--window", default="60,120,683,577",
                help="x0,y0,x1,y1 inside the art's 1px black border, in CARD coords")
ap.add_argument("--top", type=float, default=0.125,
                help="for too-tall art: crop top as fraction of art height (0=top)")
args = ap.parse_args()

card = Image.open(args.card)
top_card = Image.open(args.hires) if args.hires else card
if top_card.mode != card.mode:
    top_card = top_card.convert(card.mode)
if args.hires and card.mode == "RGBA":
    # upscalers can soften the transparent rounded corners; reuse the original alpha, scaled
    top_card.putalpha(card.getchannel("A").resize(top_card.size, Image.LANCZOS))
S = top_card.width / card.width
X0, Y0, X1, Y1 = (round(int(v) * S) for v in args.window.split(","))

art = ImageOps.exif_transpose(Image.open(args.art)).convert("RGB")
w, h = X1 - X0, Y1 - Y0
target = w / h
aw, ah = art.size
if aw / ah > target:  # too wide: crop sides, centered
    cw = round(ah * target); left = (aw - cw) // 2
    box = (left, 0, left + cw, ah)
else:  # too tall: crop height starting at --top
    ch = round(aw / target)
    top = min(max(round(ah * args.top), 0), ah - ch)
    box = (0, top, aw, top + ch)
patch = art.crop(box).resize((w, h), Image.LANCZOS)
out = top_card.copy()
out.paste(patch.convert(out.mode), (X0, Y0))
dpi = round(out.width / 2.5)  # standard card is 2.5in wide
out.save(args.out, dpi=(dpi, dpi))
print(f"crop {box} -> saved {args.out} {out.size} @ {dpi} dpi (2.5x3.5in)")
