"""Replace the full-bleed art of a Sun & Moon TAG TEAM GX card with a photo.

Masks are measured/computed on the original low-res card (717x1000). If --hires is
given (an AI-upscaled copy of the same card), the mask is scaled up and the final
composite is built at that resolution so the photo keeps its detail.

usage: python tagteam.py CARD PHOTO OUT [--hires CARD_4X.png] [--cx .52] [--cy .5] [--zoom 1]
"""
import argparse
import numpy as np
from PIL import Image, ImageOps, ImageDraw, ImageFilter
from scipy import ndimage as ndi

ap = argparse.ArgumentParser()
ap.add_argument("card"); ap.add_argument("photo"); ap.add_argument("out")
ap.add_argument("--hires", help="upscaled copy of CARD to composite at")
ap.add_argument("--cx", type=float, default=0.52, help="horizontal crop center in photo (0..1)")
ap.add_argument("--cy", type=float, default=0.5, help="vertical crop center in photo (0..1)")
ap.add_argument("--zoom", type=float, default=1.0, help=">1 zooms into the photo")
ap.add_argument("--mask-debug", help="write the keep-mask here")
args = ap.parse_args()

card = Image.open(args.card).convert("RGB")
W, H = card.size
a = np.asarray(card).astype(int)
R, G, B = a[..., 0], a[..., 1], a[..., 2]
L = (R * 299 + G * 587 + B * 114) // 1000
sat = a.max(-1) - a.min(-1)

# ---- art window (inside the silver frame), in 717x1000 card coords ----
AX0, AY0, AX1, AY1 = 28, 28, 687, 977

def poly(pts):
    m = Image.new("L", (W, H), 0); ImageDraw.Draw(m).polygon(pts, fill=255)
    return np.asarray(m) > 0

def circle(cx, cy, r):
    m = Image.new("L", (W, H), 0); ImageDraw.Draw(m).ellipse((cx - r, cy - r, cx + r, cy + r), fill=255)
    return np.asarray(m) > 0

def box(x0, y0, x1, y1):
    m = np.zeros((H, W), bool); m[y0:y1, x0:x1] = True; return m

keep = ~box(AX0, AY0, AX1, AY1)
white = (L > 190) & (sat < 70)
white_ish = (L > 200) & (sat < 60)

# frame-yellow (R≈G, low B) is distinct from the art's orange sunset; dilate to grab black outlines
yellow = (R > 170) & (G > 170) & (abs(R - G) < 30) & (B < 110)
yellow = ndi.binary_opening(yellow, iterations=1)
yel = ndi.binary_fill_holes(ndi.binary_dilation(yellow, iterations=3))
keep |= yel & (box(0, 0, 300, 142) | box(200, 890, 717, 1000))   # only banner/rule-box areas

# solid UI shapes
tl = poly([(0, 0), (248, 0), (216, 42), (0, 42)])                   # top-left chevrons
keep |= tl & ((sat < 30) | (L < 70) | yel)
tr = poly([(636, 0), (717, 0), (717, 75), (687, 64), (660, 28)])    # top-right corner
keep |= tr & ((sat < 25) | (L < 70))
keep |= poly([(26, 44), (118, 42), (124, 60), (118, 80), (26, 80)])   # BASIC pill
keep |= poly([(18, 92), (304, 92), (254, 139), (18, 139)])            # TAG TEAM banner
gx = poly([(400, 36), (514, 36), (492, 87), (384, 87)])               # GX logo: cyan/white/dark only
gxm = gx & (((B - R > 40) & (B > 100)) | white_ish | (L < 90))
keep |= ndi.binary_dilation(ndi.binary_fill_holes(ndi.binary_closing(gxm, iterations=2)), iterations=1) & gx
keep |= circle(651, 60, 23)                                           # HP type icon
for cx in (70, 110, 150, 190):                                        # attack costs
    keep |= circle(cx, 617, 21)
