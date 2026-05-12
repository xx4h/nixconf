# Home Manager module for nixconf.
#
# Used via the flake's `homeManagerModules.default` (or `.nixconf`) output.
# `self` is captured at flake evaluation so the default package resolves
# from this flake's outputs, regardless of whether the user has applied
# our overlay to their nixpkgs.
self:
{ config, lib, pkgs, ... }:

let
  cfg = config.programs.nixconf;
  defaultPackage = self.packages.${pkgs.stdenv.hostPlatform.system}.nixconf;
in
{
  options.programs.nixconf = {
    enable = lib.mkEnableOption
      "nixconf, a repository manager for multi-repo NixOS configurations";

    package = lib.mkOption {
      type = lib.types.package;
      default = defaultPackage;
      defaultText = lib.literalExpression
        "nixconf.packages.\${pkgs.stdenv.hostPlatform.system}.nixconf";
      description = "The nixconf package to install.";
    };

    enableBashIntegration = lib.mkEnableOption "Bash shell completion" // {
      default = true;
    };

    enableZshIntegration = lib.mkEnableOption "Zsh shell completion" // {
      default = true;
    };

    enableFishIntegration = lib.mkEnableOption "Fish shell completion" // {
      default = true;
    };
  };

  config = lib.mkIf cfg.enable {
    home.packages = [ cfg.package ];

    # The package already drops completion files under
    # $out/share/{bash-completion,zsh/site-functions,fish/vendor_completions.d},
    # which HM-managed shells pick up automatically when their respective
    # `enableCompletion` option is on. The hooks below additionally source
    # completions directly from the binary, matching the idiom used by
    # `programs.atuin` and similar modules — useful when a shell isn't
    # configured to auto-load vendor completions.
    programs.bash.initExtra = lib.mkIf cfg.enableBashIntegration ''
      source <(${lib.getExe cfg.package} completion bash)
    '';

    programs.zsh.initContent = lib.mkIf cfg.enableZshIntegration ''
      source <(${lib.getExe cfg.package} completion zsh)
    '';

    programs.fish.interactiveShellInit = lib.mkIf cfg.enableFishIntegration ''
      ${lib.getExe cfg.package} completion fish | source
    '';
  };
}
