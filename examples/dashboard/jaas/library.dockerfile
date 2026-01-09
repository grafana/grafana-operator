FROM scratch

# The fully qualified path is the URL of your repository + the (optional) subfolder of your Jsonnet files
COPY src/something /git.example.com/my/jsonnet/library/src/something
