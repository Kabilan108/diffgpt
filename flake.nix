{
  description = "AI-powered git diff analyzer using OpenAI GPT";
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };
  outputs = { self, nixpkgs }: let
    system = "x86_64-linux";
    pkgs = import nixpkgs { inherit system; };
  in {
    packages.${system}.default = pkgs.buildGoModule rec {
      pname = "diffgpt";
      version = "latest";
      src = ./.;

      vendorHash = "sha256-YMPiHe2DEA/1E8wtB1GJf/pvJ0vl3TjfquZdvDA9NDU=";

      buildPhase = ''
        runHook preBuild
        make build
        runHook postBuild
      '';

      installPhase = ''
        runHook preInstall
        mkdir -p $out/bin
        cp build/diffgpt $out/bin/
        runHook postInstall
      '';
    };
    devShells.${system}.default = pkgs.mkShell {
      buildInputs = with pkgs; [
        self.packages.${system}.default
        go
        gopls
      ];
    };
  };
}
