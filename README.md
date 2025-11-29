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

How to clean up:

```bash
pnpm lint
pnpm typecheck
```

## CLI helpers

The repository includes a Go-based CLI for operational tasks.

### Cloudflare DNS

- Requires `go` (1.25+) and Cloudflare credentials. You can authenticate with either:
  - an API token: `export CLOUDFLARE_API_TOKEN="<token>"`
  - or an API key and email: `export CLOUDFLARE_API_KEY="<key>"; export CLOUDFLARE_API_EMAIL="<email>"`
- `make export-cloudflare-token TOKEN=<token>` will append the export command to `.env.cloudflare`; source that file to load the token in a shell.

- List the DNS records for a zone:

  ```bash
  pmg cloudflare list example.com
  ```

  The command prints each record's type, name, content, TTL (with `auto` for managed TTLs), and whether it is proxied.
