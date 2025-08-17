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

### Supabase

```sh
supabase link --project-ref <project-id>
# You can get <project-id> from your project's dashboard URL: https://supabase.com/dashboard/project/<project-id>
supabase login
# OR 
export SUPABASE_ACCESS_TOKEN=<token>
supabase db pull
```

```.env
SUPABASE_URL="YOUR_SUPABASE_URL"
SUPABASE_KEY="YOUR_SUPABASE_ANON_KEY"
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
