{
  outputs = { self, nixpkgs }:
  let pkgs = nixpkgs.legacyPackages.x86_64-linux; in
  rec {

    packages.x86_64-linux.icat = pkgs.buildGoModule {
      pname = "icat";
      version = "1.0";

      src = self;

      vendorHash = "sha256-v47V+DNmAImNZy/TXEuX1b4IVp960QPk0Lb+T4e6brI=";
    };

    defaultPackage.x86_64-linux = packages.x86_64-linux.icat;
  };
}
