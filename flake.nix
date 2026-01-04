{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs?ref=nixos-unstable";
    systems.url = "github:nix-systems/default";

    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      inputs.nixpkgs-lib.follows = "nixpkgs";
    };

    disko = {
      url = "github:nix-community/disko";
      inputs.nixpkgs.follows = "nixpkgs";
    };

    clan-core = {
      url = "https://git.clan.lol/clan/clan-core/archive/main.tar.gz";

      inputs = {
        nixpkgs.follows = "nixpkgs";
        systems.follows = "systems";
        flake-parts.follows = "flake-parts";
        disko.follows = "disko";
        treefmt-nix.follows = "treefmt-nix";
      };
    };

    dotfiles = {
      url = "github:UnstoppableMango/dotfiles";

      inputs = {
        nixpkgs.follows = "nixpkgs";
        systems.follows = "systems";
        flake-parts.follows = "flake-parts";
        treefmt-nix.follows = "treefmt-nix";
      };
    };

    nixos = {
      url = "github:UnstoppableMango/nixos";

      inputs = {
        nixpkgs.follows = "nixpkgs";
        flake-parts.follows = "flake-parts";
        disko.follows = "disko";
        dotfiles.follows = "dotfiles";
        treefmt-nix.follows = "treefmt-nix";
      };
    };

    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    inputs@{
      flake-parts,
      ...
    }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = import inputs.systems;

      imports = with inputs; [
        clan-core.flakeModules.default
        treefmt-nix.flakeModule
      ];

      clan = {
        imports = [ ./clan.nix ];

        machines = {
          agreus = {
            imports = with inputs; [
              nixos.nixosModules.agreus
            ];

            nixpkgs.hostPlatform = "x86_64-linux";

            # Enable remote Clan commands over SSH
            clan.core.networking.targetHost = "root@jon";
          };

          hades = {
            imports = with inputs; [
              nixos.nixosModules.hades
            ];

            nixpkgs.hostPlatform = "aarch64-linux";
          };
        };
      };

      perSystem =
        { inputs', pkgs, ... }:
        {
          devShells.default = pkgs.mkShell {
            packages = with pkgs; [
              inputs'.clan-core.packages.clan-cli
              nil
              nixfmt
            ];
          };

          treefmt = {
            programs.nixfmt.enable = true;
          };
        };
    };
}
