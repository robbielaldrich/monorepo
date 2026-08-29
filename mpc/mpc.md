# Notes

MPC = MakePlayingCards.

Original working notes, now handled by the tool — see README.md.

- Use cardconjurer.
  → `mpc build` drives it headlessly; `mpc serve` opens it for hand-tweaking.
- Change scale to make full art.
  → `@frame fullart` / `@frame borderless`, or `@art-scale` / `@art-fit`.
- Edit the resulting png, since mpc will scale down so the corners are inside.
  → `out/mpc/` already has 0.12in bleed on every edge and square corners, so
    MPC has nothing left to scale. See "Why the bleed step matters" in README.md.
