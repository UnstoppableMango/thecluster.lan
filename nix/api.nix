{
  pkgs ? import <nixpkgs> { },
  lib,
}:
pkgs.buildGoApplication {
  pname = "thecluster-api";
  inherit version;

  src = lib.cleanSource ./src/api;
  modules = ./src/api/gomod2nix.toml;
  subPackages = [ "./cmd/thecluster-api" ];

  ldflags = [
    "-s"
    "-w"
  ];

  meta = {
    description = "THECLUSTER API";
    license = lib.licenses.mit;
  };
}
