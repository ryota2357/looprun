{
  lib,
  buildGoModule,
}:

buildGoModule {
  pname = "looprun";
  version = "0.0.0";

  src = lib.cleanSource ./.;

  vendorHash = "sha256-m5mBubfbXXqXKsygF5j7cHEY+bXhAMcXUts5KBKoLzM=";

  meta = with lib; {
    description = "Repeat a given command";
    homepage = "https://github.com/ryota2357/looprun";
    license = licenses.mit;
    mainProgram = "looprun";
  };
}
