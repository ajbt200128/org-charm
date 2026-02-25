{ pkgs, lib, config, inputs, ... }:

{
  # https://devenv.sh/packages/
  packages = [ pkgs.git ];

  # https://devenv.sh/languages/
  languages.go.enable = true;

  # https://devenv.sh/basics/
  enterShell = ''
    echo "org-charm development environment"
    go version
  '';

  # https://devenv.sh/tests/
  enterTest = ''
    echo "Running tests"
    go test ./...
  '';
}
