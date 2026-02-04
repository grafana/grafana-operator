// Import grafonnet
local grafonnet = import "github.com/grafana/grafonnet/gen/grafonnet-latest/main.libsonnet";

// Import your own stuff. The library has to exists within the OCI image, no network requests are made.
local something = import "git.example.com/my/jsonnet/library/src/something/main.libsonnet";
