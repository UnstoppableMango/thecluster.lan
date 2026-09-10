{
  pkgs ? import <nixpkgs> { },
  api,
  web,
}:
pkgs.runCommand "thecluster-app" { } ''
  mkdir -p $out/bin $out/wwwroot
  cp -r ${api}/bin/. $out/bin/
  cp -r ${web}/wwwroot/. $out/wwwroot/
''
