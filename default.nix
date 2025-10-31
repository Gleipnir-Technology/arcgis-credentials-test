{ pkgs ? import <nixpkgs> { } }:
pkgs.buildGoModule rec {
        meta = {
                description = "ArcGIS Credentials Test";
                homepage = "https://github.com/Gleipnir-Technology/arcgis-credentials-test";
        };
        pname = "arcgis-credentials-test";
        src = ./.;
        subPackages = [];
        version = "0.0.1";
        # Needs to be updated after every modification of go.mod/go.sum
        vendorHash = "sha256-3NQcLcPNfba20LXenLu/RcZistBzEGsRb75IogvSR68=";
}
