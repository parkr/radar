# NixOS module for radar: `services.radar.enable = true;` runs the server as a
# hardened systemd service. It keeps no local state: the radar is a GitHub
# issue, and the optional Atom feed is rebuilt from it.
#
# Imported as `inputs.radar.nixosModules.default`. It takes the flake itself only
# to label the package with its revision.
self:
{
  config,
  lib,
  pkgs,
  utils,
  ...
}:

let
  cfg = config.services.radar;
  inherit (lib)
    mkOption
    mkEnableOption
    mkIf
    types
    ;

  # An IPv6 literal needs brackets in host:port; leave an already-bracketed one.
  bindHost =
    if lib.hasPrefix "[" cfg.listenAddress || !(lib.hasInfix ":" cfg.listenAddress) then
      cfg.listenAddress
    else
      "[${cfg.listenAddress}]";

  # Binding below 1024 needs CAP_NET_BIND_SERVICE, which the service otherwise
  # drops along with every other capability.
  privilegedPort = cfg.port < 1024;

  noSpaces = "[^[:space:]]+";

  # RuntimeDirectory=radar. The feed config is written here at start because it
  # holds a secret and so can't live in the Nix store.
  feedConfigPath = "/run/radar/feed.json";

  feedBase = pkgs.writeText "radar-feed.json" (
    builtins.toJSON {
      Title = cfg.feed.title;
      URL = cfg.feed.url;
      Description = cfg.feed.description;
      AuthorName = cfg.feed.authorName;
    }
  );

  # Merges the API key (a systemd credential) into the feed config.
  writeFeedConfig = pkgs.writeShellScript "radar-feed-config" ''
    ${lib.getExe pkgs.jq} --rawfile key "$CREDENTIALS_DIRECTORY/feed-api-key" \
      '. + {APIKey: ($key | rtrimstr("\n"))}' ${feedBase} > "$RUNTIME_DIRECTORY/feed.json"
  '';
