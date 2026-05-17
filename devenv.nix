{
  pkgs,
  lib,
  config,
  ...
}:
let
  templ-newer = pkgs.templ.overrideAttrs (oldAttrs: rec {
    version = "0.3.1020";
    src = pkgs.fetchFromGitHub {
      owner = "a-h";
      repo = "templ";
      rev = "v${version}";
      hash = "sha256-wv7qKZfnavz8lxfaOaIJJySNsXsjke1ADJuv2kgQOHE=";
    };
    vendorHash = "sha256-i4uDGZb3VZUvIyO2Tt53VR1Do/3OYtj6JccGoFnnlbs=";
  });
in
{
  languages.go.enable = true;
  packages = with pkgs; [
    air
    go
    go-tools
    tailwindcss_4
    templ-newer
  ];
}
