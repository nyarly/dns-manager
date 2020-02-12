{ pkgs ? import <nixpkgs> {} }:
let
  inherit (pkgs) go_1_13;
in
  pkgs.mkShell {
    buildInputs = [go_1_13];
  }
