# Walkthrough schema generator

Writes `../walkthrough.schema.json` from `pidl.Walkthrough`, the Go types of
the walkthrough format. It is a separate module so that the pidl module
keeps no dependencies.

```bash
go run .        # regenerate the schema after changing the walkthrough types
go test ./...   # the schema is current, and every built-in example validates
```
