{
  description = "Note sync with GPG and Git";

  inputs.nixpkgs.url = "nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }:
    let
      lastModifiedDate = self.lastModifiedDate or self.lastModified or "19700101";
      version = builtins.substring 0 8 lastModifiedDate;
      supportedSystems = [ "x86_64-linux" "x86_64-darwin" "aarch64-linux" "aarch64-darwin" ];
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
      nixpkgsFor = forAllSystems (system: import nixpkgs { inherit system; });
    in
    {
      packages = forAllSystems (system:
        let
          pkgs = nixpkgsFor.${system};
        in
        {
          note-sync = pkgs.buildGoModule {
            pname = "note-sync";
            inherit version;
            src = ./.;
            vendorSha256 = "sha256-PK1OOqf2VmgT8EcfPGKgL987KNvP6e+tERyqAA515d4=";
          };
        });

      defaultPackage = forAllSystems (system: self.packages.${system}.note-sync);
    };
}
