# Attribution

## Card Conjurer

All card rendering in this tool is done by **Card Conjurer**, which was created
by **Kyle Burton**. It is the work of years, and it is the only reason this
directory is a few hundred lines instead of a few thousand.

- Upstream repository: <https://github.com/Investigamer/cardconjurer>
- Support the original creator: <https://www.paypal.me/kyleburtondonate>

Card Conjurer's own site was taken offline in November 2022 after Wizards of the
Coast served its creator with a cease and desist. The repository above exists to
keep the application usable locally and to preserve its templates.

### This tool does not redistribute Card Conjurer

`mpc fetch` clones the upstream repository into your own cache directory
(`~/.cache/mpc/cardconjurer` by default). No frames, masks, fonts, mana symbols
or set symbols are vendored into this monorepo. `mpc build` starts a loopback
HTTP server over that checkout purely so a headless browser can load the app;
nothing is served off this machine.

### Frame artwork

Card Conjurer's legal page credits card frame images, or elements used within
them, to: **Chilli_Axe, Kentu, thevodkaboy, Timmy XD69, Sheepwave, FeuerAmeise,
Smyris, Checkbox and TheGnomeRobotic**.

### Licensing

The upstream repository ships no open-source licence. Card Conjurer's own terms
state that all other content is *Copyright © 2020 Card Conjurer*, and that the
service is **intended for personal use only, not commercial**. Treat anything
this tool produces the same way.

## Wizards of the Coast

Card Conjurer is neither affiliated with, sponsored by, nor endorsed by Wizards
of the Coast. The fonts, mana symbols, card frames and other related images it
uses are trademarks and copyrights of Wizards of the Coast, LLC, a subsidiary of
Hasbro, Inc. Nor is this tool affiliated with Legend Story Studios or Scryfall.

## Your artwork

Card Conjurer's terms put the responsibility for artwork on the person making
the card:

> All user-uploaded material is property of the original artist, and it is the
> user's responsibility to ensure that these materials are properly credited.

That is why `@artist` exists and why `mpc build` warns when a card has art but
no artist credit. Fill it in.

## MakePlayingCards

Not affiliated with or endorsed by MakePlayingCards.com. The bleed dimensions in
`internal/print` are taken from their published poker-card template.
