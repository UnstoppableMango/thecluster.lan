{
  pkgs ? import <nixpkgs> { },
  lib,
  version,
}:
pkgs.buildGoApplication {
  pname = "thecluster-api";
  inherit version;

  src = lib.cleanSource ../api;
  modules = ../api/gomod2nix.toml;
  subPackages = [ "cmd/thecluster-api" ];

  ldflags = [
    "-s"
    "-w"
  ];

  meta = {
    description = "THECLUSTER API";
    license = lib.licenses.mit;
  };
}
