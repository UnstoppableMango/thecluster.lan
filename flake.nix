{
  description = "THECLUSTER internal dashboard";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs?ref=nixos-unstable";
    systems.url = "github:nix-systems/default";

    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      inputs.nixpkgs-lib.follows = "nixpkgs";
    };

    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };

    bun2nix = {
      url = "github:nix-community/bun2nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };

    gomod2nix = {
      url = "github:nix-community/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
      inputs.flake-utils.inputs.systems.follows = "systems";
    };
  };

  outputs =
    inputs@{ flake-parts, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = import inputs.systems;

      imports = with inputs; [
        treefmt-nix.flakeModule
      ];

      perSystem =
        { inputs', system, ... }:
        let
          pkgs = import inputs.nixpkgs {
            inherit system;
            overlays = with inputs; [
              bun2nix.overlays.default
              gomod2nix.overlays.default
            ];
          };

          version = "0.1.0";

          api = pkgs.callPackage ./nix/api.nix {
            inherit pkgs version;
          };
          web = pkgs.callPackage ./nix/web.nix {
            inherit pkgs version;
            bun2nix = inputs'.bun2nix.packages.default;
          };
          app = pkgs.callPackage ./nix/app.nix {
            inherit pkgs api web;
          };
          ctr = pkgs.callPackage ./nix/ctr.nix {
            inherit pkgs app;
          };
        in
        {
          _module.args = { inherit pkgs; };

          packages = {
            inherit
              api
              app
              ctr
              web
              ;
            default = app;
          };

          apps.api = {
            type = "app";
            program = "${api}/bin/thecluster-api";
            meta.description = "THECLUSTER API";
          };

          devShells.default = pkgs.mkShell {
            packages = with pkgs; [
              actionlint
              bun
              bun2nix
              docker
              go
              gomod2nix
              gopls
              kubernetes-helm
              nil
              nixfmt
              nodejs
            ];

            BUN = "${pkgs.bun}/bin/bun";
            BUN2NIX = "${pkgs.bun2nix}/bin/bun2nix";
            GO = "${pkgs.go}/bin/go";
            GOMOD2NIX = "${pkgs.gomod2nix}/bin/gomod2nix";
            HELM = "${pkgs.kubernetes-helm}/bin/helm";
            NIXFMT = "${pkgs.nixfmt}/bin/nixfmt";
          };

          treefmt = {
            programs.gofmt.enable = true;
            programs.nixfmt.enable = true;
          };
        };
    };
}
