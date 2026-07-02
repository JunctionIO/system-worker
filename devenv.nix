{ pkgs, ... }:
{
  packages = [
    pkgs.go
  ];

  enterShell = ''
    set +x
    set -a; [ -f .env ] && source .env; set +a
  '';

  processes.worker.exec = "go run .";
}
