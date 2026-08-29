test:
  go test ./...

# Render the example wedding cards.
cards file="mpc/examples/wedding.mpc":
  go run ./mpc/cmd/mpc build {{file}}

# One-off: fetch the Card Conjurer checkout mpc renders with.
cards-setup:
  go run ./mpc/cmd/mpc fetch
