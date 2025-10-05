# api

The OpenAPI spec at `openapi.yaml` uses OpenAPI 3.1.
Because `oapi-codegen` doesn’t fully support 3.1 yet (see the [open issue][4]), we apply an Overlay
during generation to temporarily downshift to OpenAPI 3.0 for codegen while keeping 3.1 as the source of truth.

To generate the Go server, run:
```sh
go generate
```
The generated files are written to `internal/api`.

To review changes:
```sh
git diff internal/api
```

Relevant documentation:
- [OpenAPI Specification v3.1.2][1]
- [OpenAPI Specification v3.0.3][2]
- [Overlay Specification v1.0.0][3]

[1]: https://spec.openapis.org/oas/v3.1.2.html
[2]: https://spec.openapis.org/oas/v3.0.3.html
[3]: https://spec.openapis.org/overlay/v1.0.0.html
[4]: https://github.com/oapi-codegen/oapi-codegen/issues/373
