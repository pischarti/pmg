# PostMil Generation

## Dev Environment with Nix

- Install [nix](https://zero-to-nix.com/concepts/nix-installer/)
- Add 'experimental-features = nix-command flakes' to ~/.config/nix/nix.conf to enable this feature
- Build dev shell `nix develop`
- Add `nix` language server `nix profile install nixpkgs#nixd`
- To make it look nice again, use 'zsh'

## Setup

Make sure to install the dependencies:

```bash
pnpm install
```

## Development Server

Start the development server on `http://localhost:3000`:

```bash
pnpm dev
```

## Production

Build the application for production:

```bash
pnpm build
```

Locally preview production build:

```bash
pnpm preview
```

How to stash changes:

```bash
git stash --include-untracked
```

How to lint:

```bash
pnpm lint
pnpm lint --fix
```
