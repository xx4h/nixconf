{
  description = "nixconf - Repository manager for NixOS multi-repo configuration";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    let
      version = let
        versionFile = builtins.readFile ./VERSION;
        trimmed = builtins.replaceStrings [ "\n" "\r" " " ] [ "" "" "" ] versionFile;
      in
        if trimmed != "" then trimmed
        else if (self ? shortRev) then self.shortRev
        else "dev";

      commit = if (self ? rev) then self.rev else "dirty";

      date = let
        raw = self.lastModifiedDate or "19700101000000";
        year = builtins.substring 0 4 raw;
        month = builtins.substring 4 2 raw;
        day = builtins.substring 6 2 raw;
        hour = builtins.substring 8 2 raw;
        min = builtins.substring 10 2 raw;
        sec = builtins.substring 12 2 raw;
      in "${year}-${month}-${day}T${hour}:${min}:${sec}Z";

      mkNixconf = pkgs: pkgs.buildGoModule {
        pname = "nixconf";
        inherit version;

        src = ./.;

        vendorHash = "sha256-haYm6K44hDagVNx5D0tRfc8uLjTwrbiFhiNqKiBD5Uo=";

        env.CGO_ENABLED = 0;

        nativeBuildInputs = [ pkgs.installShellFiles ];

        ldflags = [
          "-s"
          "-w"
          "-X github.com/xx4h/nixconf/cmd.version=${version}"
          "-X github.com/xx4h/nixconf/cmd.commit=${commit}"
          "-X github.com/xx4h/nixconf/cmd.date=${date}"
        ];

        postInstall = pkgs.lib.optionalString
          (pkgs.stdenv.buildPlatform.canExecute pkgs.stdenv.hostPlatform) ''
          installShellCompletion --cmd nixconf \
            --bash <($out/bin/nixconf completion bash) \
            --zsh  <($out/bin/nixconf completion zsh) \
            --fish <($out/bin/nixconf completion fish)
        '';

        meta = with pkgs.lib; {
          description = "Repository manager for NixOS multi-repo configuration";
          homepage = "https://github.com/xx4h/nixconf";
          license = licenses.asl20;
          maintainers = [ ];
          mainProgram = "nixconf";
        };
      };
    in
    {
      overlays.default = final: prev: {
        nixconf = mkNixconf final;
      };

      homeManagerModules.default = import ./nix/hm-module.nix self;
      homeManagerModules.nixconf = self.homeManagerModules.default;
    }
    //
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages = {
          nixconf = mkNixconf pkgs;
          default = self.packages.${system}.nixconf;
        };

        apps = {
          nixconf = flake-utils.lib.mkApp {
            drv = self.packages.${system}.nixconf;
          };
          default = self.apps.${system}.nixconf;
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            golangci-lint
            goreleaser
            go-task
            gotools
            editorconfig-checker
            prettier
            git
            nix
          ];

          shellHook = ''
            echo "nixconf development shell"
            echo "Go: $(go version)"
          '';
        };
      }
    );
}
