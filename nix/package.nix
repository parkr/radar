{
  lib,
  buildGoModule,
  rev ? "dev",
}:

buildGoModule {
  pname = "radar";
  version = rev;

  src = lib.fileset.toSource {
    root = ./..;
    fileset = lib.fileset.difference ./.. (
      lib.fileset.unions [
        ./.
        (./.. + "/.github")
        (./.. + "/flake.nix")
        (lib.fileset.maybeMissing (./.. + "/flake.lock"))
      ]
    );
  };

  # Update with script/update-vendor-hash after changing go.mod / go.sum.
  vendorHash = "sha256-9Es4bVLRFiM99RXHftss/Ua5TGc+/s+A5mPVauTV8P8=";

  # The Docker image is built static; so is this.
  env.CGO_ENABLED = "0";

  subPackages = [
    "cmd/radar"
    "cmd/radar-healthcheck"
    "cmd/radar-poster"
  ];

  # Tests fetch real web pages for their titles, which the sandbox has no network
  # for; the Docker build runs them.
  doCheck = false;

  meta = {
    description = "Keeps track of links you need to get back to, as a daily GitHub issue";
    homepage = "https://github.com/parkr/radar";
    mainProgram = "radar";
  };
}
