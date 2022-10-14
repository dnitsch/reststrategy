# Examples

Both the seeder (CLI) and controller modules use [pocketbase.io]() for integration tests, ensure you have it running locally, you can follow [these instructions](../README.md#mock-app).

Below is a sample to get you started with the CLI, once you have built the `seeder` locally you should be able to run the following from the workspace/repo root.

for linux/windows exchange the `darwin` for `linux|windows`

```bash
make build_seeder
./seeder/dist/seeder-darwin run -p ./test/pocketbase-cli-get-started.yaml -v
```

## CLI download/install

Follow instructions [here](./installation.md) to install the CLI from a pre-built package.
