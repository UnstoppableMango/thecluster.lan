{
  pkgs ? import <nixpkgs> { },
  app,
}:
pkgs.dockerTools.streamLayeredImage {
  name = "thecluster.lan/dashboard";
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
