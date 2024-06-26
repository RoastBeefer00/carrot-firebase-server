{ pkgs ? import <nixpkgs> {}
, golang ? pkgs.go
}:
pkgs.stdenv.mkDerivation {
    name = "go-shell";

    buildInputs = [ golang
                  ];

    shellHook = ''
        # export GOPATH=`pwd`
        unset GOPATH
        export PATH=$GOPATH/bin:$PATH
    '';
}
