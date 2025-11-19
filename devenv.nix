{
  pkgs,
  lib,
  config,
  ...
}:
let
  templ-newer = pkgs.templ.overrideAttrs (oldAttrs: rec {
    version = "0.3.960";
    src = pkgs.fetchFromGitHub {
      owner = "a-h";
      repo = "templ";
      rev = "v${version}";
      hash = "sha256-GCbqaRC9KipGdGfgnGjJu04/rJlg+2lgi2vluP05EV4="; # Run once to get the correct hash, then fill it in
    };
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
