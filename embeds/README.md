# Embeds

This directory is used for embedding FS structs into go code.

# Grafonnet

The Grafonnet Jsonnet is used to allow users to define their dashboards using the jsonnet framework.

To update the library, run:

```shell
git submodule update
```

**Note**: the root level .gitignore will ignore everything apart from grafonnet/ and grafonnet-7.0/ directories.
be careful if you need to add other files than these two, update the project root .gitignore accordingly
