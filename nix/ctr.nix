{
  pkgs ? import <nixpkgs> { },
}:
pkgs.dockerTools.streamLayeredImage {
  name = "thecluster.lan";
  tag = "latest";

  contents = [
    app
    pkgs.cacert
  ];

  config = {
    Cmd = [ "/bin/thecluster-api" ];
    Env = [
      "PORT=8080"
      "STATIC_DIR=/wwwroot"
    ];
    ExposedPorts = {
      "8080/tcp" = { };
    };
  };
}
