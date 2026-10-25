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

      # `services.radar.enable = true;` -- see nix/module.nix for the options.
      #   imports = [ inputs.radar.nixosModules.default ];
      nixosModules.default = import ./nix/module.nix self;

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

      # Tests the NixOS module without booting a VM: evaluates it in minimal
      # systems and asserts on the result (including that bad input is rejected),
      # then runs the real binary with the exact arguments and environment the
      # module generates. Linux only.
      checks = nixpkgs.lib.genAttrs [ "x86_64-linux" "aarch64-linux" ] (
        system:
        let
          inherit (nixpkgs) lib;
          pkgs = nixpkgs.legacyPackages.${system};

          # Just enough for a system to evaluate: no bootloader, a stub root.
          fixture = {
            boot.loader.grub.enable = false;
            fileSystems."/" = {
              device = "none";
              fsType = "tmpfs";
            };
            system.stateVersion = "26.05";
          };

          feed = {
            title = "owner/name";
            url = "https://radar.example.com/feed.atom";
            description = "Links to remember.";
            authorName = "Various";
            apiKeyFile = "/run/secrets/radar-feed-key";
          };

          # A system with services.radar enabled; `settings` overrides the
          # defaults below, `extra` is any other configuration.
          evalWith =
            settings: extra:
            lib.nixosSystem {
              inherit system;
              modules = [
                self.nixosModules.default
                fixture
                extra
                {
                  services.radar = {
                    enable = true;
                    repo = "owner/name";
                    allowedSenders = [ "me@example.com" ];
                    mailgunDomain = "example.com";
                    environmentFile = "/run/secrets/radar.env";
                  }
                  // settings;
                }
              ];
            };
          service = settings: (evalWith settings { }).config.systemd.services.radar;
          exec = settings: (service settings).serviceConfig.ExecStart;
          failedAssertions =
            settings: extra:
            map (a: a.message) (lib.filter (a: !a.assertion) (evalWith settings extra).config.assertions);
          # Option types reject bad values with an evaluation error.
          rejected =
            settings:
            let
              s = service settings;
            in
            !(builtins.tryEval (builtins.deepSeq [ s.serviceConfig s.environment ] true)).success;
          has = lib.hasInfix;
          check = lib.assertMsg;

          tests = [
            # Defaults.
            (check (has ''"-http=127.0.0.1:8291"'' (exec { })) "listens on 127.0.0.1:8291 by default")
            (check (has ''"-hour=03"'' (exec { })) "hour defaults to 03, the binary's own default")
            (check (!(has "-feedConfig" (exec { }))) "no feed config unless a feed is configured")
            (check ((service { }).serviceConfig.ExecStartPre == [ ]) "no feed step without a feed")
            (check ((service { }).serviceConfig.DynamicUser) "runs as a dynamic user")
            (check (
              (service { }).serviceConfig.EnvironmentFile == "/run/secrets/radar.env"
            ) "secrets come from the environment file")
            (check (lib.hasPrefix "radar-" (evalWith { } { }).config.services.radar.package.name) "package is built")

            # Environment.
            (check ((service { }).environment.RADAR_REPO == "owner/name") "repo reaches RADAR_REPO")
            (check ((service { }).environment.MG_DOMAIN == "example.com") "domain reaches MG_DOMAIN")
            (check (
              (service {
                allowedSenders = [
                  "a@example.com"
                  "b@example.com"
                ];
              }).environment.RADAR_ALLOWED_SENDERS == "a@example.com,b@example.com"
            ) "senders join with commas")
            (check ((service { mention = "@me"; }).environment.RADAR_MENTION == "@me") "mention reaches RADAR_MENTION")

            # Schedule: the hour is read in the configured zone.
            (check ((service { }).environment.TZ == "UTC") "the zone defaults to UTC, like the Docker image")
            (check (has "tzdata" (service { }).environment.ZONEINFO) "the zone database is supplied, not assumed")
            (check (
              let
                s = service {
                  hour = "01";
                  timeZone = "America/Los_Angeles";
                };
              in
              s.environment.TZ == "America/Los_Angeles" && has ''"-hour=01"'' s.serviceConfig.ExecStart
            ) "hour and timeZone pass through")
            (check (rejected { hour = "24"; }) "hour 24 is rejected")
            (check (rejected { hour = "8"; }) "a one-digit hour is rejected")
            (check (rejected { hour = "ab"; }) "a non-numeric hour is rejected")
            (check (rejected { timeZone = ""; }) "an empty zone is rejected")

            # repo is split on "/" with no check in radar, so the type is the check.
            (check (rejected { repo = "name"; }) "a repo without an owner is rejected")
            (check (rejected { repo = "a/b/c"; }) "a repo with extra segments is rejected")
            (check (rejected { repo = "a b/c"; }) "a repo with whitespace is rejected")
            (check (rejected { repo = "/c"; }) "a repo with an empty owner is rejected")

            # Senders and domain.
            (check (
              failedAssertions { allowedSenders = [ ]; } { } != [ ]
            ) "an empty sender list is rejected by an assertion")
            (check (rejected { allowedSenders = [ "" ]; }) "an empty sender is rejected")
            (check (rejected { allowedSenders = [ "a@x.com,b@x.com" ]; }) "a sender with a comma is rejected")
            (check (rejected { mailgunDomain = ""; }) "an empty Mailgun domain is rejected")

            # The feed.
            (check (
              let
                s = service { inherit feed; };
              in
              has ''"-feedConfig=/run/radar/feed.json"'' s.serviceConfig.ExecStart
              && builtins.length s.serviceConfig.ExecStartPre == 1
              && s.serviceConfig.LoadCredential == [ "feed-api-key:/run/secrets/radar-feed-key" ]
              && s.serviceConfig.RuntimeDirectory == "radar"
            ) "a feed adds its config path, key credential and merge step")
            (check (
              !(has "radar-feed-key" (builtins.toJSON (service { inherit feed; }).environment))
            ) "the feed key's path is not in the environment")

            # Ports.
            (check (rejected { port = 0; }) "port 0 is rejected")
            (check (rejected { port = 70000; }) "port 70000 is rejected")
            (check (rejected { listenAddress = "a b"; }) "a listenAddress with whitespace is rejected")
            (check (has ''"-http=[::1]:8291"'' (exec { listenAddress = "::1"; })) "a bare IPv6 address is bracketed")
            (check (has ''"-http=[::1]:8291"'' (exec { listenAddress = "[::1]"; })) "an already-bracketed one is left alone")
            (check (has "%%eth0" (exec { listenAddress = "fe80::1%eth0"; })) "a zone id's % is escaped for systemd")
            (check (!((service { }).serviceConfig ? AmbientCapabilities)) "no ambient capabilities by default")
            (check ((service { }).serviceConfig.CapabilityBoundingSet == "") "all capabilities dropped by default")
            (check (
              (service { port = 80; }).serviceConfig.AmbientCapabilities == [ "CAP_NET_BIND_SERVICE" ]
              && (service { port = 80; }).serviceConfig.CapabilityBoundingSet == [ "CAP_NET_BIND_SERVICE" ]
            ) "a port below 1024 gets CAP_NET_BIND_SERVICE")
            (check (
              (evalWith { openFirewall = true; } { }).config.networking.firewall.allowedTCPPorts == [ 8291 ]
              && (evalWith { } { }).config.networking.firewall.allowedTCPPorts == [ ]
            ) "openFirewall opens the port, and only when asked")
          ];

          # The real binary, started with the arguments and environment the
          # module generates for these settings, in the build sandbox (which has
          # loopback and no zone database, so UTC). Proves the flags parse, the
          # hour reaches the schedule, /health answers and the feed is gated.
          runtimeSettings = {
            port = 18291;
            hour = "05";
            inherit feed;
          };
          runtimeService = service runtimeSettings;
          # ExecStart is `"binary" "arg" ...`; none of these args need unescaping.
          unquote = a: lib.removeSuffix "\"" (lib.removePrefix "\"" a);
          runtimeArgs = map (a: lib.replaceStrings [ "/run/radar/feed.json" ] [ "feed.json" ] (unquote a)) (
            lib.tail (lib.splitString " " (exec runtimeSettings))
          );
          runtimeBinary = lib.getExe (evalWith runtimeSettings { }).config.services.radar.package;
          runtimeEnv = lib.concatStringsSep " " (
            lib.mapAttrsToList (n: v: "${n}=${lib.escapeShellArg v}") (
              builtins.removeAttrs runtimeService.environment [ "ZONEINFO" ]
            )
          );
        in
        {
          nixos-module =
            assert builtins.all (t: t) tests;
            pkgs.runCommand "radar-nixos-module-tests" { } "touch $out";

          nixos-module-runtime =
            pkgs.runCommand "radar-nixos-module-runtime"
              {
                nativeBuildInputs = [ pkgs.curl ];
              }
              ''
                cd "$TMPDIR"
                echo '{"Title":"t","URL":"u","Description":"d","AuthorName":"a","APIKey":"secret"}' > feed.json
                env ${runtimeEnv} GITHUB_ACCESS_TOKEN=x MG_API_KEY=x MG_PUBLIC_API_KEY=x \
                  ${runtimeBinary} ${lib.escapeShellArgs runtimeArgs} > radar.log 2>&1 &
                pid=$!
                trap 'kill $pid; cat radar.log' EXIT
                url=http://127.0.0.1:${toString runtimeSettings.port}
                for _ in $(seq 1 50); do curl -sf $url/health >/dev/null && break; sleep 0.2; done
                curl -sf $url/health | grep -q '"Ok":true'
                grep -q "Will generate radar at 05:00 every day" radar.log
                code=$(curl -s -o /dev/null -w '%{http_code}' $url/feed.atom)
                test "$code" = 401
                touch $out
              '';
        }
      );

      formatter = forAllSystems (pkgs: pkgs.nixfmt-rfc-style);
    };
}