in
{
  options.services.radar = {
    enable = mkEnableOption "radar, which turns emailed links into a daily GitHub issue";

    package = mkOption {
      type = types.package;
      default = pkgs.callPackage ./package.nix {
        rev = self.shortRev or self.dirtyShortRev or "dev";
      };
      defaultText = lib.literalExpression "(built from this flake's source with your nixpkgs)";
      description = ''
        The radar package to run. By default it is built from this flake's source
        with the *consumer's* `pkgs`, so it shares your nixpkgs (and its Go and
        libc) rather than pulling in a second copy.
      '';
    };

    repo = mkOption {
      type = types.strMatching "[^/[:space:]]+/[^/[:space:]]+";
      example = "owner/name";
      description = ''
        The GitHub repository whose issues are the radar (`RADAR_REPO`), as
        `owner/name`. radar splits this on `/` without checking, so anything
        else would crash it; the type rejects it up front.
      '';
    };

    allowedSenders = mkOption {
      type = types.listOf (types.strMatching "[^,[:space:]]+");
      example = [ "me@example.com" ];
      description = ''
        Addresses whose mail is accepted (`RADAR_ALLOWED_SENDERS`); all others
        are rejected. Required: an empty list would reject everything. Entries
        can't contain commas or whitespace.
      '';
    };

    mention = mkOption {
      type = types.str;
      default = "";
      example = "@parkr";
      description = ''
        User or team mentioned in each day's issue (`RADAR_MENTION`). radar logs
        a warning when it is empty.
      '';
    };

    mailgunDomain = mkOption {
      type = types.nonEmptyStr;
      example = "example.com";
      description = "The Mailgun domain inbound mail arrives on (`MG_DOMAIN`).";
    };

    environmentFile = mkOption {
      type = types.path;
      example = "/run/secrets/radar.env";
      description = ''
        A systemd `EnvironmentFile` with the secrets, kept out of the Nix store:

        - `GITHUB_ACCESS_TOKEN`: a token that can open issues on `repo`
        - `MG_API_KEY` and `MG_PUBLIC_API_KEY`: the Mailgun keys
        - `MG_FROM_EMAIL` (optional): the address replies come from
      '';
    };

    hour = mkOption {
      type = types.strMatching "([01][0-9]|2[0-3])";
      default = "03";
      example = "08";
      description = ''
        Hour of the day (`-hour`, two digits, 00 to 23), in `timeZone`, at which
        the day's issue is generated.

        radar checks this on a timer that fires every hour from when the
        process *starts* (not on the hour), and generates at the first check
        that falls in this hour. After a restart, the next check is an hour
        later, so a restart after this hour has begun skips that day's issue.
        Send `SIGUSR2` to generate one immediately regardless of the hour.
      '';
    };

    timeZone = mkOption {
      type = types.nonEmptyStr;
      default = "UTC";
      example = "America/Los_Angeles";
      description = ''
        Zone `hour` is read in. radar compares `hour` to the process's local
        hour, so this is set as the service's `TZ`, with the zone database
        supplied so it doesn't depend on the host having one. The Docker image
        runs in UTC; so does this by default.
      '';
    };

    listenAddress = mkOption {
      type = types.strMatching noSpaces;
      default = "127.0.0.1";
      example = "[::1]";
      description = ''
        Address to listen on (`-http`), loopback by default for a reverse proxy.
        A bare IPv6 address is bracketed for you.
      '';
    };

    port = mkOption {
      type = types.ints.between 1 65535;
      default = 8291;
      description = ''
        Port to listen on (`-http`). A port below 1024 gets
        `CAP_NET_BIND_SERVICE` automatically.
      '';
    };

    openFirewall = mkOption {
      type = types.bool;
      default = false;
      description = ''
        Open `port` in the firewall. Leave off behind a reverse proxy on the
        default loopback address; turn on only if you set a non-loopback
        `listenAddress` and want radar reachable directly.
      '';
    };

    feed = mkOption {
      default = null;
      description = ''
        Serve the issues as an Atom feed at `/feed.atom` (`-feedConfig`). The
        feed requires its API key, so it is off unless this is set.
      '';
      type = types.nullOr (
        types.submodule {
          options = {
            title = mkOption {
              type = types.str;
              example = "owner/name";
            };
            url = mkOption {
              type = types.str;
              example = "https://radar.example.com/feed.atom";
              description = "The feed's own public URL.";
            };
            description = mkOption {
              type = types.str;
              default = "";
            };
            authorName = mkOption {
              type = types.str;
              default = "";
            };
            apiKeyFile = mkOption {
              type = types.path;
              example = "/run/secrets/radar-feed-key";
              description = ''
                File holding the feed's API key and nothing else (a trailing
                newline is dropped). Passed to the service as a systemd
                credential and merged into the feed config at start.
              '';
            };
          };
        }
      );
    };
  };

  config = mkIf cfg.enable {
    assertions = [
      {
        assertion = cfg.allowedSenders != [ ];
        message = "services.radar.allowedSenders must list at least one address: radar rejects mail from anyone it isn't told about.";
      }
    ];

    networking.firewall.allowedTCPPorts = mkIf cfg.openFirewall [ cfg.port ];

    systemd.services.radar = {
      description = "radar, links to remember as a daily GitHub issue";
      after = [ "network-online.target" ];
      wants = [ "network-online.target" ];
      wantedBy = [ "multi-user.target" ];

      environment = {
        RADAR_REPO = cfg.repo;
        RADAR_ALLOWED_SENDERS = lib.concatStringsSep "," cfg.allowedSenders;
        RADAR_MENTION = cfg.mention;
        MG_DOMAIN = cfg.mailgunDomain;
        TZ = cfg.timeZone;
        ZONEINFO = "${pkgs.tzdata}/share/zoneinfo";
      };

      serviceConfig = {
        # systemd expands `%` specifiers and `$VAR` even inside quotes, so the
        # arguments need its escaping, not the shell's.
        ExecStart = utils.escapeSystemdExecArgs (
          [
            (lib.getExe cfg.package)
            "-http=${bindHost}:${toString cfg.port}"
            "-hour=${cfg.hour}"
          ]
          ++ lib.optional (cfg.feed != null) "-feedConfig=${feedConfigPath}"
        );
        ExecStartPre = lib.optional (cfg.feed != null) writeFeedConfig;
        LoadCredential = lib.optional (cfg.feed != null) "feed-api-key:${cfg.feed.apiKeyFile}";
        EnvironmentFile = cfg.environmentFile;
        RuntimeDirectory = "radar";
        RuntimeDirectoryMode = "0700";
        DynamicUser = true;
        Restart = "on-failure";
        RestartSec = 5;

        # Parses mail from the internet, so it gets nothing it doesn't need.
        NoNewPrivileges = true;
        CapabilityBoundingSet = if privilegedPort then [ "CAP_NET_BIND_SERVICE" ] else "";
        ProtectSystem = "strict";
        ProtectHome = true;
        PrivateTmp = true;
        PrivateDevices = true;
        ProtectKernelTunables = true;
        ProtectKernelModules = true;
        ProtectControlGroups = true;
        RestrictAddressFamilies = [
          "AF_INET"
          "AF_INET6"
          "AF_UNIX"
        ];
        RestrictNamespaces = true;
        LockPersonality = true;
      }
      // lib.optionalAttrs privilegedPort { AmbientCapabilities = [ "CAP_NET_BIND_SERVICE" ]; };
    };
  };
}
