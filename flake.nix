{
  description = "looprun package and development environment";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
      treefmt-nix,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages.default = pkgs.callPackage ./looprun.nix { };

        devShells.default = pkgs.mkShellNoCC {
          packages = with pkgs; [
            go
            gopls
            cobra-cli
            nil
          ];
        };

        formatter = treefmt-nix.lib.mkWrapper pkgs {
          projectRootFile = "flake.nix";
          programs = {
            nixfmt.enable = true;
            gofmt.enable = true;
            prettier = {
              enable = true;
              includes = [ "*.md" ];
            };
          };
          settings.global.excludes = [
            ".envrc"
            "LICENSE"
          ];
        };
      }
    );
}
