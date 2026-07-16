# Hanzo App

**AI web & app builder.** Describe what you want and Hanzo App generates the
HTML, CSS, and JavaScript for it, live in the browser — then iterates with you
through follow-up edits. Publish and share instantly.

<div align="center">
  <a href="https://hanzo.app" target="_blank">Visit Hanzo App →</a>
</div>

## What it is

Hanzo App is the builder surface of the Hanzo AI cloud (sibling to
[Hanzo Chat](https://hanzo.chat)). It runs entirely on
[Hanzo Cloud](https://hanzo.ai) as its backend:

- **Generation** streams from `api.hanzo.ai/v1` — the Hanzo inference gateway.
- **Sign-in** is Hanzo IAM (hanzo.id) via OIDC/PKCE; work is scoped per org.
- **Storefront** (the build → become-a-store flywheel) reads the org's catalog,
  cart, and checkout from Hanzo Commerce.

## Develop

```bash
npm install
npm run dev        # http://localhost:3000
```

Copy `.env.production.example` to `.env` and fill in the Hanzo IAM and Commerce
values for your deployment.

## Tech

Next.js · React 19 · TypeScript · `@hanzo/ui`. Provider-agnostic model picker;
Hanzo Cloud is the default backend.

## License

MIT — see [LICENSE](LICENSE) and [NOTICE](NOTICE).
