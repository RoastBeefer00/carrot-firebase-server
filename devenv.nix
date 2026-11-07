{
  pkgs,
  lib,
  config,
  ...
}:
let
  # Runs the templ version pinned in go.mod, so generated code always matches the
  # runtime it compiles against. Upgrade it with `go get github.com/a-h/templ@latest`.
  templ = pkgs.writeShellScriptBin "templ" ''exec go tool templ "$@"'';

  firestorePort = "8200";
  firestoreProject = "demo-carrot";
  firestoreJar = "${pkgs.google-cloud-sdk.components.cloud-firestore-emulator}/google-cloud-sdk/platform/cloud-firestore-emulator/cloud-firestore-emulator.jar";
in
{
  languages.go.enable = true;
  packages = with pkgs; [
    air
    go
    go-tools
    tailwindcss_4
    templ
  ];

  # These take priority over .env, since godotenv doesn't override variables that
  # are already set. The credentials are placeholders: login is local, and the
  # encryption key only protects the local session cookie.
  env = {
    APP_ENV = "development";
    FIRESTORE_EMULATOR_HOST = "127.0.0.1:${firestorePort}";
    FIRESTORE_PROJECT_ID = firestoreProject;
    INSECURE_COOKIES = "true";
    GOOGLE_CLIENT_ID = "local-dev-unused";
    GOOGLE_CLIENT_SECRET = "local-dev-unused";
    REDIRECT_URL = "http://localhost:8090/oauth2/callback";
    ENCRYPTION_KEY_BASE64 = "bG9jYWwtZGV2LW9ubHktbm90LWEtcmVhbC1zZWNyZXQ=";
    ANTHROPIC_API_KEY = "local-dev-unused";
  };

  # The emulator holds everything in memory, so stopping it discards all data.
  processes.firestore.exec = ''
    exec ${pkgs.jdk21_headless}/bin/java -Duser.language=en -cp ${firestoreJar} \
      com.google.cloud.datastore.emulator.firestore.CloudFirestore \
      --host 127.0.0.1 --port ${firestorePort} --project_id ${firestoreProject}
  '';

  processes.app.exec = ''
    until (exec 3<>/dev/tcp/127.0.0.1/${firestorePort}) 2>/dev/null; do sleep 0.5; done
    go run ./cmd/seed || exit 1
    exec air
  '';
}
