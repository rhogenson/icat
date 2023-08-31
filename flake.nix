{
  outputs = { self, nixpkgs }:
  let pkgs = nixpkgs.legacyPackages.x86_64-linux; in
  rec {

    packages.x86_64-linux.icat = pkgs.buildGoModule {
      pname = "icat";
      version = "1.0";

      src = self;

      vendorHash = "sha256-WiEK8nq13mdTFnyxDiMRWM1tb60mlZY0fVmtlH5ZWgw=";
    };

    defaultPackage.x86_64-linux = packages.x86_64-linux.icat;
  };
}
