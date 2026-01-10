{
  description = "Entorno de desarrollo puro para Go";

  inputs = {
    # Usamos unstable para tener las últimas versiones de Go y gopls
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs }:
    let
      system = "x86_64-linux";
      pkgs = import nixpkgs { inherit system; };
    in
    {
      devShells.${system}.default = pkgs.mkShell {
        name = "go-dev-shell";

        # Paquetes necesarios para el desarrollo
        buildInputs = with pkgs; [
          # Lenguaje
          go
          
          # Herramientas de soporte (LSP, Debugger, Linters)
          gopls
          delve
          golangci-lint
          gotools # Herramientas como goimports
        ];

        # Variables de entorno y scripts de inicio
        shellHook = ''
          export GOPATH="$HOME/go"
          export PATH="$GOPATH/bin:$PATH"
          
          echo "🚀 Go Development Environment"
          echo "Directorio de trabajo: $(pwd)"
          go version
        '';
      };
    };
}
