{
  description = "My website, built with Django.";

  outputs = { self, nixpkgs }:
  let
    system = "x86_64-linux";
    pkgs = import nixpkgs { inherit system; };
  in
  {
    devShell.${system} = with pkgs; stdenv.mkDerivation {
      name = "h4n-io-shell";
      buildInputs = [
        python39Full
        nodePackages.serverless
      ];
    };
  };
}
