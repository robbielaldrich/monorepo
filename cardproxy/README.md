# cardproxy

Custom-art proxies of trading cards (printed with a different card back). Take a
card scan, swap its art for a personal photo, output a print-ready PNG.

## Setup (per machine)

```sh
./setup.sh      # makes .venv (pillow/numpy/scipy) and downloads Real-ESRGAN into tools/
```

macOS only as written. Real-ESRGAN is the ncnn-vulkan universal binary (arm64 + x86_64),
from the official GitHub release v0.2.5.0. On Linux/Windows, swap the zip URL in `setup.sh`.

## Workflow

1. **Get the card scan** at the highest resolution you can find.
   - Magic: Scryfall "PNG" download (745×1040, transparent rounded corners). That's the max.
   - Pokémon: ~734×1024 is about the best available online.
2. **AI-upscale the card 4×**: `./upscale.sh card.png card_4x.png`
3. **Composite** at the upscaled resolution (masks/coords are measured on the original,
   then scaled up):
   - Rectangular art window (old-frame Magic):
     `.venv/bin/python framed.py card.png photo.jpg out.png --hires card_4x.png [--top 0.125] [--window x0,y0,x1,y1]`
   - Full-art Pokémon TAG TEAM:
     `.venv/bin/python tagteam.py card.jpg photo.jpg out.png --hires card_4x.png [--cx .52 --cy .5 --zoom 1]`
4. **Look at the output** (full, plus zoomed crops of the text and bars) before printing.

Outputs have their DPI set to `width / 2.5in`, so the printer's quality check reads a true value.

## Things learned

- **Printer "Low resolution ~96 DPI" warning.** The printer's quality check read the
  DPI label stored in the file (the source JPG said 96), not the real pixel density.
  717 px across a 2.5" card is really ~287 DPI. Still, scans are soft at print size, so:
- **Upscaling:** Real-ESRGAN `realesrgan-x4plus-anime` gives clearly crisper card
  text and outlines than `realesrgan-x4plus` or Lanczos (compared side by side on the
  Tag Team rules text). Don't upscale the photo; the scripts paste the full-res photo
  directly at the 4× size.
- **realesrgan-ncnn-vulkan segfaults (exit 139) if it can't find `models/`.** It looks
  relative to the working directory, so always pass `-m <dir>/models` (upscale.sh does).
- **The upscaler softens transparent corners.** framed.py reuses the original PNG's
  alpha channel, scaled up.
- **Magic 8th/Mirrodin frame (745×1040 Scryfall PNG):** the art window inside its 1px
  black border is x 60–682, y 120–576 (623×457, aspect ≈1.363). Found by scanning
  pixel brightness for the black border line.
- **Pokémon full-art (TAG TEAM, 717×1000):** the art is behind everything, so the approach is
  "photo under the art window + a keep-mask of the card's own parts on top". The mask is built from:
  - Frame outside the art window (x 28–687, y 28–977).
  - Frame yellow picked out by color: R≈G, low B (e.g. 250,249,37). The sunset
    orange in the art has G≪R (e.g. 246,213,105), so they separate cleanly. Limit it to the
    banner/rule-box areas anyway.
  - Hand-measured shapes and circles for the BASIC pill, TAG TEAM banner, GX attack bar,
    weakness bar, rule box and energy icons. Measured by drawing a 10px grid over zoomed crops.
  - GX logo: shape ∩ (cyan | white | dark) colors, so the old sunset behind it is dropped.
  - Weakness bar: keep only low-saturation (silver) pixels, so the diagonal gaps between
    segments show the photo.
  - **Text with white outlines:** find connected dark (or blue, for GX attack text) groups of
    pixels, keep only those whose 2px surrounding ring is mostly white. That's what separates
    letters from dark art. Then expand by 3px to include the outline.
  - At 4×: bilinear-upscale the mask, shrink it by ~1 original px (MinFilter) to drop
    old-art fringes, then blur it slightly.
  - Darken the photo toward the bottom (y > 860) so the white-outlined credit text
    stays readable, like the original's dark footer.
- Always check zoomed crops. Leftovers show up as thin slivers of the old art's colors
  (lavender/orange) at shape edges; fix by nudging the shape or adding a color test.

## Known gaps / ideas

- The "©" in the Pokémon copyright line is lost by the text mask.
- The name and attack text are unchanged (still Latias&Latios etc.). Next step could be to
  re-render the name/attacks with a matching font and a white outline.
- Coordinates are specific to each frame layout. A new layout needs new measurements
  (grid-overlay crops + pixel sampling along rows and columns work well).
- Unknown whether the printer wants bleed (~1/8" extra per side). If so, add a margin.

## Past runs (inputs live on the Desktop, not in git)

| card | inputs | output |
|---|---|---|
| Mesmeric Orb (MTG) | `~/Desktop/mesmericorb/{mesmericorb.png,newart.jpeg}` | `mesmericorb_custom_hires.png` 2980×4160 |
| Latias & Latios TAG TEAM | `~/Desktop/tagteam/{183899_in_1000x1000.jpg,IMG_0636.jpeg}` | `tagteam_custom_hires.png` 2868×4000 |
