{
  pkgs ? <nixpkgs>,
  lib,
  stdenv,
  bun2nix,
  version ? "0.1.0",
}:
bun2nix.mkDerivation {
  pname = "thecluster-web";
  inherit version;

  src = lib.cleanSource ../web;

  # nixpkgs' x86_64 bun needs AVX2 and dies with "Illegal instruction" on
  # older CPUs (zeus, a Sandy Bridge Xeon running Hercules agents). Only
  # builders that advertise x86-64-v3 take this build; see
  # UnstoppableMango/nixos#416.
  requiredSystemFeatures = lib.optionals stdenv.hostPlatform.isx86_64 [ "gccarch-x86-64-v3" ];

  bunDeps = pkgs.bun2nix.fetchBunDeps {
    bunNix = ../web/bun.nix;
  };

  buildPhase = ''
    runHook preBuild
    bun run build
    runHook postBuild
  '';

  installPhase = ''
    runHook preInstall
    mkdir -p $out/wwwroot
    cp -r dist/. $out/wwwroot/
    runHook postInstall
  '';

  meta = {
    description = "THECLUSTER web application";
    license = lib.licenses.mit;
  };
}
