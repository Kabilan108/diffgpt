{
  description = "AI-powered git diff analyzer using OpenAI GPT";
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };
  outputs = { self, nixpkgs }: let
    system = "x86_64-linux";
    pkgs = import nixpkgs { inherit system; };
  in {
    packages.${system}.default = pkgs.stdenv.mkDerivation rec {
      pname = "diffgpt";
      version = "0.4.1";
      src = pkgs.fetchurl {
        url = "https://github.com/Kabilan108/diffgpt/releases/download/v${version}/diffgpt-linux-amd64.tar.gz";
        sha256 = "sha256-7yc3gombv2HkQ9FUeHVQpcgZVrzCcxjGzRclYI2r5Xo=";
      };
      installPhase = ''
        mkdir -p $out/bin
        cp bin/diffgpt $out/bin/
        chmod +x $out/bin/diffgpt
      '';
    };
    devShells.${system}.default = pkgs.mkShell {
      buildInputs = with pkgs; [
        go
        gopls
        nodejs_20
      ];
      shellHook = ''
        export NPM_CONFIG_PREFIX="$HOME/.npm-global"
        export PATH="$HOME/.npm-global/bin:$PATH"
        if [ ! -f "$HOME/.npm-global/bin/claude" ]; then
          npm install -g @anthropic-ai/claude-code
        fi
      '';
    };
  };
}