keep |= poly([(52, 678), (708, 678), (674, 736), (52, 736), (37, 706)])  # GX attack bar
bar = box(18, 856, 690, 895)                                          # weakness/resistance/retreat
keep |= ndi.binary_opening(bar & (sat < 45), iterations=1) | circle(131, 873, 14) | circle(548, 873, 14) | (bar & (L < 90))
keep |= poly([(226, 960), (284, 898), (700, 898), (700, 980), (226, 980)])  # rule box
keep |= poly([(0, 935), (30, 952), (50, 980), (0, 980)])              # bottom-left corner

# ---- outlined text: dark/blue glyph cores that are ringed by a white outline ----
def text_mask(region, core):
    core = core & region
    lab, n = ndi.label(core, structure=np.ones((3, 3)))
    out = np.zeros_like(core)
    for i, sl in enumerate(ndi.find_objects(lab), 1):
        y0, y1 = max(sl[0].start - 3, 0), sl[0].stop + 3
        x0, x1 = max(sl[1].start - 3, 0), sl[1].stop + 3
        comp = lab[y0:y1, x0:x1] == i
        ring = ndi.binary_dilation(comp, iterations=2) & ~comp
        if ring.any() and white[y0:y1, x0:x1][ring].mean() > 0.55:
            out[y0:y1, x0:x1] |= comp
    return ndi.binary_dilation(out, iterations=3) & region

dark = L < 110
blue = (B - R > 50) & (B > 110)
text = np.zeros((H, W), bool)
for rg in [box(122, 36, 400, 92), box(410, 36, 632, 94),       # name, HP
           box(200, 590, 460, 648), box(585, 594, 670, 645),   # attack name, damage
           box(50, 638, 470, 675),                             # attack text
           box(36, 896, 225, 978)]:                            # illus / number / copyright
    text |= text_mask(rg, dark)
text |= text_mask(box(45, 738, 672, 852), blue | dark)         # GX attack text
sym = box(40, 925, 170, 958) & (white | dark)                  # set symbol + star
text |= ndi.binary_dilation(sym & ndi.binary_dilation(white & box(40, 925, 170, 958), iterations=2), iterations=1)
keep |= text

if args.mask_debug:
    Image.fromarray((keep * 255).astype(np.uint8)).save(args.mask_debug)

# ---- composite (optionally at upscaled resolution) ----
top_card = Image.open(args.hires).convert("RGB") if args.hires else card
S = top_card.width / W
alpha = Image.fromarray((keep * 255).astype(np.uint8)).resize(top_card.size, Image.BILINEAR)
if S > 1.5:  # pull the upscaled mask edge in ~1 original px to drop old-art fringes
    alpha = alpha.filter(ImageFilter.MinFilter(2 * round(S / 2) + 1))
alpha = alpha.filter(ImageFilter.GaussianBlur(0.6 * S))

x0, y0, x1, y1 = (round(v * S) for v in (AX0, AY0, AX1, AY1))
w, h = x1 - x0, y1 - y0
art = ImageOps.exif_transpose(Image.open(args.photo)).convert("RGB")
aw, ah = art.size
ch = min(ah, aw * h / w) / args.zoom
cw = ch * w / h
left = min(max(args.cx * aw - cw / 2, 0), aw - cw)
top = min(max(args.cy * ah - ch / 2, 0), ah - ch)
patch = art.resize((w, h), Image.LANCZOS, box=(left, top, left + cw, top + ch))

# darken the bottom band behind the credits, like the original art's dark footer
p = np.asarray(patch).astype(float)
yy = (np.arange(h)[:, None, None] + y0) / S
p *= 1 - 0.6 * np.clip((yy - 860) / 50, 0, 1)
patch = Image.fromarray(p.clip(0, 255).astype(np.uint8))

base = top_card.copy()
base.paste(patch, (x0, y0))
out = Image.composite(top_card, base, alpha)
dpi = round(out.width / 2.5)  # standard card is 2.5in wide
out.save(args.out, dpi=(dpi, dpi))
print(f"saved {args.out} {out.size} @ {dpi} dpi (2.5x3.5in)")
