# carrot-firebase-server

## Local development

Local dev runs against an in-memory Firestore emulator, never the real database.

```sh
devenv up
```

This starts the emulator, loads sample recipes from `cmd/seed`, and runs the app with `air`. Open <http://localhost:8090> (live reload) or <http://localhost:8080>. You're signed in automatically as a local user, without Google.

Stopping `devenv up` discards all data, and the next run starts again from the sample recipes.

- **Admin pages:** the local user is `dev@example.com`, which isn't an admin. To test admin pages, set `env.DEV_LOGIN_EMAIL` to an address in `services.Admins` in `devenv.local.nix`.
- **Recipe file upload:** this calls the Anthropic API, so it fails with the placeholder key. Set `env.ANTHROPIC_API_KEY` in `devenv.local.nix` to use it.
- **Safety check:** with `APP_ENV=development`, the app won't open Firestore unless `FIRESTORE_EMULATOR_HOST` is set and `FIRESTORE_PROJECT_ID` starts with `demo-`. Running `air` by itself needs the emulator already up from `devenv up`.
