{
  description = "radar: keeps track of links you need to get back to, as a daily GitHub issue";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";

  outputs =
    { self, nixpkgs }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
      forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f nixpkgs.legacyPackages.${system});
    in
    {
      packages = forAllSystems (pkgs: rec {
        radar = pkgs.callPackage ./nix/package.nix {
          rev = self.shortRev or self.dirtyShortRev or "dev";
        };
        default = radar;
      });

      # `import` this in your own config for `pkgs.radar`:
      #   nixpkgs.overlays = [ inputs.radar.overlays.default ];
      overlays.default = final: _prev: {
        radar = final.callPackage ./nix/package.nix {
          rev = self.shortRev or self.dirtyShortRev or "dev";
        };
      };

      apps = forAllSystems (pkgs: {
        default = {
          type = "app";
          program = "${self.packages.${pkgs.system}.radar}/bin/radar";
        };
      });

      devShells = forAllSystems (pkgs: {
        default = pkgs.mkShell {
          packages = [
            pkgs.go
            pkgs.gopls
          ];
        };
      });

      formatter = forAllSystems (pkgs: pkgs.nixfmt-rfc-style);
    };
}
