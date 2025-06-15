let
  nixpkgs = fetchTarball "https://github.com/NixOS/nixpkgs/tarball/nixos-25.05";
  pkgs = import nixpkgs { config = {}; overlays = []; };
in

pkgs.mkShellNoCC {
  packages = with pkgs; [
    cowsay
    lolcat

    nodejs
    pnpm_8

    go
    python314
    
    jq
    git
    
    ruby
    node-gyp
    llvmPackages_20.libcxxClang
  ];
}
