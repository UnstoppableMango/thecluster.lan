{
  pkgs ? import <nixpkgs> { },
  bun2nix,
  version ? "0.1.0",
}:
let
  web = pkgs.stdenv.mkDerivation {
    pname = "thecluster-web";
    inherit version;

    src = pkgs.lib.cleanSource ./src/web;
    nativeBuildInputs = [ bun2nix.hook ];
    bunDeps = bun2nix.fetchBunDeps {
      bunNix = ./src/web/bun.nix;
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
      license = pkgs.lib.licenses.mit;
    };
  };

  api = pkgs.buildGoApplication {
    pname = "thecluster-api";
    inherit version;

    src = pkgs.lib.cleanSource ./src/api;
    modules = ./src/api/gomod2nix.toml;
    subPackages = [ "./cmd/thecluster-api" ];
    ldflags = [
      "-s"
      "-w"
    ];

    meta = {
      description = "THECLUSTER API";
      license = pkgs.lib.licenses.mit;
    };
  };

  app = pkgs.runCommand "thecluster-app" { } ''
    mkdir -p $out/bin $out/wwwroot
    cp -r ${api}/bin/. $out/bin/
    cp -r ${web}/wwwroot/. $out/wwwroot/
  '';

  docker = pkgs.dockerTools.streamLayeredImage {
    name = "thecluster.lan";
    tag = version;
    contents = [
      app
      pkgs.cacert
    ];

    config = {
      Cmd = [ "/bin/thecluster-api" ];
      Env = [
        "PORT=8080"
        "STATIC_DIR=/wwwroot"
      ];
      ExposedPorts = {
        "8080/tcp" = { };
      };
    };
  };
in
{
  inherit
    api
    app
    docker
    web
    ;
}
