{
  pkgs ? <nixpkgs>,
  lib,
  bun2nix,
  version ? "0.1.0",
}:
bun2nix.mkDerivation {
  pname = "thecluster-web";
  inherit version;

  src = lib.cleanSource ../src/web;

  bunDeps = pkgs.bun2nix.fetchBunDeps {
    bunNix = ../src/web/bun.nix;
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
